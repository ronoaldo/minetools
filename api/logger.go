package api

import (
	"log"
	"os"
)

type LogLevel int

const (
	// Debug log level
	Debug LogLevel = 0
	// Info log level
	Info LogLevel = 1
	// Warning log level
	Warning LogLevel = 5
	// Error log level
	Error LogLevel = 6
	// NoLogs disable logs completely
	NoLogs LogLevel = 9
)

// Logger is a thin stdlib log wrapper.
type Logger struct {
	Level  LogLevel
	Logger *log.Logger
}

// NewLogger creates a new Logger to help with debug/info/warning messages.
func NewLogger(prefix string) *Logger {
	logger := Logger{
		Level:  Warning,
		Logger: log.New(os.Stderr, prefix, log.Ldate|log.Ltime),
	}
	return &logger
}

// Debugf logs the provided message if debug level is enabled.
func (l *Logger) Debugf(m string, args ...interface{}) {
	if l.Level <= Debug {
		l.Logger.Printf("DEBUG: "+m, args...)
	}
}

// Infof logs the provided message if debug level is enabled.
func (l *Logger) Infof(m string, args ...interface{}) {
	if l.Level <= Info {
		l.Logger.Printf("INFO: "+m, args...)
	}
}

// Warningf logs the provided message if debug level is enabled.
func (l *Logger) Warningf(m string, args ...interface{}) {
	if l.Level <= Warning {
		l.Logger.Printf("WARNING: "+m, args...)
	}
}

// Errorf logs the provided message if debug level is enabled.
func (l *Logger) Errorf(m string, args ...interface{}) {
	if l.Level <= Error {
		l.Logger.Printf("ERROR: "+m, args...)
	}
}

var (
	// logger is the internal package logger.
	logger = NewLogger("[minetools.api] ")
)

// SetLogLevel sets the Logging level for the default API logs.
func SetLogLevel(level LogLevel) {
	logger.Level = level
}

// Debugf logs the provided message if debug level is enabled.
func Debugf(m string, args ...interface{}) {
	logger.Debugf(m, args...)
}

// Infof logs the provided message if debug level is enabled.
func Infof(m string, args ...interface{}) {
	logger.Infof(m, args...)
}

// Warningf logs the provided message if debug level is enabled.
func Warningf(m string, args ...interface{}) {
	logger.Warningf(m, args...)
}

// Errorf logs the provided message if debug level is enabled.
func Errorf(m string, args ...interface{}) {
	logger.Errorf(m, args...)
}
