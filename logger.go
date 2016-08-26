package log

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

var (
	pool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, maxLogSize)
		},
	}

	utsname string
)

func init() {
	hname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	utsname = hname
}

// Logger is a collection of properties how to output logs.
// Properties are initially set by NewLogger.  They can be customized
// later by Logger methods.
type Logger struct {
	topic     atomic.Value
	threshold int32
	defaults  atomic.Value
	format    atomic.Value

	mu     sync.Mutex
	output io.Writer
}

// NewLogger constructs a new Logger struct.
//
// Attributes are initialized as follows:
//    Topic:     path.Base(os.Args[0])
//    Threshold: LvInfo
//    Formatter: PlainFormat
//    Output:    os.Stderr
//    Defaults:  nil
func NewLogger() *Logger {
	l := &Logger{
		output: os.Stderr,
	}
	l.SetTopic(normalizeTopic(path.Base(os.Args[0])))
	l.SetThreshold(LvInfo)
	l.SetDefaults(nil)
	l.SetFormatter(PlainFormat{})
	return l
}

func normalizeTopic(n string) string {
	// Topic must match [.a-z0-9-]+
	topic := strings.Map(func(r rune) rune {
		switch {
		case r == '.' || r == '-':
			return r
		case r >= '0' && r < '9':
			return r
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r < 'Z':
			return r + ('a' - 'A')
		default:
			return '-'
		}
	}, n)
	if len(topic) > maxTopicLength {
		return topic[:maxTopicLength]
	}
	return topic
}

// Topic returns the topic for the logger.
func (l *Logger) Topic() string {
	return l.topic.Load().(string)
}

// SetTopic sets a new topic for the logger.
// topic must not be empty.  Too long topic may be shortened automatically.
func (l *Logger) SetTopic(topic string) {
	if len(topic) == 0 {
		panic("Empty tag")
	}

	l.mu.Lock()
	l.topic.Store(topic)
	l.mu.Unlock()
}

// Threshold returns the current threshold of the logger.
func (l *Logger) Threshold() int {
	return int(atomic.LoadInt32(&l.threshold))
}

// Enabled returns true if the log for the given level will be logged.
// This can be used to avoid futile computation for logs being ignored.
//
//    if log.Enabled(log.LvDebug) {
//        log.Debug("message", map[string]interface{}{
//            "debug info": "...",
//        })
//    }
func (l *Logger) Enabled(level int) bool {
	return level <= l.Threshold()
}

// SetThreshold sets the threshold for the logger.
// level must be a pre-defined constant such as LvInfo.
func (l *Logger) SetThreshold(level int) {
	l.mu.Lock()
	atomic.StoreInt32(&l.threshold, int32(level))
	l.mu.Unlock()
}

// SetThresholdByName sets the threshold for the logger by the level name.
func (l *Logger) SetThresholdByName(n string) error {
	var level int
	switch n {
	case "critical", "crit":
		level = LvCritical
	case "error":
		level = LvError
	case "warning", "warn":
		level = LvWarn
	case "information", "info":
		level = LvInfo
	case "debug":
		level = LvDebug
	default:
		return fmt.Errorf("No such level: %s", n)
	}
	l.SetThreshold(level)
	return nil
}

// SetDefaults sets default field values for the logger.
// Setting nil effectively clear the defaults.
func (l *Logger) SetDefaults(d map[string]interface{}) error {
	for key := range d {
		if !IsValidKey(key) {
			return ErrInvalidKey
		}
	}

	l.mu.Lock()
	l.defaults.Store(d)
	l.mu.Unlock()
	return nil
}

// Defaults returns default field values.
func (l *Logger) Defaults() map[string]interface{} {
	return l.defaults.Load().(map[string]interface{})
}

// SetFormatter sets log formatter.
func (l *Logger) SetFormatter(f Formatter) {
	l.mu.Lock()
	l.format.Store(&f)
	l.mu.Unlock()
}

// Formatter returns the current log formatter.
func (l *Logger) Formatter() Formatter {
	return *l.format.Load().(*Formatter)
}

// SetOutput sets io.Writer for log output.
// Setting nil disables log output.
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	l.output = w
	l.mu.Unlock()
}

type logWriter struct {
	buf     []byte
	logfunc func(p []byte) (n int, err error)
}

func (w *logWriter) Write(p []byte) (int, error) {
	tbuf := p
	if len(w.buf) > 0 {
		tbuf = append(w.buf, p...)
	}
	written, err := w.logfunc(tbuf)
	n := written - len(w.buf)
	if err != nil {
		if n < 0 {
			return 0, err
		}
		return n, err
	}

	w.buf = w.buf[:0]
	remain := len(tbuf) - written
	if remain == 0 {
		return n, nil
	}
	if cap(w.buf) < remain {
		return n, errors.New("too long")
	}
	w.buf = append(w.buf, tbuf[n:]...)
	return len(p), nil
}

// Writer returns an io.Writer.
// Each line written in the writer will be logged to the logger
// with the given severity.
func (l *Logger) Writer(severity int) io.Writer {
	logfunc := func(p []byte) (n int, err error) {
		for len(p) > 0 {
			eol := bytes.IndexByte(p, '\n')
			if eol == -1 {
				return
			}
			ln := eol + 1
			err = l.Log(severity, string(p[:eol]), nil)
			if err != nil {
				return
			}
			n += ln
			p = p[ln:]
		}
		return
	}

	return &logWriter{
		buf:     make([]byte, 0, maxLogSize/2),
		logfunc: logfunc,
	}
}

// Log outputs a log message with additional fields.
// fields can be nil.
func (l *Logger) Log(severity int, msg string, fields map[string]interface{}) error {
	if severity > l.Threshold() {
		return nil
	}

	t := time.Now()
	buf := pool.Get().([]byte)
	defer pool.Put(buf)

	if l.output != nil {
		b, err := l.Formatter().Format(buf, l, t, severity, msg, fields)
		if err != nil {
			return err
		}

		l.mu.Lock()
		defer l.mu.Unlock()
		if _, err := l.output.Write(b); err != nil {
			fmt.Fprintf(os.Stderr, "logger output causes an error: %v", err)
			return errors.Wrap(err, "Logger.Log")
		}
	}

	return nil
}

// Critical outputs a critical log.
// fields can be nil.
func (l *Logger) Critical(msg string, fields map[string]interface{}) error {
	return l.Log(LvCritical, msg, fields)
}

// Error outputs an error log.
// fields can be nil.
func (l *Logger) Error(msg string, fields map[string]interface{}) error {
	return l.Log(LvError, msg, fields)
}

// Warn outputs a warning log.
// fields can be nil.
func (l *Logger) Warn(msg string, fields map[string]interface{}) error {
	return l.Log(LvWarn, msg, fields)
}

// Info outputs an informational log.
// fields can be nil.
func (l *Logger) Info(msg string, fields map[string]interface{}) error {
	return l.Log(LvInfo, msg, fields)
}

// Debug outputs a debug log.
// fields can be nil.
func (l *Logger) Debug(msg string, fields map[string]interface{}) error {
	return l.Log(LvDebug, msg, fields)
}

// WriteThrough writes data through to the underlying writer.
func (l *Logger) WriteThrough(data []byte) error {
	l.mu.Lock()
	_, err := l.output.Write(data)
	l.mu.Unlock()
	return err
}
