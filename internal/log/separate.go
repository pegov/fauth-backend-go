package log

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

type SeparateFilesLogger struct {
	accessLogPath string
	errorLogPath  string
	accessFile    *os.File
	errorFile     *os.File
	accessLogger  *log.Logger
	errorLogger   *log.Logger
	Level         Level
	Colored       bool
	pool          BufPool
}

func openLogFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0660)
}

func NewSeparateLogger(accessLogPath, errorLogPath string, colored bool) (*SeparateFilesLogger, error) {
	accessFile, err := openLogFile(accessLogPath)
	if err != nil {
		return nil, err
	}
	errorFile, err := openLogFile(errorLogPath)
	if err != nil {
		return nil, err
	}

	accessLogger := log.New(accessFile, "", log.LstdFlags|log.Lmicroseconds)
	errorLogger := log.New(errorFile, "", log.LstdFlags|log.Lmicroseconds)

	return &SeparateFilesLogger{
		accessLogPath: accessLogPath,
		errorLogPath:  errorLogPath,
		accessFile:    accessFile,
		errorFile:     errorFile,
		accessLogger:  accessLogger,
		errorLogger:   errorLogger,
		Colored:       colored,
		pool:          NewBufPool(1024, 128),
	}, nil
}

func (l *SeparateFilesLogger) log(level Level, ansiCode string, levelName string, padding int, format string, args ...any) {
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

		l.accessLogger.Print(buf)

		if level >= LevelError {
			l.errorLogger.Print(buf)
		}
	}
}

func (l *SeparateFilesLogger) Tracef(format string, args ...any) {
	l.log(LevelDebug, ansiBlue, LevelHeaderTrace, 0, format, args...)
}

func (l *SeparateFilesLogger) Debugf(format string, args ...any) {
	l.log(LevelDebug, ansiGrey, LevelHeaderDebug, 0, format, args...)
}

func (l *SeparateFilesLogger) Infof(format string, args ...any) {
	l.log(LevelInfo, ansiGreen, LevelHeaderInfo, 1, format, args...)
}

func (l *SeparateFilesLogger) Warnf(format string, args ...any) {
	l.log(LevelWarn, ansiYellow, LevelHeaderWarn, 1, format, args...)
}

func (l *SeparateFilesLogger) Errorf(format string, args ...any) {
	l.log(LevelError, ansiRed, LevelHeaderError, 0, format, args...)
}

func (l *SeparateFilesLogger) Criticalf(format string, args ...any) {
	l.log(LevelCritical, ansiMagenta, LevelHeaderCritical, 1, format, args...)
	os.Exit(1)
}

func (l *SeparateFilesLogger) Restart() error {
	if err := l.accessFile.Close(); err != nil {
		return err
	}

	if err := l.errorFile.Close(); err != nil {
		return err
	}

	accessFile, err := openLogFile(l.accessLogPath)
	if err != nil {
		return err
	}
	errorFile, err := openLogFile(l.errorLogPath)
	if err != nil {
		return err
	}
	l.accessFile = accessFile
	l.errorFile = errorFile

	accessLogger := log.New(accessFile, "", log.LstdFlags|log.Lmicroseconds)
	errorLogger := log.New(errorFile, "", log.LstdFlags|log.Lmicroseconds)
	l.accessLogger = accessLogger
	l.errorLogger = errorLogger

	return nil
}
