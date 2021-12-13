//go:build !windows
// +build !windows

package log

import (
	"bytes"
	"io"
	"io/ioutil"
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
	if s != "foobar1234" {
		t.Errorf("written data should be \"foobar1234\" but \"%v\"", s)
	}
}

func TestFileReopener(t *testing.T) {
	t.Parallel()

	f, err := ioutil.TempFile("", "gotest")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	g, err := ioutil.TempFile("", "gotest")
	if err != nil {
		t.Fatal(err)
	}
	g.Close()
	defer os.Remove(g.Name())

	w, err := NewFileReopener(f.Name(), syscall.SIGUSR2)
	if err != nil {
		t.Fatal(err)
	}

	lg := NewLogger()
	lg.SetOutput(w)
	lg.SetFormatter(Logfmt{})
	lg.Critical("hoge", nil)
	syscall.Kill(os.Getpid(), syscall.SIGUSR2)
	time.Sleep(100 * time.Millisecond)

	if err := os.Rename(f.Name(), g.Name()); err != nil {
		t.Fatal(err)
	}

	syscall.Kill(os.Getpid(), syscall.SIGUSR2)
	time.Sleep(100 * time.Millisecond)
	lg.Critical("fuga", nil)
	syscall.Kill(os.Getpid(), syscall.SIGUSR2)
	time.Sleep(100 * time.Millisecond)

	if hoge, err := ioutil.ReadFile(g.Name()); err != nil {
		t.Error(err)
	} else {
		if !bytes.Contains(hoge, []byte("hoge")) {
			t.Error("g must contain hoge")
		}
		if bytes.Contains(hoge, []byte("fuga")) {
			t.Error("g must not contain fuga")
		}
	}

	if fuga, err := ioutil.ReadFile(f.Name()); err != nil {
		t.Error(err)
	} else {
		if bytes.Contains(fuga, []byte("hoge")) {
			t.Error("f must not contain hoge")
		}
		if !bytes.Contains(fuga, []byte("fuga")) {
			t.Error("f must contain fuga")
		}
	}
}

func TestFileReopenerCorrection(t *testing.T) {
	t.Parallel()

	f, err := ioutil.TempFile("", "gotest")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("abc"))
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	g, err := ioutil.TempFile("", "gotest")
	if err != nil {
		t.Fatal(err)
	}
	g.Close()
	defer os.Remove(g.Name())

	w, err := NewFileReopener(f.Name(), syscall.SIGHUP)
	if err != nil {
		t.Fatal(err)
	}

	lg := NewLogger()
	lg.SetOutput(w)
	lg.SetFormatter(Logfmt{})
	lg.Critical("hoge", nil)

	if err := os.Rename(f.Name(), g.Name()); err != nil {
		t.Fatal(err)
	}

	syscall.Kill(os.Getpid(), syscall.SIGHUP)
	time.Sleep(100 * time.Millisecond)
	lg.Critical("fuga", nil)
	syscall.Kill(os.Getpid(), syscall.SIGHUP)
	time.Sleep(100 * time.Millisecond)

	if hoge, err := ioutil.ReadFile(g.Name()); err != nil {
		t.Error(err)
	} else {
		if !bytes.HasPrefix(hoge, []byte("abc\n")) {
			t.Error(`!bytes.HasPrefix(hoge, []byte("abc\n"))`)
		}
	}

	if fuga, err := ioutil.ReadFile(f.Name()); err != nil {
		t.Error(err)
	} else {
		if bytes.HasPrefix(fuga, []byte("\n")) {
			t.Error(`bytes.HasPrefix(fuga, []byte("\n"))`)
		}
	}
}
