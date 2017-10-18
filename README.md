# Deadline

[![GoDoc][godoc-image]][godoc-url]

> Tiny deadline helper for Go.

# Overview

This little package is intended to bring similar functionality as `net.Conn`'s
`SetDeadline()` methods family.

# Usage

```go
package main

import (
	
	"github.com/gobwas/deadline"
)

func main() {
	d := deadline.Deadline{}
	d.Set(time.Now().Add(time.Second))
	err := d.Do(func() {
		// This code will be running in a separate gorotuine.
	})
	if err != nil {
		// Even if code still running, after a second we ge an error here and
		// could handle it somehow.
	}
}
```

[godoc-image]: https://godoc.org/github.com/gobwas/deadline?status.svg
[godoc-url]:   https://godoc.org/github.com/gobwas/deadline
