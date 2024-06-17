package retry_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/arsham/retry/v3"
)

func ExampleRetry_DoContext() {
	r := &retry.Retry{
		Attempts: 4,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := r.DoContext(ctx, func() error {
		// we cancel early to prove a point.
		cancel()
		return nil
	}, func() error {
		panic("should not have been called")
	})
	fmt.Println("Error:", err)

	// Output:
	// Error: context canceled
}

func ExampleRetry_Do() {
	r := &retry.Retry{
		Attempts: 4,
	}
	err := r.Do(func() error {
		return nil
	})
	fmt.Println("Error:", err)

	// Output:
	// Error: <nil>
}

func ExampleRetry_Do_zero() {
	r := &retry.Retry{}
	err := r.Do(func() error {
		fmt.Println("this should not happen")
		return nil
	})
	fmt.Println("Error:", err)

	// Output:
	// Error: <nil>
}

func ExampleRetry_Do_error() {
	r := &retry.Retry{
		Attempts: 4,
		Delay:    time.Nanosecond,
	}
	err := r.Do(func() error {
		return errors.New("some error")
	})
	fmt.Println("Error:", err)

	// Output:
	// Error: some error
}

func ExampleRetry_Do_standardMethod() {
	r := &retry.Retry{
		Attempts: 4,
		Delay:    time.Nanosecond,
	}
	i := 0
	err := r.Do(func() error {
		i++
		fmt.Printf("Running iteration %d.\n", i)
		if i < 3 {
			return errors.New("ignored error")
		}
		return nil
	})
	fmt.Println("Error:", err)

	// Output:
	// Running iteration 1.
	// Running iteration 2.
	// Running iteration 3.
	// Error: <nil>
}

func ExampleStopError() {
	r := &retry.Retry{
		Attempts: 10,
	}
	i := 0
	err := r.Do(func() error {
		i++
		fmt.Printf("Running iteration %d.\n", i)
		if i > 2 {
			return &retry.StopError{}
		}
		return errors.New("ignored error")
	})
	fmt.Println("Error:", err)

	// Output:
	// Running iteration 1.
	// Running iteration 2.
	// Running iteration 3.
	// Error: <nil>
}

func ExampleStopError_stopErr() {
	r := &retry.Retry{
		Attempts: 10,
	}
	i := 0
	stopErr := &retry.StopError{
		Err: errors.New("this is the returned error"),
	}
	err := r.Do(func() error {
		i++
		if i > 2 {
			return stopErr
		}
		return errors.New("ignored error")
	})
	fmt.Println("Error:", err)
	fmt.Println("Stopped with:", stopErr)

	// Output:
	// Error: this is the returned error
	// Stopped with: this is the returned error
}

func ExampleIncrementalDelay() {
	// This setup will delay 20ms + 40ms + 80ms + 160ms, and a jitters at 5
	// attempts, until on the 6th attempt that it would succeed.
	r := &retry.Retry{
		Attempts: 6,
		Delay:    20 * time.Millisecond,
		Method:   retry.IncrementalDelay,
	}
	i := 0
	err := r.Do(func() error {
		i++
		if i < r.Attempts {
			return errors.New("ignored error")
		}
		return nil
	})
	fmt.Println("Error:", err)

	// Output:
	// Error: <nil>
}

func ExampleIncrementalDelayMax() {
	// This setup will delay 20ms + 40ms + 50ms + 50ms, and a jitters at 5
	// attempts, until on the 6th attempt that it would succeed.
	r := &retry.Retry{
		Attempts: 6,
		Delay:    20 * time.Millisecond,
		Method:   retry.IncrementalDelayMax(50 * time.Millisecond),
	}
	i := 0
	err := r.Do(func() error {
		i++
		if i < r.Attempts {
			return errors.New("ignored error")
		}
		return nil
	})
	fmt.Println("Error:", err)

	// Output:
	// Error: <nil>
}

func ExampleRetry_Do_multipleFuncs() {
	r := &retry.Retry{
		Attempts: 4,
		Delay:    time.Nanosecond,
	}
	err := r.Do(func() error {
		fmt.Println("Running func 1.")
		return nil
	}, func() error {
		fmt.Println("Running func 2.")
		return nil
	}, func() error {
		fmt.Println("Running func 3.")
		return nil
	})
	fmt.Println("Error:", err)

	// Output:
	// Running func 1.
	// Running func 2.
	// Running func 3.
	// Error: <nil>
}
