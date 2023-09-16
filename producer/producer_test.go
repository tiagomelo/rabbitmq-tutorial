package producer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
)

var logger = log.New(os.Stdout, "PRODUCER UNIT TEST : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

func TestNew_withoutOpts(t *testing.T) {
	mac := &mockAmqpConnection{}
	testCases := []struct {
		name          string
		mockDial      func(url string) (amqpConnection, error)
		mockClosure   func(conn *mockAmqpConnection)
		expectedError error
	}{
		{
			name: "happy path",
			mockDial: func(url string) (amqpConnection, error) {
				return mac, nil
			},
			mockClosure: func(conn *mockAmqpConnection) {
				channel := new(mockAmqpChannel)
				conn.channel = channel
			},
		},
		{
			name: "error when dialing",
			mockDial: func(url string) (amqpConnection, error) {
				return nil, errors.New("dialing error")
			},
			mockClosure:   func(conn *mockAmqpConnection) {},
			expectedError: errors.New("failed to dial to url: dialing error"),
		},
		{
			name: "error when getting channel from conn",
			mockDial: func(url string) (amqpConnection, error) {
				return mac, nil
			},
			mockClosure: func(conn *mockAmqpConnection) {
				conn.errChannel = errors.New("getting channel error")
			},
			expectedError: errors.New("failed to get channel from conn: getting channel error"),
		},
		{
			name: "error when declaring queue",
			mockDial: func(url string) (amqpConnection, error) {
				return mac, nil
			},
			mockClosure: func(conn *mockAmqpConnection) {
				channel := new(mockAmqpChannel)
				channel.errQueueDeclare = errors.New("declaring queue error")
				conn.channel = channel
			},
			expectedError: errors.New(`failed to declare queue with name "someQueue": declaring queue error`),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				mac.errChannel = nil
			}()

			dial = tc.mockDial
			tc.mockClosure(mac)
			p, err := New("url", "someQueue", logger)
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf("expected no error, got %v", err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf("expected error %v, got nil", tc.expectedError)
				}
				require.NotNil(t, p)
			}
		})
	}
}

func TestNew_withOpts(t *testing.T) {
	mac := &mockAmqpConnection{}
	testCases := []struct {
		name          string
		mockDial      func(url string) (amqpConnection, error)
		mockClosure   func(conn *mockAmqpConnection)
		expectedError error
	}{
		{
			name: "happy path",
			mockDial: func(url string) (amqpConnection, error) {
				return mac, nil
			},
			mockClosure: func(conn *mockAmqpConnection) {
				channel := new(mockAmqpChannel)
				conn.channel = channel
			},
		},
		{
			name: "error when dialing",
			mockDial: func(url string) (amqpConnection, error) {
				return nil, errors.New("dialing error")
			},
			mockClosure:   func(conn *mockAmqpConnection) {},
			expectedError: errors.New("failed to dial to url: dialing error"),
		},
		{
			name: "error when getting channel from conn",
			mockDial: func(url string) (amqpConnection, error) {
				return mac, nil
			},
			mockClosure: func(conn *mockAmqpConnection) {
				conn.errChannel = errors.New("getting channel error")
			},
			expectedError: errors.New("failed to get channel from conn: getting channel error"),
		},
		{
			name: "error when declaring queue",
			mockDial: func(url string) (amqpConnection, error) {
				return mac, nil
			},
			mockClosure: func(conn *mockAmqpConnection) {
				channel := new(mockAmqpChannel)
				channel.errQueueDeclare = errors.New("declaring queue error")
				conn.channel = channel
			},
			expectedError: errors.New(`failed to declare queue with name "someQueue": declaring queue error`),
		},
		{
			name: "error when declaring exchange",
			mockDial: func(url string) (amqpConnection, error) {
				return mac, nil
			},
			mockClosure: func(conn *mockAmqpConnection) {
				channel := new(mockAmqpChannel)
				channel.errExchangeDeclare = errors.New("declaring exchange error")
				conn.channel = channel
			},
			expectedError: errors.New(`failed to declare exchange with name "exchangeName": declaring exchange error`),
		},
		{
			name: "error when binding queue to exchange",
			mockDial: func(url string) (amqpConnection, error) {
				return mac, nil
			},
			mockClosure: func(conn *mockAmqpConnection) {
				channel := new(mockAmqpChannel)
				channel.errQueueBind = errors.New("bindind queue error")
				conn.channel = channel
			},
			expectedError: errors.New(`failed to bind queue "someQueue" to exchange "exchangeName": bindind queue error`),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				mac.errChannel = nil
			}()

			dial = tc.mockDial
			tc.mockClosure(mac)
			logger := log.New(os.Stdout, "PRODUCER UNIT TEST : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
			p, err := New("url", "someQueue", logger, WithExchange("exchangeName", "exchangeKind", "bindingKey"))
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf("expected no error, got %v", err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf("expected error %v, got nil", tc.expectedError)
				}
				require.NotNil(t, p)
			}
		})
	}
}

