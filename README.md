# Retry

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/arsham/retry)
[![go.dev reference](https://img.shields.io/badge/godoc-reference-5272B4)](https://pkg.go.dev/github.com/arsham/retry?tab=doc)
[![Build Status](https://travis-ci.com/arsham/retry.svg?branch=master)](https://travis-ci.com/arsham/retry)
[![Coverage Status](https://codecov.io/gh/arsham/retry/branch/master/graph/badge.svg)](https://codecov.io/gh/arsham/retry)
[![Go Report Card](https://goreportcard.com/badge/github.com/arsham/retry)](https://goreportcard.com/report/github.com/arsham/retry)

This library supports `Go >= 1.13`.

`Retry` calls your function, and if it errors it calls it again with a delay.
Eventually it returns the last error or nil if one call is successful.

```go
l := &retry.Retry{
    Attempts: 666,
    Delay:    time.Millisecond,
}
err := l.Do(func() error {
    // do some work.
    return nil
})
```

If you want to stop retrying you can return a special error:

```go
err := l.Do(func() error {
    if specialCase {
        return &retry.StopError{
            Err: errors.New("a special stop"),
        }
    }
    return nil
})
```

The standard behaviour is to delay the amount you set. You can pass any function
with this signature to change the delay behaviour:

```go
func(attempt int, delay time.Duration) time.Duration
```

You can also pass the `retry.IncrementalDelay` function that would increase the
delay with a jitter to prevent [Thundering
herd](https://en.wikipedia.org/wiki/Thundering_herd_problem).


```go
l := &retry.Retry{
    Attempts: 666,
    Delay:    10 * time.Millisecond,
    Method:   retry.IncrementalDelay,
}
err := l.Do(func() error {
    if specialCase {
        return &retry.StopError{
            Err: errors.New("a special stop"),
        }
    }
    return nil
})
```
