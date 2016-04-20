package log

import (
	"bytes"
	"io"
	"os"
	"syscall"
	"testing"
	"time"
)

type bufferOpener struct {
	bytes.Buffer
	nopen  int
	nclose int
}

func (c *bufferOpener) Close() error {
	c.nclose++
	return nil
}

func (c *bufferOpener) Open() (io.WriteCloser, error) {
	c.nopen++
	return c, nil
}

func TestReopenWriter(t *testing.T) {
	t.Parallel()

	var buf bufferOpener
	w, err := NewReopenWriter(&buf, syscall.SIGUSR1)
	if err != nil {
		t.Fatal(err)
	}

	w.Write([]byte("foobar"))

	syscall.Kill(os.Getpid(), syscall.SIGUSR1)
	time.Sleep(time.Second)

	w.Write([]byte("1234"))

	if buf.nopen != 2 {
		t.Errorf("number of open should be 2 but %v", buf.nopen)
	}
	if buf.nclose != 1 {
		t.Errorf("number of close should be 1 but %v", buf.nclose)
	}
	s := buf.String()
	if "foobar1234" != s {
		t.Errorf("written data should be \"foobar1234\" but \"%v\"", s)
	}
}
