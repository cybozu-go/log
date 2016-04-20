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
)

func reservedKey(k string) bool {
	switch k {
	case FnTag, FnLoggedAt, FnSeverity, FnUtsname, FnMessage:
		return true
	}
	return false
}

// Logger is a collection of properties how to output logs.
// Properties are initially set by NewLogger.  They can be customized
// later by Logger methods.
type Logger struct {
	utsname string

	lock      *sync.Mutex
	tag       string
	threshold int
	defaults  map[string]interface{}
	output    io.Writer
	buffer    []byte
	fluentd   *fluentConn
}

// NewLogger constructs a new Logger struct.
// The struct is initialized as follows:
//    tag:       the program name normalized for tag spec.
//    threshold: LvInfo, i.e., debug logs are disabled.
//    output:    os.Stderr
//    defaults:  no default fields.
func NewLogger() *Logger {
	hname, err := os.Hostname()
	if err != nil {
		panic("os.Hostname() returns err: " + err.Error())
	}
	fluentd, err := newFluentConn()
	if err != nil {
		panic("failed to connect fluentd: " + err.Error())
	}

	return &Logger{
		utsname:   hname,
		lock:      new(sync.Mutex),
		tag:       normalizeTag(path.Base(os.Args[0])),
		threshold: LvInfo,
		defaults:  nil,
		output:    os.Stderr,
		buffer:    make([]byte, 0, maxLogSize+4096),
		fluentd:   fluentd,
	}
}

func normalizeTag(n string) string {
	// tags must match [.a-z0-9-]+
	tag := strings.Map(func(r rune) rune {
		switch {
		case r == '.' || r == '-':
			return r
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r < 'Z':
			return r + ('a' - 'A')
		default:
			return '-'
		}
	}, n)
	if len(tag) > maxTagLength {
		return tag[:maxTagLength]
	}
	return tag
}

// Tag returns the current tag for the logger.
func (l *Logger) Tag() string {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.tag
}

// SetTag sets a new tag for the logger.
// tag must not be empty.  Too long tag may be shortened automatically.
func (l *Logger) SetTag(tag string) {
	if len(tag) == 0 {
		panic("Empty tag")
	}

	l.lock.Lock()
	defer l.lock.Unlock()
	l.tag = normalizeTag(tag)
}

// Threshold returns the current threshold of the logger.
func (l *Logger) Threshold() int {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.threshold
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
	l.lock.Lock()
	defer l.lock.Unlock()
	return level <= l.threshold
}

// SetThreshold sets the threshold for the logger.
// level must be a pre-defined constant such as LvInfo.
func (l *Logger) SetThreshold(level int) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.threshold = level
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
func (l *Logger) SetDefaults(d map[string]interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.defaults = d
}

// SetOutput sets io.Writer for log output.
// Setting nil disables log output.
func (l *Logger) SetOutput(w io.Writer) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.output = w
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
	t := time.Now()

	l.lock.Lock()
	defer l.lock.Unlock()

	if severity > l.threshold {
		return nil
	}

	// If fluentd is/was running on the host, the framework never gives
	// up sending logs to fluentd.
	if l.fluentd != nil {
		for {
			buf, err := msgpackfmt(l, t, severity, msg, fields)
			if err != nil {
				return err
			}
			err = l.fluentd.SendMessage(buf)
			if err == nil {
				break
			}
			l.fluentd.Close()
			for {
				os.Stderr.WriteString("reconnecting to fluentd...\n")
				conn, _ := newFluentConn()
				if conn != nil {
					l.fluentd = conn
					break
				}
				time.Sleep(time.Second)
			}
		}
	}

	if l.output != nil {
		buf, err := logfmt(l, t, severity, msg, fields)
		if err != nil {
			return err
		}
		_, err = l.output.Write(buf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "logger output causes an error: %v", err)
		}
		return err
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
