// Package retry invokes a given function until it succeeds. It sleeps in
// between attempts based the DelayMethod. It is useful in situations that an
// action might succeed after a few attempt due to unavailable resources or
// waiting for a condition to happen.
//
// The default DelayMethod sleeps exactly the same amount of time between
// attempts. You can use the IncrementalDelay method to increment the delays
// between attempts. It gives a jitter to the delay to prevent Thundering herd
// problems. If the delay is 0 in either case, it does not sleep between tries.
// The IncrementalDelay has a maximum delay of 1 second, but if you need a more
// flexible delay, you can use the IncrementalDelayMax method and give it a max
// delay.
package retry

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"runtime/debug"
	"time"
)

// DelayMethod determines how the delay behaves. The current attempt is passed
// on each iteration, with the delay value of the Retry object.
type DelayMethod func(attempt int, delay time.Duration) time.Duration

// StopError causes the Do method stop trying and will return the Err. This
// error then is returned by the Do method.
type StopError struct {
	Err error
}

func (s StopError) Error() string { return s.Err.Error() }

func (s StopError) Unwrap() error { return s.Err }

// Retry attempts to call a given function until it succeeds, or returns a
// StopError value for a certain amount of times. It will delay between calls
// for any errors based on the provided Method. Retry is concurrent safe and
// the zero value does not do anything.
type Retry struct {
	Method   DelayMethod
	Delay    time.Duration
	MaxDelay time.Duration
	Attempts int
}

type repeatFunc func() error

// Do calls fn until it returns nil or a StopError. It delays and retries if
// the fn returns any errors or panics. The value of the returned error, or the
// Err of a StopError, or an error with the panic message will be returned at
// the last cycle.
func (r Retry) Do(fn1 repeatFunc, fns ...repeatFunc) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return r.DoContext(ctx, fn1, fns...)
}

// DoContext calls fn until it returns nil or a StopError. It delays and
// retries if the fn returns any errors or panics. If the context is cancelled,
// it will stop iterating and returns the error reported by ctx.Err() method.
// The value of the returned error, or the Err of a StopError, or an error with
// the panic message will be returned at the last cycle.
func (r Retry) DoContext(ctx context.Context, fn1 repeatFunc, fns ...repeatFunc) error {
	method := r.Method
	if method == nil {
		method = StandardDelay
	}
	var err error
	for i := range r.Attempts {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		err = r.do(ctx, fn1, fns...)
		if err == nil {
			return nil
		}
		var e *StopError
		if errors.As(err, &e) {
			return e.Err
		}
		time.Sleep(method(i+1, r.Delay))
	}

	return err
}

var errPanic = errors.New("function caused a panic")

func (r Retry) do(ctx context.Context, fn1 repeatFunc, fns ...repeatFunc) error {
	var err error
	for _, fn := range append([]repeatFunc{fn1}, fns...) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		func() {
			defer func() {
				if e := recover(); e != nil {
					switch x := e.(type) {
					case error:
						err = fmt.Errorf("%w: %w\n%s", errPanic, x, debug.Stack())
					default:
						err = fmt.Errorf("%w: %s\n%s", errPanic, e, debug.Stack())
					}
				}
			}()
			err = fn()
		}()
		if err != nil {
			return err
		}
	}

	return nil
}

// StandardDelay always delays the same amount of time.
func StandardDelay(_ int, delay time.Duration) time.Duration { return delay }

// IncrementalDelay increases the delay between attempts up to a second. It
// adds a jitter to prevent Thundering herd. If the delay is 0, it always
// returns 0.
func IncrementalDelay(attempt int, delay time.Duration) time.Duration {
	return IncrementalDelayMax(time.Second)(attempt, delay)
}

// IncrementalDelayMax returns a DelayMethod that increases the delay between
// attempts up to the given maximum duration. It adds a jitter to prevent
// Thundering herd. If the delay is 0, it always returns 0.
func IncrementalDelayMax(maximum time.Duration) func(int, time.Duration) time.Duration {
	return func(attempt int, delay time.Duration) time.Duration {
		if delay == 0 {
			return 0
		}
		if delay > maximum {
			delay = maximum
		}
		d := int64(delay)
		//nolint:gosec // the rand package is used for fast random number generation.
		jitter := rand.Int63n(d) / 2

		return (delay * time.Duration(attempt)) + time.Duration(jitter)
	}
}
