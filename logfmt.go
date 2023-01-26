package log

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Logfmt implements Formatter for logfmt format.
//
// https://brandur.org/logfmt
// https://gist.github.com/kr/0e8d5ee4b954ce604bb2
type Logfmt struct {
	// Utsname can normally be left blank.
	// If not empty, the string is used instead of the hostname.
	// Utsname must match this regexp: ^[a-z][a-z0-9-]*$
	Utsname string
}

// String returns "logfmt".
func (f Logfmt) String() string {
	return "logfmt"
}

// Format implements Formatter.Format.
func (f Logfmt) Format(buf []byte, l *Logger, t time.Time, severity int,
	msg string, fields map[string]interface{}) ([]byte, error) {
	var err error

	buf = append(buf, "topic="...)
	buf = append(buf, l.Topic()...)
	buf = append(buf, " logged_at="...)
	buf = t.UTC().AppendFormat(buf, RFC3339Micro)
	buf = append(buf, " severity="...)
	if ss, ok := severityMap[severity]; ok {
		buf = append(buf, ss...)
	} else {
		buf = strconv.AppendInt(buf, int64(severity), 10)
	}
	buf = append(buf, " utsname="...)
	if len(f.Utsname) > 0 {
		buf = append(buf, f.Utsname...)
	} else {
		buf = append(buf, utsname...)
	}
	buf = append(buf, " message="...)
	buf, err = appendLogfmt(buf, msg)
	if err != nil {
		return nil, err
	}

	for k, v := range fields {
		if !IsValidKey(k) {
			return nil, ErrInvalidKey
		}
		buf = append(buf, ' ')
		buf = append(buf, k...)
		buf = append(buf, '=')
		buf, err = appendLogfmt(buf, v)
		if err != nil {
			return nil, err
		}
	}

	for k, v := range l.Defaults() {
		if _, ok := fields[k]; ok {
			continue
		}
		buf = append(buf, ' ')
		buf = append(buf, k...)
		buf = append(buf, '=')
		buf, err = appendLogfmt(buf, v)
		if err != nil {
			return nil, err
		}
	}

	return append(buf, '\n'), nil
}

func appendLogfmt(buf []byte, v interface{}) ([]byte, error) {
	var err error

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
	}

	value := reflect.ValueOf(v)
	typ := value.Type()
	kind := typ.Kind()

	// string-keyed maps
	if kind == reflect.Map && typ.Key().Kind() == reflect.String {
		buf = append(buf, '{')
		first := true
		for iter := value.MapRange(); iter.Next(); {
			if !first {
				buf = append(buf, ' ')
			}
			key := iter.Key().String()
			if regexpValidKey.MatchString(key) {
				buf = append(buf, key...)
			} else {
				buf, err = appendLogfmt(buf, key)
				if err != nil {
					return nil, err
				}
			}
			buf = append(buf, '=')
			buf, err = appendLogfmt(buf, iter.Value().Interface())
			if err != nil {
				return nil, err
			}
			first = false
		}
		return append(buf, '}'), nil
	}

	// slices and arrays
	if kind == reflect.Slice || kind == reflect.Array {
		buf = append(buf, '[')
		first := true
		for i := 0; i < value.Len(); i++ {
			if !first {
				buf = append(buf, ' ')
			}
			buf, err = appendLogfmt(buf, value.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			first = false
		}
		return append(buf, ']'), nil
	}

	// other types are just formatted as string with "%v".
	return appendLogfmt(buf, fmt.Sprintf("%v", v))
}
