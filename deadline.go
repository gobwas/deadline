package deadline

import (
	"sync"
	"time"
)

var ErrDeadline = deadlineError{}

// Do is a helper method that runs callback in a separate goroutine with given
// deadline. If deadline expires earlier than callback returns, it returns
// ErrDeadline. In other cases returned error is nil.
func Do(deadline time.Time, cb func()) error {
	d := Deadline{}
	d.Set(deadline)
	return d.Do(cb)
}

// Deadline contains deadline handling logic. It is intended to be much like
// net.Conn SetDeadline() logic. That is, it is possible to set deadlines
// sequentially overwriting previous value and moving point of time when Done()
// channel will be closed.
type Deadline struct {
	// Goer allows to set up custom goroutine starter. It is useful when client
	// uses some goroutine pool.
	Goer GoFunc

	mu    sync.Mutex
	done  chan struct{}
	timer *time.Timer
}

// Do runs callback in a separate goroutine. It returns when callcack returns
// or when deadline exceeded. In case of deadline, it returns ErrDeadline.
// In other cases returned error is always nil.
func (d *Deadline) Do(cb func()) error {
	var (
		done = d.Done()
		ok   = acquireDone()
	)
	goer(d.Goer, done, func() {
		defer close(ok)
		cb()
	})
	select {
	case <-ok:
		releaseDone(ok)
		return nil
	case <-done:
		return ErrDeadline
	}
}

// Done returns a channel which closure means deadline expiration.
func (d *Deadline) Done() <-chan struct{} {
	d.mu.Lock()
	if d.done == nil {
		d.done = acquireDone()
	}
	done := d.done
	d.mu.Unlock()
	return done
}

// Set sets up new deadline point. If previous deadline was not reached yet,
// but Done() channel was retreived before this Set(), that channel will be
// closed when new deadline will be expired.
//
// It is safe to call Set() from different goroutines.
func (d *Deadline) Set(t time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// We need to guarantee that nobody else owns d.done for writing.
	if d.timer != nil && !d.timer.Stop() {
		<-d.done
	}
	if t.IsZero() {
		return
	}
	if d.done == nil {
		d.done = acquireDone()
	} else {
		select {
		case <-d.done:
			// If done become closed, we need to reinitiate it by a new struct.

			// Writing d.done is safe here without synchronization because we
			// always await for the timer goroutine exit or timer stop (see
			// d.timer.Stop() above).
			d.done = acquireDone()
		default:
		}
	}
	n := t.Sub(time.Now())
	if n < 0 {
		// Close d.done immediately because deadline already exceeded.
		close(d.done)
		return
	}
	if d.timer == nil {
		d.timer = time.AfterFunc(n, func() {
			close(d.done)
		})
	} else {
		// We do not check d.timer.Stop() here cause it is not a problem, if
		// deadline has been reached and some routine was cancelled.
		d.timer.Reset(n)
	}
}

// GoFunc runs given callback in a separate goroutine. If by any reason it is
// not possible to start new goroutine, and the given cancelation channel
// become non-empty (closed) implementation must not try to start the goroutine
// and exit immediately.
type GoFunc func(<-chan struct{}, func())

func goer(g GoFunc, cancel <-chan struct{}, task func()) {
	if g == nil {
		go task()
	} else {
		g(cancel, task)
	}
}

type deadlineError struct{ error }

func (d deadlineError) Error() string   { return "deadline exceeded" }
func (d deadlineError) Timeout() bool   { return true }
func (d deadlineError) Temporary() bool { return true }

var donePool sync.Pool

func acquireDone() chan struct{} {
	if v := donePool.Get(); v != nil {
		return v.(chan struct{})
	}
	return make(chan struct{})
}

func releaseDone(ch chan struct{}) {
	donePool.Put(ch)
}
