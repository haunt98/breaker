package breaker

import (
	"errors"
	"sync"
	"time"
)

// https://docs.microsoft.com/en-us/azure/architecture/patterns/circuit-breaker

const (
	closedStatus int = iota + 1
	openStatus
	halfOpenStatus
)

var (
	CircuitBreakerOpenError = errors.New("circuit breaker open")
	UnknownStatusError      = errors.New("unknown status")
)

type CircuitBreaker interface {
	Do(fn func() (interface{}, error)) (interface{}, error)
}

type circuitBreaker struct {
	sync.Mutex

	status int

	failureCounter   int
	failureThreshold int
	failureStart     time.Time
	failureTimeout   time.Duration

	successCounter   int
	successThreshold int
}

func NewCircuitBreaker(failureThreshold int, failureTimeout time.Duration, successThreshold int) CircuitBreaker {
	return &circuitBreaker{
		status:           closedStatus,
		failureThreshold: failureThreshold,
		failureTimeout:   failureTimeout,
		successThreshold: successThreshold,
	}
}

func (cb *circuitBreaker) Do(fn func() (interface{}, error)) (interface{}, error) {
	switch cb.status {
	case closedStatus:
		return cb.doClosed(fn)
	case openStatus:
		return cb.doOpen(fn)
	case halfOpenStatus:
		return cb.doHalfOpen(fn)
	default:
		return nil, UnknownStatusError
	}
}

func (cb *circuitBreaker) doClosed(fn func() (interface{}, error)) (interface{}, error) {
	cb.Lock()
	defer cb.Unlock()

	cb.failureCounter = 0

	result, err := fn()
	if err != nil {
		cb.failureCounter++
		if cb.failureCounter >= cb.failureThreshold {
			cb.status = openStatus
			cb.failureStart = time.Now()
		}

		return nil, err
	}

	return result, nil
}

func (cb *circuitBreaker) doOpen(fn func() (interface{}, error)) (interface{}, error) {
	cb.Lock()
	defer cb.Unlock()

	if time.Since(cb.failureStart) < cb.failureTimeout {
		return nil, CircuitBreakerOpenError
	}

	cb.status = halfOpenStatus

	return cb.doHalfOpen(fn)
}

func (cb *circuitBreaker) doHalfOpen(fn func() (interface{}, error)) (interface{}, error) {
	cb.Lock()
	defer cb.Unlock()

	cb.successCounter = 0

	result, err := fn()
	if err != nil {
		return nil, err
	}

	cb.successCounter++
	if cb.successCounter >= cb.successThreshold {
		cb.status = closedStatus
	}

	return result, nil
}
