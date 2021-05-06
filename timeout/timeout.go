package timeout

import "time"

//go:generate mockgen -source=timeout.go -destination=gomock/timeout.go

type Timeout interface {
	Start()
	IsStop() bool
}

type timeout struct {
	start    time.Time
	duration time.Duration
}

func NewTimeout(duration time.Duration) Timeout {
	return &timeout{
		duration: duration,
	}
}

func (t *timeout) Start() {
	t.start = time.Now()
}

func (t *timeout) IsStop() bool {
	return time.Since(t.start) >= t.duration
}
