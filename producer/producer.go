package producer

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
)

// maxElapsedTime defines the maximum duration to try to reconnect
// after a connection failure.
const maxElapsedTime = 2 * time.Minute

// dial is a function variable for connecting to the RabbitMQ server.
// This allows for easy mocking in tests.
var dial = func(url string) (amqpConnection, error) {
	conn, err := amqp.Dial(url)
	return &amqpConnectionWrapper{conn}, err
}

// exchangeConfig holds the configuration of the RabbitMQ exchange.
type exchangeConfig struct {
	name       string
	kind       string
	bindingKey string
}

// Producer defines the interface for producing messages.
type Producer interface {
	Publish(ctx context.Context, msg []byte) error
	Close() error
}

// producer is the concrete implementation of the Producer interface.
// It contains necessary fields for publishing messages and maintaining
// a connection to RabbitMQ.
type producer struct {
	url            string
	queueName      string
	logger         *log.Logger
	exchangeConfig *exchangeConfig
	conn           amqpConnection
	channel        amqpChannel
	closeCh        chan *amqp.Error

	backoffStrategy backoffStrategy
}

// connect establishes a connection to the RabbitMQ server.
func (p *producer) connect() error {
	conn, err := dial(p.url)
	if err != nil {
		errorMsg := fmt.Errorf("failed to dial to %s: %v", p.url, err)
		p.logger.Println(errorMsg)
		return errors.New(errorMsg.Error())
	}
	p.conn = conn
	return nil
}

// monitorClose listens for closing events on the channel.
// If the channel closes due to an error, it will attempt to reconnect using a backoff strategy.
func (p *producer) monitorClose() {
	err := <-p.closeCh
	if err != nil {
		p.logger.Printf("channel closed with error: %v", err)
		operation := func() error {
			return p.setupChannel()
		}
		if err := backoff.Retry(operation, p.backoffStrategy); err != nil {
			p.logger.Printf("failed to reconnect after backoff: %v", err)
		}
	}
}

// exchangeName returns the name of the exchange if it exists,
// otherwise an empty string.
func (p *producer) exchangeName() string {
	if p.exchangeConfig != nil {
		return p.exchangeConfig.name
	}
	return ""
}

// bindingKey returns the binding key if an exchange exists,
// otherwise it returns the queue name.
func (p *producer) bindingKey() string {
	if p.exchangeConfig != nil {
		return p.exchangeConfig.bindingKey
	}
	return p.queueName
}

// setupQueue declares a new queue if it doesn't already exist.
func (p *producer) setupQueue() error {
	if _, err := p.channel.QueueDeclare(
		p.queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	); err != nil {
		errorMsg := fmt.Errorf(`failed to declare queue with name "%s": %v`, p.queueName, err)
		p.logger.Println(errorMsg)
		return errors.New(errorMsg.Error())
	}
	return nil
}

// setupExchange declares a new exchange and binds it to the queue
// if it doesn't already exist.
func (p *producer) setupExchange() error {
	if p.exchangeConfig != nil {
		if err := p.channel.ExchangeDeclare(
			p.exchangeConfig.name,
			p.exchangeConfig.kind,
			true,  // durable
			false, // autoDelete
			false, // Internal
			false, // noWait
			nil,   // args,
		); err != nil {
			errorMsg := fmt.Errorf(`failed to declare exchange with name "%s": %v`, p.exchangeConfig.name, err)
			p.logger.Println(errorMsg)
			return errors.New(errorMsg.Error())
		}
		if err := p.channel.QueueBind(
			p.queueName,
			p.exchangeConfig.bindingKey,
			p.exchangeConfig.name,
			false, // noWait
			nil,   // args
		); err != nil {
			errorMsg := fmt.Errorf(`failed to bind queue "%s" to exchange "%s": %v`, p.queueName, p.exchangeConfig.name, err)
			p.logger.Println(errorMsg)
			return errors.New(errorMsg.Error())
		}
	}
	return nil
}

// setupChannel establishes a channel for message publishing.
func (p *producer) setupChannel() error {
	ch, err := p.conn.Channel()
	if err != nil {
		errorMsg := fmt.Errorf("failed to get channel from conn: %v", err)
		p.logger.Println(errorMsg)
		return errors.New(errorMsg.Error())
	}
	p.channel = ch
	p.closeCh = make(chan *amqp.Error)
	p.channel.NotifyClose(p.closeCh)
	go p.monitorClose()
	if err := p.setupQueue(); err != nil {
		return err
	}
	if err := p.setupExchange(); err != nil {
		return err
	}
	return nil
}

// Close shuts down the channel and connection.
func (p *producer) Close() error {
	var err error
	if chErr := p.channel.Close(); chErr != nil {
		errorMsg := fmt.Errorf("error closing channel: %v", chErr)
		err = fmt.Errorf(errorMsg.Error())
		p.logger.Println(errorMsg)
	}
	if connErr := p.conn.Close(); connErr != nil {
		errorMsg := fmt.Errorf("error closing connection: %v", connErr)
		err = fmt.Errorf(errorMsg.Error())
		p.logger.Println(errorMsg)
	}
	return err
}

// Publish sends a message to the RabbitMQ server.
func (p *producer) Publish(ctx context.Context, msg []byte) error {
	if err := p.channel.PublishWithContext(
		ctx,
		p.exchangeName(),
		p.bindingKey(),
		false,
		false,
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         []byte(msg),
			DeliveryMode: amqp.Persistent,
			Priority:     0,
		}); err != nil {
		errorMsg := fmt.Errorf(`failed to publish to queue "%s": %v`, p.queueName, err)
		p.logger.Println(errorMsg)
		return errors.New(errorMsg.Error())
	}
	return nil
}

// New creates a new instance of the producer.
// It accepts the RabbitMQ server URL, queue name, a logger,
// and an optional list of configuration options.
func New(url, queueName string, logger *log.Logger, opts ...Option) (Producer, error) {
	p := &producer{
		url:             url,
		queueName:       queueName,
		logger:          logger,
		backoffStrategy: NewBackoffStrategyWrapper(maxElapsedTime),
	}
	for _, option := range opts {
		option(p)
	}
	err := p.connect()
	if err != nil {
		return nil, err
	}
	err = p.setupChannel()
	if err != nil {
		p.conn.Close()
		return nil, err
	}
	return p, nil
}
