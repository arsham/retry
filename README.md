# Retry

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/arsham/retry)
[![go.dev reference](https://img.shields.io/badge/godoc-reference-5272B4)](https://pkg.go.dev/github.com/arsham/retry?tab=doc)
[![Build Status](https://github.com/arsham/retry/actions/workflows/go.yml/badge.svg)](https://github.com/arsham/retry/actions/workflows/go.yml)
[![Coverage Status](https://codecov.io/gh/arsham/retry/branch/master/graph/badge.svg)](https://codecov.io/gh/arsham/retry)
[![Go Report Card](https://goreportcard.com/badge/github.com/arsham/retry)](https://goreportcard.com/report/github.com/arsham/retry)

This library supports `Go >= 1.22` by getting `github.com/arsham/retry/v3`. For
older versions (`Go >= 1.20`) import `github.com/arsham/retry/v2`, or older Go
versions import `github.com/arsham/retry`.

`Retry` calls your function, and if it errors it calls it again with a delay.
Eventually it returns the last error or nil if one call is successful.

```go
r := &retry.Retry{
	Attempts: 666,
	Delay:    time.Millisecond,
}
err := r.Do(func() error {
	// do some work.
	return nil
})
```

You can provide multiple functions:

```go
err := r.Do(func() error {
    return nil
}, func() error {
    return nil
}}
```

You can use the `DoContext` and pass a context object to stop when the context
is cancelled:

```go
err := r.DoContext(ctx, func() error {
	return nil
})
```

If you want to stop retrying you can return a special error:

```go
err := r.Do(func() error {
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
r := &retry.Retry{
	Attempts: 666,
	Delay:    10 * time.Millisecond,
	Method:   retry.IncrementalDelay,
}
err := r.Do(func() error {
	if specialCase {
		return &retry.StopError{
			Err: errors.New("a special stop"),
		}
	}
	return nil
})
```

## License

Use of this source code is governed by the Apache 2.0 license. License can be
found in the [LICENSE](./LICENSE) file.

<!--
vim: foldlevel=1
-->
