package retry_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/arsham/retry/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetry_Do(t *testing.T) {
	t.Parallel()
	t.Run("Return", testRetryDoReturn)
	t.Run("Zero", testRetryDoZero)
	t.Run("Stopping", testRetryDoStopping)
	t.Run("Panic", testRetryDoPanic)
	t.Run("Sleep", testRetryDoSleep)
	t.Run("MultiFunc", testRetryDoMultiFunc)
	t.Run("ErrorIs", testRetryDoErrorIs)
}

func testRetryDoReturn(t *testing.T) {
	t.Parallel()
	l := &retry.Retry{
		Attempts: 10,
	}
	err := l.Do(func() error {
		return nil
	})
	require.NoError(t, err)

	calls := -1
	err = l.Do(func() error {
		calls++
		if calls < 4 {
			return assert.AnError
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 4, calls, "expected 3 calls")

	err = l.Do(func() error {
		return assert.AnError
	})
	require.EqualError(t, err, assert.AnError.Error())

	l = &retry.Retry{
		Attempts: -1,
	}
	err = l.Do(func() error {
		t.Error("didn't expect to be called")
		return nil
	})
	require.NoError(t, err)
}

func testRetryDoZero(t *testing.T) {
	t.Parallel()
	l := &retry.Retry{}
	calls := 0
	err := l.Do(func() error {
		calls++
		return nil
	})
	require.NoError(t, err)
	assert.Zero(t, calls, "expected zero calls")

	l.Attempts = 10
	err = l.Do(func() error {
		calls++
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, calls, "expected 1 call")
}

func testRetryDoStopping(t *testing.T) {
	t.Parallel()
	l := &retry.Retry{
		Attempts: 10,
	}
	calls := 0
	err := l.Do(func() error {
		calls++
		if calls >= 4 {
			return &retry.StopError{Err: assert.AnError}
		}
		return assert.AnError
	})
	assert.Equal(t, 4, calls)
	assert.Equal(t, err, assert.AnError)
}

func testRetryDoPanic(t *testing.T) {
	t.Parallel()
	l := &retry.Retry{
		Attempts: 3,
	}
	calls := -1
	require.NotPanics(t, func() {
		err := l.Do(func() error {
			calls++
			if calls < l.Attempts-1 {
				panic(assert.AnError)
			}
			return nil
		})
		require.NoError(t, err)
	})

	require.NotPanics(t, func() {
		err := l.Do(func() error {
			panic(assert.AnError)
		})
		require.ErrorIs(t, err, assert.AnError)
	})

	require.NotPanics(t, func() {
		err := l.Do(func() error {
			panic(assert.AnError.Error())
		})
		require.Error(t, err)
	})
}

func testRetryDoSleep(t *testing.T) {
	t.Run("StandardMethod", testRetryDoSleepStandardMethod)
	t.Run("IncrementalMethod", testRetryDoSleepIncrementalMethod)
	t.Run("IncrementalMaxMethod", testRetryDoSleepIncrementalMaxMethod)
}

func testRetryDoSleepStandardMethod(t *testing.T) {
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
	require.ErrorIs(t, err, assert.AnError)

	expected := started.Add(time.Second)
	assert.WithinDurationf(t, expected, finished, delay,
		"expected to take 1s, got %s", finished.Sub(started))
}

func testRetryDoSleepIncrementalMethod(t *testing.T) {
	t.Run("UnderSecond", testRetryDoSleepIncrementalMethodUnderSecond)
	t.Run("OverSecond", testRetryDoSleepIncrementalMethodOverSecond)
	t.Run("Zero", testRetryDoSleepIncrementalMethodZero)
}

func testRetryDoSleepIncrementalMethodUnderSecond(t *testing.T) {
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
	require.ErrorIs(t, err, assert.AnError)
	assert.True(t, finished.After(expected),
		fmt.Sprintf("wanted to take more than %s, took %s", expected.Sub(started), finished.Sub(started)),
	)
	assert.True(t, finished.Before(expected.Add(6*delay)),
		fmt.Sprintf("took (%s) more than expected: %s", finished.Sub(started), expected.Add(6*delay)),
	)
}

func testRetryDoSleepIncrementalMethodOverSecond(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("slow test")
	}
	rty := &retry.Retry{
		Attempts: 2,
		Delay:    10 * time.Second,
		Method:   retry.IncrementalDelay,
	}

	count := 0
	started := time.Now()
	err := rty.Do(func() error {
		count++
		return assert.AnError
	})
	finished := time.Now()
	expected := started.Add(3 * time.Second)

	require.ErrorIs(t, err, assert.AnError)
	assert.Equal(t, rty.Attempts, count)

	assert.True(t, finished.After(expected),
		fmt.Sprintf("wanted to take more than %s, took %s", expected.Sub(started), finished.Sub(started)),
	)
	assert.True(t, finished.Before(expected.Add(2*time.Second)),
		fmt.Sprintf("took (%s) more than expected: %s", finished.Sub(started), 4*time.Second),
	)
}

func testRetryDoSleepIncrementalMethodZero(t *testing.T) {
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
	require.ErrorIs(t, err, assert.AnError)
	assert.Equal(t, l.Attempts, count)

	expected := started.Add(time.Second)
	assert.True(t, finished.Before(expected),
		fmt.Sprintf("took (%s) more than expected: %s", finished.Sub(started), time.Second),
	)
}

func testRetryDoSleepIncrementalMaxMethod(t *testing.T) {
	t.Run("UnderSecond", testRetryDoSleepIncrementalMaxMethodUnderSecond)
	t.Run("OverSecond", testRetryDoSleepIncrementalMaxMethodOverTwoSeconds)
	t.Run("Zero", testRetryDoSleepIncrementalMaxMethodZero)
}

func testRetryDoSleepIncrementalMaxMethodUnderSecond(t *testing.T) {
	t.Parallel()
	// In this setup, the delays would be (almost) 100, 200, 300, 300. So in
	// almost 900 ms there would be 4 calls. There is a 4*delay amount of
	// wiggle added.
	delay := 100 * time.Millisecond
	rty := &retry.Retry{
		Attempts: 4,
		Delay:    delay,
		Method:   retry.IncrementalDelayMax(delay * 3),
	}

	count := 0
	started := time.Now()
	err := rty.Do(func() error {
		count++
		return assert.AnError
	})
	finished := time.Now()
	expected := started.Add(900 * time.Millisecond)

	assert.Equal(t, rty.Attempts, count)
	require.ErrorIs(t, err, assert.AnError)
	assert.True(t, finished.After(expected),
		fmt.Sprintf("wanted to take more than %s, took %s", expected.Sub(started), finished.Sub(started)),
	)
	assert.True(t, finished.Before(expected.Add(6*delay)),
		fmt.Sprintf("took (%s) more than expected: %s", finished.Sub(started), expected.Add(6*delay)),
	)
}

func testRetryDoSleepIncrementalMaxMethodOverTwoSeconds(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("slow test")
	}
	l := &retry.Retry{
		Attempts: 2,
		Delay:    10 * time.Second,
		Method:   retry.IncrementalDelayMax(2 * time.Second),
	}

	count := 0
	started := time.Now()
	err := l.Do(func() error {
		count++
		return assert.AnError
	})
	finished := time.Now()
	expected := started.Add(4 * time.Second)

	require.ErrorIs(t, err, assert.AnError)
	assert.Equal(t, l.Attempts, count)

	assert.True(t, finished.After(expected),
		fmt.Sprintf("wanted to take more than %s, took %s", expected.Sub(started), finished.Sub(started)),
	)
	assert.True(t, finished.Before(expected.Add(2*2*time.Second)),
		fmt.Sprintf("took (%s) more than expected: %s", finished.Sub(started), 4*time.Second),
	)
}

func testRetryDoSleepIncrementalMaxMethodZero(t *testing.T) {
	t.Parallel()
	rty := &retry.Retry{
		Attempts: 50,
		Method:   retry.IncrementalDelayMax(time.Second / 2),
	}

	count := 0
	started := time.Now()
	err := rty.Do(func() error {
		count++
		return assert.AnError
	})
	finished := time.Now()
	require.ErrorIs(t, err, assert.AnError)
	assert.Equal(t, rty.Attempts, count)

	expected := started.Add(time.Second)
	assert.True(t, finished.Before(expected),
		fmt.Sprintf("took (%s) more than expected: %s", finished.Sub(started), time.Second),
	)
}

func testRetryDoMultiFunc(t *testing.T) {
	t.Parallel()
	t.Run("FirstErrors", testRetryDoMultiFuncFirstErrors)
	t.Run("SecondErrors", testRetryDoMultiFuncSecondErrors)
	t.Run("NoErrors", testRetryDoMultiFuncNoErrors)
}

func testRetryDoMultiFuncFirstErrors(t *testing.T) {
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
	assert.ErrorIs(t, err, assert.AnError)
}

func testRetryDoMultiFuncSecondErrors(t *testing.T) {
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
	require.ErrorIs(t, err, assert.AnError)
	assert.Equal(t, 3, calls)
}

func testRetryDoMultiFuncNoErrors(t *testing.T) {
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
	require.NoError(t, err)
	assert.Equal(t, 2, calls)
}

func testRetryDoErrorIs(t *testing.T) {
	t.Parallel()
	r := &retry.Retry{
		Attempts: 1,
	}
	err := r.Do(func() error {
		return &retry.StopError{
			Err: &retry.StopError{Err: assert.AnError},
		}
	})

	assert.ErrorIs(t, err, assert.AnError, err)
}

func TestRetry_DoContext(t *testing.T) {
	t.Parallel()
	// Since the Do() uses the DoContext() method internally, there is no point
	// duplicating the effort. We just test some cases.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := &retry.Retry{
		Attempts: 100,
		Delay:    time.Millisecond,
	}
	calls := 0
	err := r.DoContext(ctx, func() error {
		calls++
		if calls < 3 {
			return assert.AnError
		}
		return nil
	}, func() error {
		if calls > 5 {
			cancel()
		}
		return assert.AnError
	})

	assert.Equal(t, 6, calls)
	require.ErrorIs(t, err, context.Canceled, err)

	err = r.DoContext(ctx, func() error {
		cancel()
		return nil
	}, func() error {
		panic("should not happen")
	})

	assert.ErrorIs(t, err, context.Canceled, err)
}
