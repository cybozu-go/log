package log

import (
	"encoding"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"
	"unicode/utf8"
)

// JSONFormat implements Formatter for JSON Lines.
//
// https://jsonlines.org/
type JSONFormat struct {
	// Utsname can normally be left blank.
	// If not empty, the string is used instead of the hostname.
	// Utsname must match this regexp: ^[a-z][a-z0-9-]*$
	Utsname string
}

// String returns "json".
func (f JSONFormat) String() string {
	return "json"
}

// Format implements Formatter.Format.
func (f JSONFormat) Format(buf []byte, l *Logger, t time.Time, severity int,
	msg string, fields map[string]interface{}) ([]byte, error) {
	var err error

	buf = append(buf, `{"topic":"`...)
	buf = append(buf, l.Topic()...)
	buf = append(buf, `","logged_at":"`...)
	buf = t.UTC().AppendFormat(buf, RFC3339Micro)
	buf = append(buf, `","severity":`...)
	if ss, ok := severityMap[severity]; ok {
		buf = append(buf, '"')
		buf = append(buf, ss...)
		buf = append(buf, '"')
	} else {
		buf = strconv.AppendInt(buf, int64(severity), 10)
	}
	buf = append(buf, `,"utsname":"`...)
	if len(f.Utsname) > 0 {
		buf = append(buf, f.Utsname...)
	} else {
		buf = append(buf, utsname...)
	}
	buf = append(buf, `","message":`...)
	buf, err = appendJSON(buf, msg)
	if err != nil {
		return nil, err
	}

	for k, v := range fields {
		if !IsValidKey(k) {
			return nil, ErrInvalidKey
		}
		buf = append(buf, `,"`...)
		buf = append(buf, k...)
		buf = append(buf, `":`...)
		buf, err = appendJSON(buf, v)
		if err != nil {
			return nil, err
		}
	}

	for k, v := range l.Defaults() {
		if _, ok := fields[k]; ok {
			continue
		}
		buf = append(buf, `,"`...)
		buf = append(buf, k...)
		buf = append(buf, `":`...)
		buf, err = appendJSON(buf, v)
		if err != nil {
			return nil, err
		}
	}

	return append(buf, "}\n"...), nil
}

func appendJSON(buf []byte, v interface{}) ([]byte, error) {
	var err error

	switch t := v.(type) {
	case nil:
		return append(buf, "null"...), nil
	case bool:
		return strconv.AppendBool(buf, t), nil
	case time.Time:
		buf = append(buf, '"')
		buf = t.UTC().AppendFormat(buf, RFC3339Micro)
		return append(buf, '"'), nil
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
		return appendFloat(buf, float64(t), 32), nil
	case float64:
		return appendFloat(buf, t, 64), nil
	case string:
		return appendString(buf, t), nil
	case json.Marshaler:
		s, err := t.MarshalJSON()
		if err != nil {
			return nil, err
		}
		// normalize for JSON Lines
		for i, b := range s {
			if b == '\n' || b == '\r' {
				s[i] = ' '
			}
		}
		return append(buf, s...), nil
	case encoding.TextMarshaler:
		// TextMarshaler encodes into UTF-8 string.
		s, err := t.MarshalText()
		if err != nil {
			return nil, err
		}
		return appendString(buf, string(s)), nil
	case error:
		return appendString(buf, t.Error()), nil
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
				buf = append(buf, ',')
			}
			buf, err = appendJSON(buf, iter.Key().String())
			if err != nil {
				return nil, err
			}
			buf = append(buf, ':')
			buf, err = appendJSON(buf, iter.Value().Interface())
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
				buf = append(buf, ',')
			}
			buf, err = appendJSON(buf, value.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			first = false
		}
		return append(buf, ']'), nil
	}

	// other types are just formatted as a string with "%#v".
	return appendJSON(buf, fmt.Sprintf("%#v", v))
}

func appendFloat(buf []byte, v float64, bitSize int) []byte {
	if math.IsNaN(v) {
		return append(buf, `"NaN"`...)
	} else if math.IsInf(v, 1) {
		return append(buf, `"+Inf"`...)
	} else if math.IsInf(v, -1) {
		return append(buf, `"-Inf"`...)
	} else {
		return strconv.AppendFloat(buf, v, 'f', -1, bitSize)
	}
}

// Ref: encodeState#string in encoding/json.
func appendString(buf []byte, s string) []byte {
	buf = append(buf, '"')
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if ' ' <= b && b <= '~' && b != '"' && b != '\\' {
				i++
				continue
			}
			if start < i {
				buf = append(buf, s[start:i]...)
			}
			switch b {
			case '\\':
				buf = append(buf, `\\`...)
			case '"':
				buf = append(buf, `\"`...)
			case '\n':
				buf = append(buf, `\n`...)
			case '\r':
				buf = append(buf, `\r`...)
			case '\t':
				buf = append(buf, `\t`...)
			default:
				const hex = "0123456789abcdef"
				buf = append(buf, '\\', 'u', '0', '0', hex[b>>4], hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				buf = append(buf, s[start:i]...)
			}
			buf = append(buf, `\ufffd`...)
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		buf = append(buf, s[start:]...)
	}
	return append(buf, '"')
}
