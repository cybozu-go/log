package log

import (
	"bytes"
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
	if string(b) != "null" {
		t.Error(string(b) + " != null")
	}

	b, _ = appendLogfmt(buf, 100)
	if string(b) != "100" {
		t.Error(string(b) + " != 100")
	}

	b, _ = appendLogfmt(buf, false)
	if string(b) != "false" {
		t.Error(string(b) + " != false")
	}

	b, _ = appendLogfmt(buf, true)
	if string(b) != "true" {
		t.Error(string(b) + " != true")
	}

	b, _ = appendLogfmt(buf, `"abc `)
	if string(b) != `"\"abc "` {
		t.Error(string(b) + ` != "\"abc "`)
	}

	b, _ = appendLogfmt(buf, float32(3.14159))
	if string(b) != "3.14159" {
		t.Error(string(b) + ` != "3.14159"`)
	}

	b, _ = appendLogfmt(buf, 3.14159)
	if string(b) != "3.14159" {
		t.Error(string(b) + ` != "3.14159"`)
	}

	b, _ = appendLogfmt(buf, math.NaN())
	if string(b) != "NaN" {
		t.Error(string(b) + ` != "NaN"`)
	}

	b, _ = appendLogfmt(buf, math.Inf(1))
	if string(b) != "+Inf" {
		t.Error(string(b) + ` != "+Inf"`)
	}

	b, _ = appendLogfmt(buf, math.Inf(-1))
	if string(b) != "-Inf" {
		t.Error(string(b) + ` != "-Inf"`)
	}

	b, _ = appendLogfmt(buf, []int{-100, 100, 20000})
	if string(b) != "[-100 100 20000]" {
		t.Error(string(b) + ` != "[-100 100 20000]"`)
	}

	b, _ = appendLogfmt(buf, []int64{-100, 100, 20000})
	if string(b) != "[-100 100 20000]" {
		t.Error(string(b) + ` != "[-100 100 20000]"`)
	}

	b, _ = appendLogfmt(buf, []string{"abc", "def"})
	if string(b) != `["abc" "def"]` {
		t.Error(string(b) + ` != ["abc" "def"]`)
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
		if string(buf) != testLog1 {
			t.Error(string(buf) + " != " + testLog1)
		}
	}

	// Time is always formatted in UTC.
	ts = time.Date(2001, time.December, 3, 22, 45, 1, 123456789,
		time.FixedZone("Asia/Tokyo", 9*3600))
	if buf, err := f.Format(b, l, ts, LvDebug, "test message", nil); err != nil {
		t.Error(err)
	} else {
		if string(buf) != testLog1 {
			t.Error(string(buf) + " != " + testLog1)
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
		if string(buf) != testLog2 {
			t.Error(string(buf) + " != " + testLog2)
		}
	}

	// override the default
	fields := map[string]interface{}{
		FnSecret: false,
	}
	if buf, err := f.Format(b, l, ts, LvDebug, "test message", fields); err != nil {
		t.Error(err)
	} else {
		if string(buf) != testLog3 {
			t.Error(string(buf) + " != " + testLog3)
		}
	}
}
