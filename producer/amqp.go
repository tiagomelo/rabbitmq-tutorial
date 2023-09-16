// Package producer provides abstractions and implementations
// to interact with RabbitMQ for producing messages.
package producer

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

// amqpChannel defines an interface for operations
// that can be performed on a RabbitMQ channel.
// This interface abstracts the actual RabbitMQ channel operations
// allowing for easier mocking and testing.
type amqpChannel interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error
	QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error
	PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	NotifyClose(chan *amqp.Error) chan *amqp.Error
	Close() error
}

// amqpChannelWrapper provides a wrapper around the actual amqp.Channel
// implementing the amqpChannel interface.
type amqpChannelWrapper struct {
	*amqp.Channel
}

// amqpConnection defines an interface for operations
// that can be performed on a RabbitMQ connection.
// Like the amqpChannel interface, this allows for abstraction
// of the real RabbitMQ connection for testing purposes.
type amqpConnection interface {
	Channel() (amqpChannel, error)
	Close() error
}

// amqpConnectionWrapper provides a wrapper around the actual amqp.Connection
// implementing the amqpConnection interface.
type amqpConnectionWrapper struct {
	conn *amqp.Connection
}

// Channel retrieves a channel from the wrapped RabbitMQ connection
// and returns it wrapped in the amqpChannelWrapper.
func (cw *amqpConnectionWrapper) Channel() (amqpChannel, error) {
	c, err := cw.conn.Channel()
	return &amqpChannelWrapper{c}, err
}

// Close closes the underlying RabbitMQ connection.
func (cw *amqpConnectionWrapper) Close() error {
	return cw.conn.Close()
}
