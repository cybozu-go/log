package log

import (
	"io"
	"os"
	"os/signal"
)

type reopenWriter struct {
	sigc   chan os.Signal
	opener func() (io.WriteCloser, error)
	writer io.WriteCloser
}

// NewReopenWriter constructs a io.Writer that reopens inner io.WriteCloser.
// opener should open a inner io.WriteCloser, it is called when signals are notified.
func NewReopenWriter(opener func() (io.WriteCloser, error), sig ...os.Signal) io.Writer {
	c := make(chan os.Signal, 1)
	signal.Notify(c, sig...)
	return &reopenWriter{
		c,
		opener,
		nil,
	}
}

// Write calles inner writes.
// If signals are notified, it (re-)opens new io.WriteCloser by opener.
func (r *reopenWriter) Write(p []byte) (n int, err error) {
	select {
	case <-r.sigc:
		if r.writer != nil {
			err = r.writer.Close()
			if err != nil {
				return
			}
			r.writer = nil
		}
	default:
	}
	if r.writer == nil {
		w, e := r.opener()
		if e != nil {
			err = e
			return
		}
		r.writer = w
	}

	return r.writer.Write(p)
}
