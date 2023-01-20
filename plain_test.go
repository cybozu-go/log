package log

import (
	"bytes"
	"fmt"
	"math"
	"testing"
	"time"
	"unicode/utf8"
)

func TestAppendPlain(t *testing.T) {
	t.Parallel()

	buf := make([]byte, 0, 4096)

	b, _ := appendPlain(buf, nil)
	if got, want := string(b), "null"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, _ = appendLogfmt(buf, 100)
	if got, want := string(b), "100"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, _ = appendPlain(buf, false)
	if got, want := string(b), "false"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, _ = appendPlain(buf, true)
	if got, want := string(b), "true"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, _ = appendPlain(buf, int16(-12345))
	if got, want := string(b), "-12345"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, _ = appendPlain(buf, float32(3.14159))
	if got, want := string(b), "3.14159"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, _ = appendPlain(buf, 3.14159)
	if got, want := string(b), "3.14159"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, _ = appendPlain(buf, math.NaN())
	if got, want := string(b), "NaN"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, _ = appendPlain(buf, math.Inf(1))
	if got, want := string(b), "+Inf"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, _ = appendPlain(buf, math.Inf(-1))
	if got, want := string(b), "-Inf"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	b, err := appendPlain(buf, []string{"abc", "def"})
	if err != nil {
		t.Error(err)
	} else {
		if !bytes.Contains(b, []byte("abc")) {
			t.Error(`!bytes.Contain(b, "abc")`)
		}
		if !bytes.Contains(b, []byte("def")) {
			t.Error(`!bytes.Contain(b, "def")`)
		}
	}

	b, err = appendPlain(buf, map[string]int{
		"abc": 100,
		"def": -1,
	})
	if err != nil {
		t.Error(err)
	} else {
		if !bytes.Contains(b, []byte("abc")) {
			t.Error(`!bytes.Contain(b, "abc")`)
		}
		if !bytes.Contains(b, []byte("-1")) {
			t.Error(`!bytes.Contain(b, "-1")`)
		}
	}

	invalidUtf8 := "hello" + string([]byte{0x80})
	b, err = appendPlain(buf, invalidUtf8)
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

	b, err = appendPlain(buf, fmt.Errorf(invalidUtf8))
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

}

const (
	testPlainLog1 = `2001-12-03T13:45:01.123456Z localhost tag1 debug: "test message"` + "\n"
	testPlainLog2 = `2001-12-03T13:45:01.123456Z localhost tag2 debug: "test message" secret=true` + "\n"
	testPlainLog3 = `2001-12-03T13:45:01.123456Z localhost tag2 debug: "test message" secret=false` + "\n"
	testPlainLog4 = `2001-12-03T13:45:01.123456Z localhost tag3 debug: "test message" a=1 b=2 c=3 d=4` + "\n"
)

func TestPlainFormat1(t *testing.T) {
	t.Parallel()

	l := NewLogger()
	l.SetTopic("tag1")

	ts := time.Date(2001, time.December, 3, 13, 45, 1, 123456789, time.UTC)
	f := PlainFormat{"localhost"}
	b := make([]byte, 0, 4096)

	if buf, err := f.Format(b, l, ts, LvDebug, "test message", nil); err != nil {
		t.Error(err)
	} else {
		if got, want := string(buf), testPlainLog1; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestPlainFormat2(t *testing.T) {
	t.Parallel()

	l := NewLogger()
	l.SetTopic("tag2")
	l.SetDefaults(map[string]interface{}{FnSecret: true})

	ts := time.Date(2001, time.December, 3, 13, 45, 1, 123456789, time.UTC)
	f := PlainFormat{"localhost"}
	b := make([]byte, 0, 4096)

	if buf, err := f.Format(b, l, ts, LvDebug, "test message", nil); err != nil {
		t.Error(err)
	} else {
		if got, want := string(buf), testPlainLog2; got != want {
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
		if got, want := string(buf), testPlainLog3; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestPlainFormat3(t *testing.T) {
	t.Parallel()

	l := NewLogger()
	l.SetTopic("tag3")

	ts := time.Date(2001, time.December, 3, 22, 45, 1, 123456789,
		time.FixedZone("Asia/Tokyo", 9*3600))
	f := PlainFormat{"localhost"}
	b := make([]byte, 0, 4096)

	fields := map[string]interface{}{
		"a": 1, "b": 2, "c": 3, "d": 4,
	}

	if buf, err := f.Format(b, l, ts, LvDebug, "test message", fields); err != nil {
		t.Error(err)
	} else {
		if got, want := string(buf), testPlainLog4; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}
