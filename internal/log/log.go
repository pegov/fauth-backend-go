package log

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

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

type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Criticalf(format string, args ...any)
}

type SimpleLogger struct {
	logger  *log.Logger
	Level   Level
	Colored bool
}

func NewSimpleLogger(level Level, colored bool) *SimpleLogger {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
	return &SimpleLogger{
		logger:  logger,
		Level:   level,
		Colored: colored,
	}
}

func (l *SimpleLogger) log(level Level, ansiCode string, levelName string, padding int, format string, args ...any) {
	if l.Level <= level {
		b := make([]byte, 0, 128)
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
	l.log(LevelDebug, ansiGrey, "DEBUG", 0, format, args...)
}

func (l *SimpleLogger) Infof(format string, args ...any) {
	l.log(LevelInfo, ansiGreen, "INFO", 1, format, args...)
}

func (l *SimpleLogger) Warnf(format string, args ...any) {
	l.log(LevelWarn, ansiYellow, "WARN", 1, format, args...)
}

func (l *SimpleLogger) Errorf(format string, args ...any) {
	l.log(LevelError, ansiRed, "ERROR", 0, format, args...)
}

func (l *SimpleLogger) Criticalf(format string, args ...any) {
	l.log(LevelCritical, ansiOrange, "CRIT", 1, format, args...)
	os.Exit(1)
}
