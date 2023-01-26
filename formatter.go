package log

import (
	"regexp"
	"time"
)

var (
	regexpValidKey = regexp.MustCompile(`^[_a-z][a-z0-9_]*$`)
)

// Formatter is the interface for log formatters.
type Formatter interface {
	// Format appends formatted log data into buf.
	//
	// Format should return (nil, ErrInvalidKey) if a key in fields is
	// not valid in the sense of IsValidKey().
	Format(buf []byte, l *Logger, t time.Time, severity int,
		msg string, fields map[string]interface{}) ([]byte, error)

	// String returns the formatter name.
	String() string
}

// ReservedKey returns true if k is a field name reserved for log formatters.
func ReservedKey(k string) bool {
	switch k {
	case FnTopic, FnLoggedAt, FnSeverity, FnUtsname, FnMessage:
		return true
	}
	return false
}

// IsValidKey returns true if given key is valid for extra fields.
func IsValidKey(key string) bool {
	return regexpValidKey.MatchString(key) && !ReservedKey(key)
}
