package log

import (
	"testing"
	"time"
)

const (
	testLog1 = `topic=tag1 logged_at=2001-12-03T13:45:01.123456Z severity=debug utsname=localhost message="test message"` + "\n"
	testLog2 = `topic=tag2 logged_at=2001-12-03T13:45:01.123456Z severity=debug utsname=localhost message="test message" secret=true` + "\n"
	testLog3 = `topic=tag2 logged_at=2001-12-03T13:45:01.123456Z severity=debug utsname=localhost message="test message" secret=false` + "\n"
)

func TestAppendLogfmt(t *testing.T) {
	t.Parallel()

	b := make([]byte, 0, 4096)
	b, _ = appendLogfmt(b, nil)
	if string(b) != "null" {
		t.Error(string(b) + " != null")
	}

	b = b[:0]
	b, _ = appendLogfmt(b, 100)
	if string(b) != "100" {
		t.Error(string(b) + " != 100")
	}

	b = b[:0]
	b, _ = appendLogfmt(b, false)
	if string(b) != "false" {
		t.Error(string(b) + " != false")
	}

	b = b[:0]
	b, _ = appendLogfmt(b, true)
	if string(b) != "true" {
		t.Error(string(b) + " != true")
	}

	b = b[:0]
	b, _ = appendLogfmt(b, `"abc `)
	if string(b) != `"\"abc "` {
		t.Error(string(b) + ` != "\"abc "`)
	}

	b = b[:0]
	b, _ = appendLogfmt(b, []int{-100, 100, 20000})
	if string(b) != `[-100 100 20000]` {
		t.Error("failed to format int list")
	}

	b = b[:0]
	b, _ = appendLogfmt(b, []int64{-100, 100, 20000})
	if string(b) != `[-100 100 20000]` {
		t.Error("failed to format int64 list")
	}

	b = b[:0]
	b, _ = appendLogfmt(b, []string{"abc", "def"})
	if string(b) != `["abc" "def"]` {
		t.Error("failed to format string list")
	}

	b = b[:0]
	b, _ = appendLogfmt(b, map[string]interface{}{
		"abc":     123,
		"def ghi": nil,
	})
	if string(b) != `{abc=123 "def ghi"=null}` &&
		string(b) != `{"def ghi"=null abc=123}` {
		t.Error("failed to format map[string]interface{}")
	}
}

func TestLogfmt1(t *testing.T) {
	t.Parallel()

	utsname = "localhost"

	l := NewLogger()
	l.SetTopic("tag1")

	ts := time.Date(2001, time.December, 3, 13, 45, 1, 123456789, time.UTC)
	f := Logfmt{}
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

	utsname = "localhost"

	l := NewLogger()
	l.SetTopic("tag2")
	l.SetDefaults(map[string]interface{}{FnSecret: true})

	ts := time.Date(2001, time.December, 3, 13, 45, 1, 123456789, time.UTC)
	f := Logfmt{}
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
