package log

import (
	"fmt"
	"strconv"
	"time"
)

var (
	severityMap = map[int]string{
		LvCritical: "crit",
		LvError:    "error",
		LvWarn:     "warn",
		LvInfo:     "info",
		LvDebug:    "debug",
	}
)

func appendLogfmt(buf []byte, v interface{}) ([]byte, error) {
	switch t := v.(type) {
	case nil:
		return append(buf, "null"...), nil
	case bool:
		return strconv.AppendBool(buf, t), nil
	case int:
		return strconv.AppendInt(buf, int64(t), 10), nil
	case int64:
		return strconv.AppendInt(buf, t, 10), nil
	case time.Time:
		return t.UTC().AppendFormat(buf, time.RFC3339Nano), nil
	case string:
		return strconv.AppendQuote(buf, t), nil
	case []byte:
		return strconv.AppendQuoteToASCII(buf, string(t)), nil
	case []int:
		buf = append(buf, byte('['))
		for i, n := range t {
			if i > 0 {
				buf = append(buf, ", "...)
			}
			buf = strconv.AppendInt(buf, int64(n), 10)
		}
		return append(buf, byte(']')), nil
	case []int64:
		buf = append(buf, byte('['))
		for i, n := range t {
			if i > 0 {
				buf = append(buf, ", "...)
			}
			buf = strconv.AppendInt(buf, n, 10)
		}
		return append(buf, byte(']')), nil
	case []string:
		buf = append(buf, byte('['))
		for i, s := range t {
			if i > 0 {
				buf = append(buf, ", "...)
			}
			buf = strconv.AppendQuote(buf, s)
		}
		return append(buf, byte(']')), nil
	default:
		return nil, ErrInvalidData
	}
}

func logfmt(l *Logger, t time.Time, severity int, msg string,
	fields map[string]interface{}) ([]byte, error) {
	buf := l.buffer[:0]
	buf = append(buf, "tag="...)
	buf = append(buf, l.tag...)
	buf = append(buf, " logged_at="...)
	buf = t.UTC().AppendFormat(buf, time.RFC3339Nano)
	buf = append(buf, " severity="...)
	if ss, ok := severityMap[severity]; ok {
		buf = append(buf, ss...)
	} else {
		buf = strconv.AppendInt(buf, int64(severity), 10)
	}
	buf = append(buf, " utsname="...)
	buf = append(buf, l.utsname...)
	buf = append(buf, " message="...)
	if (len(buf) + len(msg)) > maxLogSize {
		return nil, ErrTooLarge
	}
	buf = strconv.AppendQuote(buf, msg)

	for k, v := range fields {
		if reservedKey(k) {
			continue
		}
		if len(k) > maxFieldNameLength {
			return nil, fmt.Errorf("Too long field name: %s", k)
		}

		buf = append(buf, byte(' '))
		buf = append(buf, k...)
		buf = append(buf, byte('='))
		tbuf, err := appendLogfmt(buf, v)
		if err != nil {
			return nil, err
		}
		buf = tbuf
		if len(buf) > maxLogSize {
			return nil, ErrTooLarge
		}
	}

	for k, v := range l.defaults {
		if reservedKey(k) {
			continue
		}
		if _, ok := fields[k]; ok {
			continue
		}
		if len(k) > maxFieldNameLength {
			return nil, fmt.Errorf("Too long field name: %s", k)
		}

		buf = append(buf, byte(' '))
		buf = append(buf, k...)
		buf = append(buf, byte('='))
		tbuf, err := appendLogfmt(buf, v)
		if err != nil {
			return nil, err
		}
		buf = tbuf
		if len(buf) > maxLogSize {
			return nil, ErrTooLarge
		}
	}

	buf = append(buf, byte('\n'))

	return buf, nil
}
