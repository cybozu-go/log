package log

import (
	"bytes"
	"strings"
	"testing"
)

func TestNormalizeTopic(t *testing.T) {
	t.Parallel()

	if normalizeTopic("abc.-def") != "abc.-def" {
		t.Error("abc.-def")
	}
	if normalizeTopic("Abc.-Def") != "abc.-def" {
		t.Error("Abc.-Def")
	}
	if normalizeTopic("Abc._Def") != "abc.-def" {
		t.Error("Abc._Def")
	}
}

func TestLogger(t *testing.T) {
	t.Parallel()

	l := NewLogger()
	if l.Topic() != "log.test" {
		t.Error(`topic != "log.test"`)
	}
	l.SetTopic("hoge")
	if l.Topic() != "hoge" {
		t.Error("failed to set topic")
	}

	buf := new(bytes.Buffer)
	l.SetOutput(buf)
	l.SetFormatter(Logfmt{})
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
		if !strings.Contains(s, "topic=hoge") {
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
		FnSecret: true,
	})
	buf.Reset()
	if err := l.Debug("hoge", nil); err != nil {
		t.Error(err)
	} else {
		s := string(buf.Bytes())
		if !strings.Contains(s, `secret=true`) {
			t.Error("failed to include default fields")
		}
	}

	buf.Reset()
	fields := map[string]interface{}{
		FnSecret: true,
		"custom": 10000,
	}
	if err := l.Debug("hoge", fields); err != nil {
		t.Error(err)
	} else {
		s := string(buf.Bytes())
		if !strings.Contains(s, `secret=true`) {
			t.Error("failed to specify fields")
		}
		if !strings.Contains(s, `custom=10000`) {
			t.Error("failed to specify custom field")
		}
	}
}
