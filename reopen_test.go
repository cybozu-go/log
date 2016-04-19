package log

import (
	"bytes"
	"io"
	"syscall"
	"testing"
)

type closeBuffer struct {
	bytes.Buffer
}

func (*closeBuffer) Close() error {
	return nil
}

func TestReopenWriter(t *testing.T) {
	t.Parallel()

	nopen := 0
	var buf closeBuffer
	opener := func() (io.WriteCloser, error) {
		nopen++
		return &buf, nil
	}

	w := NewReopenWriter(opener, syscall.SIGUSR1)
	w.Write([]byte("foobar"))

	// syscall.Kill(os.Getpid(), syscall.SIGUSR1)
	// send a signal directly because signal may be delayed.
	w.(*reopenWriter).sigc <- syscall.SIGUSR1

	w.Write([]byte("1234"))

	if nopen != 2 {
		t.Errorf("number of open should be 2 but %v", nopen)
	}
	s := buf.String()
	if "foobar1234" != s {
		t.Errorf("written data should be \"foobar1234\" but \"%v\"", s)
	}
}
