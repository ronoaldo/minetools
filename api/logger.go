package api

import (
	"log"
	"os"
)

type logLevel int

const (
	// Debug log level
	Debug logLevel = 0
	// Info log level
	Info logLevel = 1
	// Warning log level
	Warning logLevel = 5
	// NoLogs disable logs completely
	NoLogs logLevel = 9
)

var (
	// LogLevel controls the global logging for the API calls.
	LogLevel logLevel = Warning

	// logger is the internal package logger.
	logger = log.New(os.Stderr, "[minetools.api] ", log.Ldate|log.Ltime)
)

// Debugf logs the provided message if debug level is enabled
func Debugf(m string, args ...interface{}) {
	if LogLevel <= Debug {
		logger.Printf("DEBUG: "+m, args...)
	}
}

// Infof logs the provided message if debug level is enabled
func Infof(m string, args ...interface{}) {
	if LogLevel <= Info {
		logger.Printf("INFO: "+m, args...)
	}
}

// Warningf logs the provided message if debug level is enabled
func Warningf(m string, args ...interface{}) {
	if LogLevel <= Warning {
		logger.Printf("WARNING: "+m, args...)
	}
}
