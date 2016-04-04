package log

import (
	"strconv"
	"testing"
	"time"
)

const (
	testData1 = "\x93\xa4tag1\x0a\x84\xa9logged_at\xd2\x00\x98\x96\xa0\xa8severity\a\xa7utsname\xa9localhost\xa7message\xactest message"
	testData2 = "\x93\xa4tag2\x0a\x85\xa9logged_at\xd2\x00\x98\x96\xa0\xa8severity\a\xa7utsname\xa9localhost\xa7message\xactest message\xa6secret\xc3"
	testData3 = "\x93\xa4tag2\x0a\x85\xa9logged_at\xd2\x00\x98\x96\xa0\xa8severity\a\xa7utsname\xa9localhost\xa7message\xactest message\xa6secret\xc2"
)

func TestAppendMsgpack(t *testing.T) {
	t.Parallel()

	b := make([]byte, 0, 4096)
	if b2, err := appendMsgpack(b, nil); err != nil {
		t.Error(err)
	} else {
		if string(b2) != "\xc0" {
			t.Error("failed to encode nil")
		}
	}

	if b2, err := appendMsgpack(b, true); err != nil {
		t.Error(err)
	} else {
		if string(b2) != "\xc3" {
			t.Error("failed to encode true")
		}
	}

	if b2, err := appendMsgpack(b, false); err != nil {
		t.Error(err)
	} else {
		if string(b2) != "\xc2" {
			t.Error("failed to encode false")
		}
	}

	if b2, err := appendMsgpack(b, 12); err != nil {
		t.Error(err)
	} else {
		if string(b2) != "\x0c" {
			t.Error("failed to encode 12")
		}
	}

	if b2, err := appendMsgpack(b, 1024); err != nil {
		t.Error(err)
	} else {
		if string(b2) != "\xd1\x04\x00" {
			t.Error("failed to encode 1024")
		}
	}

	if b2, err := appendMsgpack(b, 65537); err != nil {
		t.Error(err)
	} else {
		if string(b2) != "\xd2\x00\x01\x00\x01" {
			t.Error("failed to encode 65537")
		}
	}

	if b2, err := appendMsgpack(b, -1); err != nil {
		t.Error(err)
	} else {
		if string(b2) != "\xd1\xff\xff" {
			t.Error("failed to encode -1")
		}
	}

	if b2, err := appendMsgpack(b, -9223372036854775808); err != nil {
		t.Error(err)
	} else {
		if string(b2) != "\xd3\x80\x00\x00\x00\x00\x00\x00\x00" {
			t.Error("failed to encode INT64_MIN")
		}
	}

	ts := time.Date(1970, time.January, 1, 0, 0, 0, 32000, time.UTC)
	if b2, err := appendMsgpack(b, ts); err != nil {
		t.Error(err)
	} else {
		if string(b2) != "\x20" {
			t.Error("failed to encode time.Time")
		}
	}

	if b2, err := appendMsgpack(b, "Hello World!"); err != nil {
		t.Error(err)
	} else {
		if string(b2) != "\xacHello World!" {
			t.Error("failed to encode Hello World!")
		}
	}

	if b2, err := appendMsgpack(b, "Hello World! Hello World! Hello World!"); err != nil {
		t.Error(err)
	} else {
		if string(b2) != "\xd9\x26Hello World! Hello World! Hello World!" {
			t.Error("failed to encode Hello World! * 3")
		}
	}

	if b2, err := appendMsgpack(b, []byte("\x12\xff")); err != nil {
		t.Error(err)
	} else {
		if string(b2) != "\xc4\x02\x12\xff" {
			t.Error("failed to encode binary data")
		}
	}

	if b2, err := appendMsgpack(b, []int{100, -100, -100000, 100000}); err != nil {
		t.Error(err)
	} else {
		if string(b2) != "\x94\x64\xd1\xff\x9c\xd2\xff\xfe\x79\x60\xd2\x00\x01\x86\xa0" {
			t.Error("failed to encode []int data")
		}
	}

	if b2, err := appendMsgpack(b, []int64{100, -100, -100000, 100000}); err != nil {
		t.Error(err)
	} else {
		if string(b2) != "\x94\x64\xd1\xff\x9c\xd2\xff\xfe\x79\x60\xd2\x00\x01\x86\xa0" {
			t.Error("failed to encode []int64 data")
		}
	}

	if b2, err := appendMsgpack(b, []string{"abc", "def"}); err != nil {
		t.Error(err)
	} else {
		if string(b2) != "\x92\xa3\x61\x62\x63\xa3\x64\x65\x66" {
			t.Error("failed to encode []string data")
		}
	}

	if _, err := appendMsgpack(b, []interface{}{"a", 10}); err != ErrInvalidData {
		t.Error("[]interface{} must be an invalid data")
	}
}

func TestMsgpackFmt1(t *testing.T) {
	t.Parallel()

	l := &Logger{
		utsname:  "localhost",
		tag:      "tag1",
		defaults: nil,
		buffer:   make([]byte, 0, 4096),
	}

	ts := time.Date(1970, time.January, 1, 0, 0, 10, 32000, time.UTC)

	if buf, err := msgpackfmt(l, ts, LvDebug, "test message", nil); err != nil {
		t.Error(err)
	} else {
		if string(buf) != testData1 {
			t.Error(strconv.QuoteToASCII(string(buf)) +
				" != " + strconv.QuoteToASCII(testData1))
		}
	}
}

func TestMsgpackFmt2(t *testing.T) {
	t.Parallel()

	l := &Logger{
		utsname:  "localhost",
		tag:      "tag2",
		defaults: map[string]interface{}{FnSecret: true},
		buffer:   make([]byte, 0, 4096),
	}

	ts := time.Date(1970, time.January, 1, 0, 0, 10, 32000, time.UTC)

	if buf, err := msgpackfmt(l, ts, LvDebug, "test message", nil); err != nil {
		t.Error(err)
	} else {
		if string(buf) != testData2 {
			t.Error(strconv.QuoteToASCII(string(buf)) +
				" != " + strconv.QuoteToASCII(testData2))
		}
	}

	// override the default
	fields := map[string]interface{}{
		FnSecret: false,
	}
	if buf, err := msgpackfmt(l, ts, LvDebug, "test message", fields); err != nil {
		t.Error(err)
	} else {
		if string(buf) != testData3 {
			t.Error(strconv.QuoteToASCII(string(buf)) +
				" != " + strconv.QuoteToASCII(testData3))
		}
	}
}
