package log

const (
	ansiReset  = "\033[0m"
	ansiGrey   = "\033[2m"
	ansiRed    = "\033[91m"
	ansiGreen  = "\033[92m"
	ansiYellow = "\033[93m"
	ansiOrange = "\033[95m"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelCritical
	LevelNone
)

const (
	LevelHeaderDebug    = "DEBUG"
	LevelHeaderInfo     = "INFO"
	LevelHeaderWarn     = "WARN"
	LevelHeaderError    = "ERROR"
	LevelHeaderCritical = "CRIT"
)

type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Criticalf(format string, args ...any)
	Restart() error
}
