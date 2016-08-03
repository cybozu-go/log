package log

import "time"

// PlainFormat implements Formatter to generate plain log messages.
//
// A plain log message looks like:
// DATETIME SEVERITY UTSNAME TOPIC MESSAGE [OPTIONAL FIELDS...]
type PlainFormat struct{}

// Format implements Formatter.Format.
func (f PlainFormat) Format(buf []byte, l *Logger, t time.Time, severity int,
	msg string, fields map[string]interface{}) ([]byte, error) {
	//var err error

	// assume enough capacity for mandatory fields (except for msg).
	buf = t.UTC().AppendFormat(buf, time.RFC3339Nano)

	return append(buf, byte('\n')), nil
}
