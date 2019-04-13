package core

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

const (
	LevelDebug = iota
	LevelInfo
	LevelWarning
	LevelError
	LevelFatal
)

var levels = map[int]string{
	LevelDebug:   "DEBUG",
	LevelInfo:    "INFO",
	LevelWarning: "WARN",
	LevelError:   "ERROR",
	LevelFatal:   "FATAL",
}

type Logger struct {
	out   io.Writer
	Level int
}

func NewLogger(level int, out io.Writer) *Logger {
	return &Logger{Level: level, out: out}
}

func (l *Logger) SetLevel(level string) error {
	level = strings.ToUpper(level)
	for k, v := range levels {
		if v == level {
			l.Level = k
			return nil
		}
	}
	return fmt.Errorf("unknown log level %s.", level)
}

func (l *Logger) Log(level int, format string, v ...interface{}) {
	if level >= l.Level {
		msg := fmt.Sprintf(format, v...)
		t := time.Now().Format("2006-01-02T15:04:05.000000Z07:00")
		fmt.Fprintf(l.out, "%-5s [%s] %s\n", levels[level], t, msg)
	}
}

func (l *Logger) Debug(format string, v ...interface{}) {
	l.Log(LevelDebug, format, v...)
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.Log(LevelInfo, format, v...)
}

func (l *Logger) Warning(format string, v ...interface{}) {
	l.Log(LevelWarning, format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.Log(LevelError, format, v...)
}

func (l *Logger) Fatal(format string, v ...interface{}) {
	l.Log(LevelFatal, format, v...)
	os.Exit(1)
}
