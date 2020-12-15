package breaker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCircuitBreakerDo(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Minute, 3)

	wantResult := 69

	gotResult, err := cb.Do(func() (interface{}, error) {
		return wantResult, nil
	})

	assert.NoError(t, err)
	assert.Equal(t, gotResult, wantResult)
}
