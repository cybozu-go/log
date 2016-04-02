package log

// LevelName returns the name for the defined threshold.
// An empty string is returned for undefined thresholds.
func LevelName(level int) string {
	switch level {
	case LvCritical:
		return "critical"
	case LvError:
		return "error"
	case LvWarn:
		return "warning"
	case LvInfo:
		return "info"
	case LvDebug:
		return "debug"
	}
	return ""
}
