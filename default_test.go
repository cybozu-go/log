package log

import (
	"bytes"
	"strings"
	"testing"

	_log "log"
)

const magicData = "99999999"

func TestDefaultLogger(t *testing.T) {
	t.Parallel()

	l := DefaultLogger()
	buf := new(bytes.Buffer)
	l.SetOutput(buf)
	l.SetFormatter(Logfmt{})

	_log.Print(magicData)
	if buf.Len() == 0 {
		t.Error("failed to take Go's standard logger output")
	} else {
		s := buf.String()
		if !strings.Contains(s, "severity=info") {
			t.Error("severity is not info")
		}
	}

	buf.Reset()
	_log.Print("123 456")
	if buf.Len() == 0 {
		t.Error("failed to take Go's standard logger output")
	} else {
		s := buf.String()
		if strings.Contains(s, magicData) {
			t.Error("logs are not separated")
		}
	}
}
