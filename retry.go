package retry

import (
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// DelayMethod determines how the delay behaves. The current attempt is passed
// on each iteration, with the delay value of the Retry object.
type DelayMethod func(attempt int, delay time.Duration) time.Duration

// StopError causes the Do method stop trying and will return the Err.
type StopError struct {
	Err error
}

func (s StopError) Error() string { return s.Err.Error() }

// Retry implements a Do method that would call a given function Attempts times
// until it returns nil. It will delay between calls for any errors based on the
// provided Method. Retry is concurrent safe. The zero value does not do
// anything.
type Retry struct {
	Attempts int
	Delay    time.Duration
	Method   DelayMethod
}

// Do calls fn for Attempts times until it returns nil or a StopError. If
// retries is 0 fn would not be called. It delays and retries if the fn returns
// any errors or panics.
func (r Retry) Do(fn func() error) error {
	method := r.Method
	if method == nil {
		method = StandardDelay
	}
	var err error
	for i := 0; i < r.Attempts; i++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("function caused a panic: %v", r)
				}
			}()
			err = fn()
		}()
		if err == nil {
			return nil
		}
		if v, ok := err.(StopError); ok {
			return v.Err
		}
		time.Sleep(method(i+1, r.Delay))
	}
	return err
}

// StandardDelay always delays the same amount of time.
func StandardDelay(_ int, delay time.Duration) time.Duration { return delay }

// IncrementalDelay increases the delay between attempts. It adds a jitter to
// prevent Thundering herd.
func IncrementalDelay(attempt int, delay time.Duration) time.Duration {
	d := int64(delay)
	if d > 1000000000 { // a second
		d = 1000000000
	}
	jitter := rand.Int63n(d) / 2
	return (delay * time.Duration(attempt)) + time.Duration(jitter)
}
