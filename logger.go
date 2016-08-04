package log

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
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
	mu        sync.RWMutex
	topic     string
	threshold int
	defaults  map[string]interface{}
	format    Formatter
	output    io.Writer
}

// NewLogger constructs a new Logger struct.
// The struct is initialized as follows:
//    topic:        the program name normalized for tag spec.
//    threshold: LvInfo, i.e., debug logs are disabled.
//    format:    PlainFormat
//    output:    os.Stderr
//    defaults:  no default fields.
func NewLogger() *Logger {
	return &Logger{
		topic:     normalizeTopic(path.Base(os.Args[0])),
		threshold: LvInfo,
		defaults:  nil,
		format:    PlainFormat{},
		output:    os.Stderr,
	}
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
	l.mu.RLock()
	topic := l.topic
	l.mu.RUnlock()
	return topic
}

// SetTopic sets a new topic for the logger.
// topic must not be empty.  Too long topic may be shortened automatically.
func (l *Logger) SetTopic(topic string) {
	if len(topic) == 0 {
		panic("Empty tag")
	}

	l.mu.Lock()
	l.topic = normalizeTopic(topic)
	l.mu.Unlock()
}

// Threshold returns the current threshold of the logger.
func (l *Logger) Threshold() int {
	l.mu.RLock()
	th := l.threshold
	l.mu.RUnlock()
	return th
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
	l.mu.RLock()
	enabled := (level <= l.threshold)
	l.mu.RUnlock()
	return enabled
}

// SetThreshold sets the threshold for the logger.
// level must be a pre-defined constant such as LvInfo.
func (l *Logger) SetThreshold(level int) {
	l.mu.Lock()
	l.threshold = level
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
	l.defaults = d
	l.mu.Unlock()
	return nil
}

// Defaults returns default field values.
func (l *Logger) Defaults() map[string]interface{} {
	l.mu.RLock()
	defaults := l.defaults
	l.mu.RUnlock()
	return defaults
}

// SetFormatter sets log formatter.
func (l *Logger) SetFormatter(f Formatter) {
	l.mu.Lock()
	l.format = f
	l.mu.Unlock()
}

// SetOutput sets io.Writer for log output.
// Setting nil disables log output.
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	l.output = w
	l.mu.Unlock()
}

// Writer returns an io.Writer.
// Each line written in the writer will be logged to the logger
// with the given severity.
func (l *Logger) Writer(severity int) io.Writer {
	r, w := io.Pipe()

	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			if err := l.Log(severity, scanner.Text(), nil); err != nil {
				r.CloseWithError(err)
			}
		}
		if err := scanner.Err(); err != nil {
			r.CloseWithError(err)
		}
	}()

	return w
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
		b, err := l.format.Format(buf, l, t, severity, msg, fields)
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
