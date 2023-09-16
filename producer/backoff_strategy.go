// Package producer provides abstractions and implementations
// related to producing messages, including retry strategies.
package producer

import (
	"time"

	"github.com/cenkalti/backoff/v4"
)

// backoffStrategy defines an interface for determining
// the duration to wait before retrying an operation.
// This interface abstracts various backoff strategies
// allowing for easier mocking and testing.
type backoffStrategy interface {
	// NextBackOff returns the duration to wait before retrying.
	NextBackOff() time.Duration
	// Reset clears any internal state of the backoff strategy.
	Reset()
}

// backoffStrategyWrapper provides a wrapper around the actual
// backoff.BackOff from the "github.com/cenkalti/backoff/v4" library.
// It implements the backoffStrategy interface.
type backoffStrategyWrapper struct {
	bo backoff.BackOff
}

// NewBackoffStrategyWrapper creates and initializes a new
// exponential backoff strategy with a specified maximum elapsed time.
func NewBackoffStrategyWrapper(maxElapsedTime time.Duration) backoffStrategy {
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = maxElapsedTime
	return &backoffStrategyWrapper{
		bo: bo,
	}
}

// NextBackOff returns the next duration to wait based on the
// exponential backoff algorithm.
func (b *backoffStrategyWrapper) NextBackOff() time.Duration {
	return b.bo.NextBackOff()
}

// Reset clears any internal state of the backoff, resetting it to its initial state.
func (b *backoffStrategyWrapper) Reset() {
	b.bo.Reset()
}