func TestPublish(t *testing.T) {
	testCases := []struct {
		name          string
		conn          *mockAmqpConnection
		mockChannel   func() *mockAmqpChannel
		expectedError error
	}{
		{
			name: "happy path",
			conn: new(mockAmqpConnection),
			mockChannel: func() *mockAmqpChannel {
				return new(mockAmqpChannel)
			},
		},
		{
			name: "error",
			conn: new(mockAmqpConnection),
			mockChannel: func() *mockAmqpChannel {
				mockedChannel := &mockAmqpChannel{
					errPublishWithContext: errors.New("publish error"),
				}
				return mockedChannel
			},
			expectedError: errors.New(`failed to publish to queue "someQueue": publish error`),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Printf("tc.conn: %v\n", tc.conn == nil)
			mockedChannel := tc.mockChannel()
			p := &producer{
				url:       "url",
				queueName: "someQueue",
				conn:      tc.conn,
				channel:   mockedChannel,
				logger:    logger,
			}
			err := p.Publish(context.TODO(), []byte("some message"))
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf("expected no error, got %v", err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf("expected error %v, got nil", tc.expectedError)
				}
			}
		})
	}
}

func TestClose(t *testing.T) {
	testCases := []struct {
		name          string
		mockConn      func() *mockAmqpConnection
		mockChannel   func() *mockAmqpChannel
		expectedError error
	}{
		{
			name: "happy path",
			mockChannel: func() *mockAmqpChannel {
				return new(mockAmqpChannel)
			},
			mockConn: func() *mockAmqpConnection {
				return new(mockAmqpConnection)
			},
		},
		{
			name: "error when closing channel",
			mockChannel: func() *mockAmqpChannel {
				mockChannel := &mockAmqpChannel{
					errClose: errors.New("channel close error"),
				}
				return mockChannel
			},
			mockConn: func() *mockAmqpConnection {
				return new(mockAmqpConnection)
			},
			expectedError: errors.New("error closing channel: channel close error"),
		},
		{
			name: "error when closing conn",
			mockChannel: func() *mockAmqpChannel {
				return new(mockAmqpChannel)
			},
			mockConn: func() *mockAmqpConnection {
				mockedConn := &mockAmqpConnection{
					errClose: errors.New("connection close error"),
				}
				return mockedConn
			},
			expectedError: errors.New("error closing connection: connection close error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			channel := tc.mockChannel()
			conn := tc.mockConn()
			p := &producer{
				conn:    conn,
				channel: channel,
				logger:  logger,
			}
			err := p.Close()
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf("expected no error, got %v", err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf("expected error %v, got nil", tc.expectedError)
				}
			}
		})
	}
}

func Test_exchangeName(t *testing.T) {
	testCases := []struct {
		name           string
		exchangeConfig *exchangeConfig
		expectedOutput string
	}{
		{
			name: "returns exchange name",
			exchangeConfig: &exchangeConfig{
				name: "exchangeName",
			},
			expectedOutput: "exchangeName",
		},
		{
			name:           "returns empty if exchange name is not present",
			exchangeConfig: &exchangeConfig{},
			expectedOutput: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &producer{
				exchangeConfig: tc.exchangeConfig,
			}
			en := p.exchangeName()
			require.Equal(t, tc.expectedOutput, en)
		})
	}
}

func Test_bindingKey(t *testing.T) {
	testCases := []struct {
		name           string
		queueName      string
		exchangeConfig *exchangeConfig
		expectedOutput string
	}{
		{
			name:      "returns binding key",
			queueName: "queueName",
			exchangeConfig: &exchangeConfig{
				bindingKey: "bindingKey",
			},
			expectedOutput: "bindingKey",
		},
		{
			name:           "returns queue name if no exchange config present",
			queueName:      "queueName",
			exchangeConfig: nil,
			expectedOutput: "queueName",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ec *exchangeConfig
			if tc.exchangeConfig != nil {
				ec = tc.exchangeConfig
			}
			p := &producer{
				queueName:      tc.queueName,
				exchangeConfig: ec,
			}
			bk := p.bindingKey()
			require.Equal(t, tc.expectedOutput, bk)
		})
	}
}

func Test_monitorClose(t *testing.T) {
	mockBOStrategy := new(mockBackoffStrategy)
	mockBOStrategy.duration = backoff.Stop

	mockConn := new(mockAmqpConnection)
	mockConn.errChannel = errors.New("getting channel")

	p := &producer{
		closeCh:         make(chan *amqp.Error),
		conn:            mockConn,
		logger:          logger,
		backoffStrategy: mockBOStrategy,
	}

	go func() {
		p.closeCh <- &amqp.Error{Reason: "Test error", Code: 500}
	}()

	p.monitorClose()
	require.Equal(t, mockBOStrategy.counter, 1)
}

type mockAmqpChannel struct {
	errQueueDeclare       error
	errExchangeDeclare    error
	errQueueBind          error
	errPublishWithContext error
	errClose              error
}

func (m *mockAmqpChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	return amqp.Queue{}, m.errQueueDeclare
}

func (m *mockAmqpChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	return m.errExchangeDeclare
}

func (m *mockAmqpChannel) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	return m.errQueueBind
}

func (m *mockAmqpChannel) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return m.errPublishWithContext
}

func (m *mockAmqpChannel) NotifyClose(chan *amqp.Error) chan *amqp.Error {
	c := make(chan *amqp.Error)
	return c
}

func (m *mockAmqpChannel) Close() error {
	return m.errClose
}

type mockAmqpConnection struct {
	channel    amqpChannel
	errChannel error
	errClose   error
}

func (m *mockAmqpConnection) Channel() (amqpChannel, error) {
	return m.channel, m.errChannel
}

func (m *mockAmqpConnection) Close() error {
	return m.errClose
}

type mockBackoffStrategy struct {
	duration time.Duration
	counter  int
}

func (m *mockBackoffStrategy) NextBackOff() time.Duration {
	m.counter++
	return m.duration
}

func (m *mockBackoffStrategy) Reset() {}
