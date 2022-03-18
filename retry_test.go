package retry_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/arsham/retry"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestLoopDo(t *testing.T) {
	t.Parallel()
	t.Run("Return", testLoopDoReturn)
	t.Run("Zero", testLoopDoZero)
	t.Run("Stopping", testLoopDoStopping)
	t.Run("Panic", testLoopDoPanic)
	t.Run("Sleep", testLoopDoSleep)
	t.Run("MultiFunc", testLoopDoMultiFunc)
	t.Run("ErrorIs", testLoopDoErrorIs)
}

func testLoopDoReturn(t *testing.T) {
	t.Parallel()
	l := &retry.Retry{
		Attempts: 10,
	}
	err := l.Do(func() error {
		return nil
	})
	assert.NoError(t, err)

	calls := -1
	err = l.Do(func() error {
		calls++
		if calls < 4 {
			return assert.AnError
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 4, calls, "expected 3 calls")

	err = l.Do(func() error {
		return assert.AnError
	})
	assert.EqualError(t, err, assert.AnError.Error())

	l = &retry.Retry{
		Attempts: -1,
	}
	err = l.Do(func() error {
		t.Error("didn't expect to be called")
		return nil
	})
	assert.NoError(t, err)
}

func testLoopDoZero(t *testing.T) {
	t.Parallel()
	l := &retry.Retry{}
	calls := 0
	err := l.Do(func() error {
		calls++
		return nil
	})
	assert.NoError(t, err)
	assert.Zero(t, calls, "expected zero calls")

	l.Attempts = 10
	err = l.Do(func() error {
		calls++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, calls, 1, "expected 1 call")
}

func testLoopDoStopping(t *testing.T) {
	t.Run("WithValue", testLoopDoStoppingValue)
	t.Run("WithPointer", testLoopDoStoppingPointer)
}

func testLoopDoStoppingValue(t *testing.T) {
	t.Parallel()
	l := &retry.Retry{
		Attempts: 10,
	}
	calls := 0
	wantErr := errors.New("stop error")
	err := l.Do(func() error {
		calls++
		if calls >= 4 {
			return retry.StopError{Err: wantErr}
		}
		return assert.AnError
	})
	assert.Equal(t, 4, calls)
	assert.Equal(t, err, wantErr)
}

func testLoopDoStoppingPointer(t *testing.T) {
	t.Parallel()
	l := &retry.Retry{
		Attempts: 10,
	}
	calls := 0
	wantErr := errors.New("stop error")
	err := l.Do(func() error {
		calls++
		if calls >= 4 {
			return &retry.StopError{Err: wantErr}
		}
		return assert.AnError
	})
	assert.Equal(t, 4, calls)
	assert.Equal(t, err, wantErr)
}

func testLoopDoPanic(t *testing.T) {
	t.Parallel()
	l := &retry.Retry{
		Attempts: 3,
	}
	calls := -1
	err := l.Do(func() error {
		calls++
		if calls < l.Attempts-1 {
			panic(assert.AnError)
		}
		return nil
	})
	assert.NoError(t, err)

	err = l.Do(func() error {
		panic(assert.AnError)
	})
	assert.Contains(t, errors.Cause(err).Error(), assert.AnError.Error())
}

func testLoopDoSleep(t *testing.T) {
	t.Run("StandardMethod", testLoopDoSleepStandardMethod)
	t.Run("IncrementalMethod", testLoopDoSleepIncrementalMethod)
}

func testLoopDoSleepStandardMethod(t *testing.T) {
	t.Parallel()
	// In this setup, we delay 10 times, So in 1 second there would be 10 calls.
	count := 0
	delay := 100 * time.Millisecond
	l := &retry.Retry{
		Attempts: 10,
		Delay:    delay,
	}

	started := time.Now()
	err := l.Do(func() error {
		count++
		return assert.AnError
	})
	finished := time.Now()

	assert.Equal(t, l.Attempts, count)
	assert.Equal(t, assert.AnError, errors.Cause(err))

	expected := started.Add(time.Second)
	assert.WithinDurationf(t, expected, finished, delay,
		"expected to take 1s, got %s", finished.Sub(started))
}

func testLoopDoSleepIncrementalMethod(t *testing.T) {
	t.Run("UnderSecond", testLoopDoSleepIncrementalMethodUnderSecond)
	t.Run("OverSecond", testLoopDoSleepIncrementalMethodOverSecond)
	t.Run("Zero", testLoopDoSleepIncrementalMethodZero)
}

func testLoopDoSleepIncrementalMethodUnderSecond(t *testing.T) {
	t.Parallel()
	// In this setup, the delays would be (almost) 100, 200, 300, 400. So in almost
	// 1 second there would be 4 calls. There is a 4*delay amount of wiggle added.
	delay := 100 * time.Millisecond
	l := &retry.Retry{
		Attempts: 4,
		Delay:    delay,
		Method:   retry.IncrementalDelay,
	}

	count := 0
	started := time.Now()
	err := l.Do(func() error {
		count++
		return assert.AnError
	})
	finished := time.Now()
	expected := started.Add(time.Second)

	assert.Equal(t, l.Attempts, count)
	assert.Equal(t, assert.AnError, errors.Cause(err))
	assert.True(t, finished.After(expected),
		fmt.Sprintf("wanted to take more than %s, took %s", expected.Sub(started), finished.Sub(started)),
	)
	assert.True(t, finished.Before(expected.Add(6*delay)),
		fmt.Sprintf("take (%s) more than expected: %s", finished.Sub(started), expected.Add(6*delay)),
	)
}

func testLoopDoSleepIncrementalMethodOverSecond(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("slow test")
	}
	l := &retry.Retry{
		Attempts: 2,
		Delay:    10 * time.Second,
		Method:   retry.IncrementalDelay,
	}

	count := 0
	started := time.Now()
	err := l.Do(func() error {
		count++
		return assert.AnError
	})
	finished := time.Now()
	expected := started.Add(3 * time.Second)

	assert.Equal(t, assert.AnError, errors.Cause(err))
	assert.Equal(t, l.Attempts, count)

	assert.True(t, finished.After(expected),
		fmt.Sprintf("wanted to take more than %s, took %s", expected.Sub(started), finished.Sub(started)),
	)
	assert.True(t, finished.Before(expected.Add(2*time.Second)),
		fmt.Sprintf("take (%s) more than expected: %s", finished.Sub(started), 4*time.Second),
	)
}

func testLoopDoSleepIncrementalMethodZero(t *testing.T) {
	t.Parallel()
	l := &retry.Retry{
		Attempts: 50,
		Method:   retry.IncrementalDelay,
	}

	count := 0
	started := time.Now()
	err := l.Do(func() error {
		count++
		return assert.AnError
	})
	finished := time.Now()
	assert.Equal(t, assert.AnError, errors.Cause(err))
	assert.Equal(t, l.Attempts, count)

	expected := started.Add(time.Second)
	assert.True(t, finished.Before(expected),
		fmt.Sprintf("take (%s) more than expected: %s", finished.Sub(started), time.Second),
	)
}

func testLoopDoMultiFunc(t *testing.T) {
	t.Parallel()
	t.Run("FirstErrors", testLoopDoMultiFuncFirstErrors)
	t.Run("SecondErrors", testLoopDoMultiFuncSecondErrors)
	t.Run("NoErrors", testLoopDoMultiFuncNoErrors)
}

func testLoopDoMultiFuncFirstErrors(t *testing.T) {
	t.Parallel()
	l := &retry.Retry{
		Attempts: 3,
	}
	err := l.Do(func() error {
		return assert.AnError
	}, func() error {
		t.Error("should not be called")
		return nil
	})
	assert.Equal(t, assert.AnError, errors.Cause(err))
}

func testLoopDoMultiFuncSecondErrors(t *testing.T) {
	t.Parallel()
	l := &retry.Retry{
		Attempts: 3,
	}

	calls := 0
	err := l.Do(func() error {
		calls++
		return nil
	}, func() error {
		return assert.AnError
	})
	assert.Equal(t, assert.AnError, errors.Cause(err))
	assert.Equal(t, 3, calls)
}

func testLoopDoMultiFuncNoErrors(t *testing.T) {
	t.Parallel()
	l := &retry.Retry{
		Attempts: 3,
	}

	calls := 0
	err := l.Do(func() error {
		calls++
		return nil
	}, func() error {
		calls++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, calls)
}

func testLoopDoErrorIs(t *testing.T) {
	t.Parallel()
	r := &retry.Retry{
		Attempts: 1,
	}
	err := r.Do(func() error {
		return &retry.StopError{
			Err: &retry.StopError{Err: assert.AnError},
		}
	})

	assert.True(t, errors.Is(err, assert.AnError), err)
}
