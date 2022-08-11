package logger

import (
	"errors"
	"log"
)

type Logger interface {
	Debug(format string, v ...any)
	Info(format string, v ...any)
	Warn(format string, v ...any)
	Error(format string, v ...any)
}

var _ Logger = (*MyLogger)(nil)

type Level int

const (
	Debug Level = iota
	Info
	Warn
	Error
)

type MyLogger struct {
	level Level
	l     *log.Logger
}

func New(level Level) *MyLogger {
	return &MyLogger{
		level: level,
		l:     log.Default(),
	}
}

func (m *MyLogger) Debug(format string, v ...any) {
	if m.level > Debug {
		return
	}
	m.l.Printf("[DEBUG] "+format, v...)
}

func (m *MyLogger) Info(format string, v ...any) {
	if m.level > Info {
		return
	}
	m.l.Printf("[INFO] "+format, v...)
}

func (m *MyLogger) Warn(format string, v ...any) {
	if m.level > Warn {
		return
	}
	m.l.Printf("[WARN] "+format, v...)
}

func (m *MyLogger) Error(format string, v ...any) {
	if m.level > Error {
		return
	}
	m.l.Printf("[ERROR] "+format, v...)
}

func LevelFromString(s string) (Level, error) {
	switch s {
	case "debug":
		return Debug, nil
	default:
		return Debug, errors.New("unknown logger level")

	}
}
