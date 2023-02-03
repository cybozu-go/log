package log

import (
	"encoding"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// PlainFormat implements Formatter to generate plain log messages.
//
// A plain log message looks like:
// DATETIME SEVERITY UTSNAME TOPIC MESSAGE [OPTIONAL FIELDS...]
type PlainFormat struct {
	// Utsname can normally be left blank.
	// If not empty, the string is used instead of the hostname.
	// Utsname must match this regexp: ^[a-z][a-z0-9-]*$
	Utsname string
}

// String returns "plain".
func (f PlainFormat) String() string {
	return "plain"
}

// Format implements Formatter.Format.
func (f PlainFormat) Format(buf []byte, l *Logger, t time.Time, severity int,
	msg string, fields map[string]interface{}) ([]byte, error) {
	var err error

	buf = t.UTC().AppendFormat(buf, RFC3339Micro)
	buf = append(buf, ' ')
	if len(f.Utsname) > 0 {
		buf = append(buf, f.Utsname...)
	} else {
		buf = append(buf, utsname...)
	}
	buf = append(buf, ' ')
	buf = append(buf, l.Topic()...)
	buf = append(buf, ' ')
	if ss, ok := severityMap[severity]; ok {
		buf = append(buf, ss...)
	} else {
		buf = strconv.AppendInt(buf, int64(severity), 10)
	}
	buf = append(buf, ": "...)
	buf, err = appendPlain(buf, msg)
	if err != nil {
		return nil, err
	}

	if len(fields) > 0 {
		keys := make([]string, 0, len(fields))
		for k := range fields {
			if !IsValidKey(k) {
				return nil, ErrInvalidKey
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			buf = append(buf, ' ')
			buf = append(buf, k...)
			buf = append(buf, '=')
			buf, err = appendPlain(buf, fields[k])
			if err != nil {
				return nil, err
			}
		}
	}

	for k, v := range l.Defaults() {
		if _, ok := fields[k]; ok {
			continue
		}
		buf = append(buf, ' ')
		buf = append(buf, k...)
		buf = append(buf, '=')
		buf, err = appendPlain(buf, v)
		if err != nil {
			return nil, err
		}
	}

	return append(buf, '\n'), nil
}

func appendPlain(buf []byte, v interface{}) ([]byte, error) {
	switch t := v.(type) {
	case nil:
		return append(buf, "null"...), nil
	case bool:
		return strconv.AppendBool(buf, t), nil
	case time.Time:
		return t.UTC().AppendFormat(buf, RFC3339Micro), nil
	case int:
		return strconv.AppendInt(buf, int64(t), 10), nil
	case int8:
		return strconv.AppendInt(buf, int64(t), 10), nil
	case int16:
		return strconv.AppendInt(buf, int64(t), 10), nil
	case int32:
		return strconv.AppendInt(buf, int64(t), 10), nil
	case int64:
		return strconv.AppendInt(buf, t, 10), nil
	case uint:
		return strconv.AppendUint(buf, uint64(t), 10), nil
	case uint8:
		return strconv.AppendUint(buf, uint64(t), 10), nil
	case uint16:
		return strconv.AppendUint(buf, uint64(t), 10), nil
	case uint32:
		return strconv.AppendUint(buf, uint64(t), 10), nil
	case uint64:
		return strconv.AppendUint(buf, t, 10), nil
	case float32:
		return strconv.AppendFloat(buf, float64(t), 'f', -1, 32), nil
	case float64:
		return strconv.AppendFloat(buf, t, 'f', -1, 64), nil
	case string:
		if !utf8.ValidString(t) {
			// the next line replaces invalid characters.
			t = strings.ToValidUTF8(t, string(utf8.RuneError))
		}
		return strconv.AppendQuote(buf, t), nil
	case encoding.TextMarshaler:
		// TextMarshaler encodes into UTF-8 string.
		s, err := t.MarshalText()
		if err != nil {
			return nil, err
		}
		return strconv.AppendQuote(buf, string(s)), nil
	case error:
		s := t.Error()
		if !utf8.ValidString(s) {
			// the next line replaces invalid characters.
			s = strings.ToValidUTF8(s, string(utf8.RuneError))
		}
		return strconv.AppendQuote(buf, s), nil
	default:
		// other types are just formatted as string with "%v".
		return appendPlain(buf, fmt.Sprintf("%v", t))
	}
}
