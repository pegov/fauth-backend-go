package log

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

type SimpleLogger struct {
	logger  *log.Logger
	Level   Level
	Colored bool
	pool    BufPool
}

func NewSimpleLogger(level Level, colored bool) *SimpleLogger {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
	return &SimpleLogger{
		logger:  logger,
		Level:   level,
		Colored: colored,
		pool:    NewBufPool(1024, 128),
	}
}

func (l *SimpleLogger) log(level Level, ansiCode string, levelName string, padding int, format string, args ...any) {
	if l.Level <= level {
		b := l.pool.Next()
		buf := bytes.NewBuffer(b)
		buf.WriteString(ansiCode)
		buf.WriteString("| ")
		for range padding {
			buf.WriteString(" ")
		}
		buf.WriteString(levelName)
		buf.WriteString(" | ")
		buf.WriteString(ansiReset)
		s := fmt.Sprintf(format, args...)
		buf.WriteString(s)
		l.logger.Print(buf)
	}
}

func (l *SimpleLogger) Debugf(format string, args ...any) {
	l.log(LevelDebug, ansiGrey, LevelHeaderDebug, 0, format, args...)
}

func (l *SimpleLogger) Infof(format string, args ...any) {
	l.log(LevelInfo, ansiGreen, LevelHeaderInfo, 1, format, args...)
}

func (l *SimpleLogger) Warnf(format string, args ...any) {
	l.log(LevelWarn, ansiYellow, LevelHeaderWarn, 1, format, args...)
}

func (l *SimpleLogger) Errorf(format string, args ...any) {
	l.log(LevelError, ansiRed, LevelHeaderError, 0, format, args...)
}

func (l *SimpleLogger) Criticalf(format string, args ...any) {
	l.log(LevelCritical, ansiOrange, LevelHeaderCritical, 1, format, args...)
	os.Exit(1)
}

func (l *SimpleLogger) Restart() error {
	return nil
}
