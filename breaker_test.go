package breaker

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	mock_timeout "github.com/haunt98/breaker/pkg/timeout/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCircuitBreakerDo(t *testing.T) {
	ctrl := gomock.NewController(t)

	failureThreshold := 3
	failureTimeout := mock_timeout.NewMockTimeout(ctrl)
	successThreshold := 5

	cb := NewCircuitBreaker(failureThreshold, failureTimeout, successThreshold)

	// Closed

	assert.Equal(t, ClosedStatus, cb.GetStatus())

	wantResult := 69
	gotResult, gotErr := cb.Do(successFn(wantResult))
	assert.NoError(t, gotErr)
	assert.Equal(t, wantResult, gotResult)

	// Open

	failureTimeout.EXPECT().Start()
	wantErr := errors.New("error")
	for i := 0; i < failureThreshold; i++ {
		gotResult, gotErr = cb.Do(failureFn(wantErr))
		assert.Equal(t, wantErr, gotErr)
		assert.Nil(t, gotResult)
	}
	assert.Equal(t, OpenStatus, cb.GetStatus())

	failureTimeout.EXPECT().IsStop().Return(false)
	gotResult, gotErr = cb.Do(successFn(420))
	assert.Equal(t, gotErr, CircuitBreakerOpenError)
	assert.Nil(t, gotResult)

	failureTimeout.EXPECT().IsStop().Return(false)
	gotResult, gotErr = cb.Do(failureFn(errors.New("some error")))
	assert.Equal(t, gotErr, CircuitBreakerOpenError)
	assert.Nil(t, gotResult)

	// Half open

	wantResult = 1337

	failureTimeout.EXPECT().IsStop().Return(true)
	gotResult, gotErr = cb.Do(successFn(wantResult))
	assert.NoError(t, gotErr)
	assert.Equal(t, wantResult, gotResult)
	assert.Equal(t, HalfOpenStatus, cb.GetStatus())
	gotResult, gotErr = cb.Do(successFn(wantResult))
	assert.NoError(t, gotErr)
	assert.Equal(t, wantResult, gotResult)
	assert.Equal(t, HalfOpenStatus, cb.GetStatus())

	gotResult, gotErr = cb.Do(successFn(wantResult))
	assert.NoError(t, gotErr)
	assert.Equal(t, wantResult, gotResult)

	// Closed

	wantResult = 1111

	// Already 2 times before
	for i := 0; i < successThreshold-2; i++ {
		gotResult, gotErr = cb.Do(successFn(wantResult))
		assert.NoError(t, gotErr)
		assert.Equal(t, wantResult, gotResult)
	}
	assert.Equal(t, ClosedStatus, cb.GetStatus())
}

func successFn(v interface{}) func() (interface{}, error) {
	return func() (interface{}, error) {
		return v, nil
	}
}

func failureFn(err error) func() (interface{}, error) {
	return func() (interface{}, error) {
		return nil, err
	}
}
