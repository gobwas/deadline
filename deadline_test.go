package deadline

import (
	"testing"
	"time"
)

func TestDeadline(t *testing.T) {
	for _, test := range []struct {
		delay    time.Duration
		deadline time.Duration
		err      error
	}{
		{
			delay:    time.Millisecond * 10,
			deadline: time.Millisecond,
			err:      ErrDeadline,
		},
		{
			delay:    time.Millisecond * 10,
			deadline: time.Millisecond * 100,
			err:      nil,
		},
	} {
		t.Run("", func(t *testing.T) {
			d := Deadline{}
			d.Set(time.Now().Add(test.deadline))
			ok := make(chan struct{})
			err := d.Do(func() {
				defer close(ok)
				time.Sleep(test.delay)
			})
			if err != test.err {
				t.Errorf("unexpected error: %v; want %v", err, test.err)
			}
			// Avoid races.
			<-ok
		})
	}
}
