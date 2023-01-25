package log

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"
	"unicode/utf8"
)

const (
	testLog1 = `topic=tag1 logged_at=2001-12-03T13:45:01.123456Z severity=debug utsname=localhost message="test message"` + "\n"
	testLog2 = `topic=tag2 logged_at=2001-12-03T13:45:01.123456Z severity=debug utsname=localhost message="test message" secret=true` + "\n"
	testLog3 = `topic=tag2 logged_at=2001-12-03T13:45:01.123456Z severity=debug utsname=localhost message="test message" secret=false` + "\n"
)

func TestAppendLogfmt(t *testing.T) {
	t.Parallel()

	buf := make([]byte, 0, 4096)

	b, _ := appendLogfmt(buf, nil)
	if got, want := string(b), "null"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, _ = appendLogfmt(buf, 100)
	if got, want := string(b), "100"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, _ = appendLogfmt(buf, false)
	if got, want := string(b), "false"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, _ = appendLogfmt(buf, true)
	if got, want := string(b), "true"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, _ = appendLogfmt(buf, `"abc `)
	if got, want := string(b), `"\"abc "`; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, _ = appendLogfmt(buf, float32(3.14159))
	if got, want := string(b), "3.14159"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, _ = appendLogfmt(buf, 3.14159)
	if got, want := string(b), "3.14159"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, _ = appendLogfmt(buf, math.NaN())
	if got, want := string(b), "NaN"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, _ = appendLogfmt(buf, math.Inf(1))
	if got, want := string(b), "+Inf"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, _ = appendLogfmt(buf, math.Inf(-1))
	if got, want := string(b), "-Inf"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, _ = appendLogfmt(buf, []int{-100, 100, 20000})
	if got, want := string(b), "[-100 100 20000]"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, _ = appendLogfmt(buf, []int64{-100, 100, 20000})
	if got, want := string(b), "[-100 100 20000]"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, _ = appendLogfmt(buf, []string{"abc", "def"})
	if got, want := string(b), `["abc" "def"]`; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, _ = appendLogfmt(buf, map[string]interface{}{
		"abc":     123,
		"def ghi": nil,
	})
	if string(b) != `{abc=123 "def ghi"=null}` &&
		string(b) != `{"def ghi"=null abc=123}` {
		t.Error("failed to format map[string]interface{}")
	}

	invalidUtf8 := "hello" + string([]byte{0x80})
	b, err := appendLogfmt(buf, invalidUtf8)
	if err != nil {
		t.Error(err)
	} else {
		if !utf8.ValidString(string(b)) {
			t.Error(`!utf8.ValidString(b)`)
		}
		if bytes.Contains(b, []byte("x80")) {
			t.Error(`bytes.Contains(b, "x80")`)
		}
		if !bytes.Contains(b, []byte("hello")) {
			t.Error(`!bytes.Contains(b, "hello")`)
		}
	}

	b, err = appendLogfmt(buf, fmt.Errorf(invalidUtf8))
	if err != nil {
		t.Error(err)
	} else {
		if !utf8.ValidString(string(b)) {
			t.Error(`!utf8.ValidString(b)`)
		}
		if bytes.Contains(b, []byte("x80")) {
			t.Error(`bytes.Contains(b, "x80")`)
		}
		if !bytes.Contains(b, []byte("hello")) {
			t.Error(`!bytes.Contains(b, "hello")`)
		}
	}

	input := "\x00\x00\x00\x00"
	expected := `"\x00\x00\x00\x00"`
	buf = make([]byte, 1, len(expected))
	_, err = appendLogfmt(buf, input)
	if !errors.Is(err, ErrTooLarge) {
		t.Errorf("got: %#v, want: %#v", err, ErrTooLarge)
	}
}

func TestLogfmt1(t *testing.T) {
	t.Parallel()

	l := NewLogger()
	l.SetTopic("tag1")

	ts := time.Date(2001, time.December, 3, 13, 45, 1, 123456789, time.UTC)
	f := Logfmt{"localhost"}
	b := make([]byte, 0, 4096)

	if buf, err := f.Format(b, l, ts, LvDebug, "test message", nil); err != nil {
		t.Error(err)
	} else {
		if got, want := string(buf), testLog1; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	// Time is always formatted in UTC.
	ts = time.Date(2001, time.December, 3, 22, 45, 1, 123456789,
		time.FixedZone("Asia/Tokyo", 9*3600))
	if buf, err := f.Format(b, l, ts, LvDebug, "test message", nil); err != nil {
		t.Error(err)
	} else {
		if got, want := string(buf), testLog1; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestLogfmt2(t *testing.T) {
	t.Parallel()

	l := NewLogger()
	l.SetTopic("tag2")
	l.SetDefaults(map[string]interface{}{FnSecret: true})

	ts := time.Date(2001, time.December, 3, 13, 45, 1, 123456789, time.UTC)
	f := Logfmt{"localhost"}
	b := make([]byte, 0, 4096)

	if buf, err := f.Format(b, l, ts, LvDebug, "test message", nil); err != nil {
		t.Error(err)
	} else {
		if got, want := string(buf), testLog2; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	// override the default
	fields := map[string]interface{}{
		FnSecret: false,
	}
	if buf, err := f.Format(b, l, ts, LvDebug, "test message", fields); err != nil {
		t.Error(err)
	} else {
		if got, want := string(buf), testLog3; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}
