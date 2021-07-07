// Package retry invokes a given function until it succeeds. It sleeps in
// between attempts based the DelayMethod. It is useful in situations that an
// action might succeed after a few attempt due to unavailable resources or
// waiting for a condition to happen.
//
// The default DelayMethod sleeps exactly the same amount of time between
// attempts. You can use the IncrementalDelay method to increment the delays
// between attempts. It gives a jitter to the delay to prevent Thundering herd
// problems. If the delay is 0 in either case, it does not sleep between tries.
package retry

import (
	"errors"
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

// StopError causes the Do method stop trying and will return the Err. This
// error then is returned by the Do method.
type StopError struct {
	Err error
}

func (s StopError) Error() string { return s.Err.Error() }

// Retry attempts to call a given function until it succeeds, or returns a
// StopError value for a certain amount of times. It will delay between calls
// for any errors based on the provided Method. Retry is concurrent safe and
// the zero value does not do anything.
type Retry struct {
	Method   DelayMethod
	Delay    time.Duration
	Attempts int
}

// Do calls fn until it returns nil or a StopError. It delays and retries if
// the fn returns any errors or panics. The value fo the returned error, or the
// Err of a StopError, or an error with the panic message will be returned at
// the last cycle.
func (r Retry) Do(fn func() error) error {
	method := r.Method
	if method == nil {
		method = StandardDelay
	}
	var err error
	for i := 0; i < r.Attempts; i++ {
		func() {
			defer func() {
				if e := recover(); e != nil {
					err = fmt.Errorf("function caused a panic: %v", e)
				}
			}()
			err = fn()
		}()
		if err == nil {
			return nil
		}
		var (
			v1 StopError
			v2 *StopError
		)
		if errors.As(err, &v1) {
			return v1.Err
		}
		if errors.As(err, &v2) {
			return v2.Err
		}
		time.Sleep(method(i+1, r.Delay))
	}
	return err
}

// StandardDelay always delays the same amount of time.
func StandardDelay(_ int, delay time.Duration) time.Duration { return delay }

// IncrementalDelay increases the delay between attempts up to a second. It
// adds a jitter to prevent Thundering herd. If the delay is 0, it always
// returns 0.
func IncrementalDelay(attempt int, delay time.Duration) time.Duration {
	if delay == 0 {
		return 0
	}
	if delay > time.Second {
		delay = time.Second
	}
	d := int64(delay)
	// nolint:gosec // the rand package is used for fast ransom number generation.
	jitter := rand.Int63n(d) / 2
	return (delay * time.Duration(attempt)) + time.Duration(jitter)
}
