package retry_test

import (
	"fmt"
	"time"

	"github.com/arsham/retry"
	"github.com/pkg/errors"
)

func ExampleRetry_Do() {
	l := &retry.Retry{
		Attempts: 4,
	}
	err := l.Do(func() error {
		return nil
	})
	fmt.Println("Error:", err)

	// Output:
	// Error: <nil>
}

func ExampleRetry_Do_zero() {
	l := &retry.Retry{}
	err := l.Do(func() error {
		fmt.Println("this should not happen")
		return nil
	})
	fmt.Println("Error:", err)

	// Output:
	// Error: <nil>
}

func ExampleRetry_Do_error() {
	l := &retry.Retry{
		Attempts: 4,
		Delay:    time.Nanosecond,
	}
	err := l.Do(func() error {
		return errors.New("some error")
	})
	fmt.Println("Error:", err)

	// Output:
	// Error: some error
}

func ExampleRetry_Do_standard() {
	l := &retry.Retry{
		Attempts: 4,
		Delay:    time.Nanosecond,
	}
	i := 0
	err := l.Do(func() error {
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

func ExampleRetry_Do_incremental() {
	l := &retry.Retry{
		Attempts: 4,
		Delay:    time.Nanosecond,
		Method:   retry.IncrementalDelay,
	}
	i := 0
	err := l.Do(func() error {
		i++
		if i < 3 {
			return errors.New("ignored error")
		}
		return nil
	})
	fmt.Println("Error:", err)

	// Output:
	// Error: <nil>
}

func ExampleRetry_Do_stop() {
	l := &retry.Retry{
		Attempts: 10,
	}
	i := 0
	err := l.Do(func() error {
		i++
		fmt.Printf("Running iteration %d.\n", i)
		if i > 2 {
			return retry.StopError{}
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

func ExampleRetry_Do_stopErr() {
	l := &retry.Retry{
		Attempts: 10,
	}
	i := 0
	stopErr := &retry.StopError{
		Err: errors.New("this is the returned error"),
	}
	err := l.Do(func() error {
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
