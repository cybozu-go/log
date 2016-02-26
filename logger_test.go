package log

import (
	"bytes"
	"strings"
	"testing"
)

func TestNormalizeTag(t *testing.T) {
	t.Parallel()

	if normalizeTag("abc.-def") != "abc.-def" {
		t.Error("abc.-def")
	}
	if normalizeTag("Abc.-Def") != "abc.-def" {
		t.Error("Abc.-Def")
	}
	if normalizeTag("Abc._Def") != "abc.-def" {
		t.Error("Abc._Def")
	}
}

func TestLogger(t *testing.T) {
	t.Parallel()

	l := NewLogger()
	if len(l.utsname) == 0 {
		t.Error("no utsname")
	}
	if l.Tag() != "log.test" {
		t.Error(`tag != "log.test"`)
	}
	l.SetTag("hoge")
	if l.Tag() != "hoge" {
		t.Error("failed to set tag")
	}

	buf := new(bytes.Buffer)
	l.SetOutput(buf)
	if err := l.Debug("hoge", nil); err != nil {
		t.Error(err)
	}
	if buf.Len() != 0 {
		t.Error("debug log should be ignored")
	}

	l.SetThreshold(LvDebug)
	if err := l.Debug("hoge", nil); err != nil {
		t.Error(err)
	}
	if buf.Len() == 0 {
		t.Error("debug log should not be ignored")
	} else {
		s := string(buf.Bytes())
		if !strings.Contains(s, "tag=hoge") {
			t.Error("Invalid log: " + s)
		}
	}

	if err := l.SetThresholdByName("hoge"); err == nil {
		t.Error("hoge must not be a valid log level")
	}
	if l.SetThresholdByName("critical"); l.Threshold() != LvCritical {
		t.Error("Failed to set threshold as critical")
	}
	if l.SetThresholdByName("crit"); l.Threshold() != LvCritical {
		t.Error("Failed to set threshold as critical")
	}
	if l.SetThresholdByName("debug"); l.Threshold() != LvDebug {
		t.Error("Failed to set threshold as debug")
	}

	l.SetDefaults(map[string]interface{}{
		FnLoggedBy: "logger_test",
	})
	buf.Reset()
	if err := l.Debug("hoge", nil); err != nil {
		t.Error(err)
	} else {
		s := string(buf.Bytes())
		if !strings.Contains(s, `logged_by="logger_test"`) {
			t.Error("failed to include default fields")
		}
	}

	buf.Reset()
	fields := map[string]interface{}{
		FnLoggedBy: "a",
		"_custom":  10000,
	}
	if err := l.Debug("hoge", fields); err != nil {
		t.Error(err)
	} else {
		s := string(buf.Bytes())
		if !strings.Contains(s, `logged_by="a"`) {
			t.Error("failed to specify fields")
		}
		if !strings.Contains(s, `_custom=10000`) {
			t.Error("failed to specify custom field")
		}
	}
}
