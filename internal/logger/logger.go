package logger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

// InitLogger initializes the logger with the specified level
func InitLogger(level string) {
	Logger = logrus.New()
	
	// Set log level
	logLevel, err := logrus.ParseLevel(strings.ToLower(level))
	if err != nil {
		logLevel = logrus.InfoLevel
		Logger.Warnf("Invalid log level '%s', using 'info'", level)
	}
	Logger.SetLevel(logLevel)

	// Set formatter
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05",
		ForceColors:     true,
		PadLevelText:    true,
	})

	// Set output
	Logger.SetOutput(os.Stdout)
}

// Info logs an info message
func Info(args ...interface{}) {
	if Logger != nil {
		Logger.Info(args...)
	}
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Infof(format, args...)
	}
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	if Logger != nil {
		Logger.Warn(args...)
	}
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Warnf(format, args...)
	}
}

// Error logs an error message
func Error(args ...interface{}) {
	if Logger != nil {
		Logger.Error(args...)
	}
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Errorf(format, args...)
	}
}

// Debug logs a debug message
func Debug(args ...interface{}) {
	if Logger != nil {
		Logger.Debug(args...)
	}
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Debugf(format, args...)
	}
}

// WithField creates a log entry with a field
func WithField(key string, value interface{}) *logrus.Entry {
	if Logger != nil {
		return Logger.WithField(key, value)
	}
	return logrus.NewEntry(logrus.New())
}

// WithFields creates a log entry with multiple fields
func WithFields(fields logrus.Fields) *logrus.Entry {
	if Logger != nil {
		return Logger.WithFields(fields)
	}
	return logrus.NewEntry(logrus.New())
}

// WithError creates a log entry with an error
func WithError(err error) *logrus.Entry {
	if Logger != nil {
		return Logger.WithError(err)
	}
	return logrus.NewEntry(logrus.New())
}