package log

const (
	ansiReset   = "\033[0m"
	ansiGrey    = "\033[2m"
	ansiRed     = "\033[31m"
	ansiGreen   = "\033[32m"
	ansiYellow  = "\033[33m"
	ansiBlue    = "\033[34m"
	ansiMagenta = "\033[35m"
)

type Level int

const (
	LevelTrace Level = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelCritical
	LevelNone
)

const (
	LevelHeaderTrace    = "TRACE"
	LevelHeaderDebug    = "DEBUG"
	LevelHeaderInfo     = "INFO"
	LevelHeaderWarn     = "WARN"
	LevelHeaderError    = "ERROR"
	LevelHeaderCritical = "CRIT"
)

type Logger interface {
	Tracef(format string, args ...any)
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Criticalf(format string, args ...any)
	Restart() error
}
