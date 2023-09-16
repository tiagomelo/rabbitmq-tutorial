// Package producer provides utilities for producing messages.
package producer

// Option represents a function that configures a producer.
// This type is used to implement the functional options pattern,
// which provides a flexible way to configure a producer instance.
type Option func(p *producer)

// WithExchange returns an Option that configures a producer
// to use a specific exchange. The exchange is defined by its name,
// type (kind), and binding key.
//
// name: The name of the exchange.
// kind: The type of the exchange (e.g., "direct", "fanout", "topic", etc.).
// bindingKey: The binding key used for routing messages.
func WithExchange(name, kind, bindingKey string) Option {
	return func(p *producer) {
		p.exchangeConfig = &exchangeConfig{
			name:       name,
			kind:       kind,
			bindingKey: bindingKey,
		}
	}
}
