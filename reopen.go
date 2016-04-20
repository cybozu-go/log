package log

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
)

// WriterOpener opens a new io.WriteCloser.
type WriterOpener interface {
	Open() (io.WriteCloser, error)
}

type reopenWriter struct {
	lock    sync.Mutex
	lastErr error
	writer  io.WriteCloser
}

// NewReopenWriter constructs a io.Writer that reopens inner io.WriteCloser
// when signals are notified.
func NewReopenWriter(opener WriterOpener, sig ...os.Signal) (io.Writer, error) {
	w, err := opener.Open()
	if err != nil {
		return nil, err
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, sig...)
	r := &reopenWriter{
		writer: w,
	}
	reopen := func() {
		r.lock.Lock()
		defer r.lock.Unlock()
		if r.writer != nil {
			err := r.writer.Close()
			// io.Closer does not guarantee that it is safe to call it twice.
			r.writer = nil
			if err != nil {
				r.lastErr = err
				return
			}
		}
		w, err := opener.Open()
		if err != nil {
			r.lastErr = err
			return
		}
		r.writer = w
		r.lastErr = nil
	}
	go func() {
		for range c {
			reopen()
		}
	}()
	return r, nil
}

// Write calles inner writes.
// If some error has happened when re-opening, this reports the error.
func (r *reopenWriter) Write(p []byte) (n int, err error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.lastErr != nil {
		err = fmt.Errorf("unusable due to %v", r.lastErr)
		return
	}
	return r.writer.Write(p)
}
