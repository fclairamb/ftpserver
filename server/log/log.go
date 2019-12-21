// Package log provides a way to use the go-kit/log package without having to deal
// with their very-opiniated/crazy choice of returning an error all the time: https://github.com/go-kit/kit/issues/164
package log

import (
	"fmt"

	gklog "github.com/go-kit/kit/log"
	gklevel "github.com/go-kit/kit/log/level"
)

// Logger interface
type Logger interface {
	// Log(keyvals ...interface{})
	Debug(keyvals ...interface{})
	Info(keyvals ...interface{})
	Warn(keyvals ...interface{})
	Error(keyvals ...interface{})
	With(keyvals ...interface{}) Logger
}

type gKLogger struct {
	logger gklog.Logger
}

var (
	// DefaultCaller adds a "caller" property
	DefaultCaller = gklog.Caller(4)
	// DefaultTimestampUTC adds a "ts" property
	DefaultTimestampUTC = gklog.DefaultTimestampUTC
)

func (logger *gKLogger) checkError(err error) {
	if err != nil {
		fmt.Println("Logging faced this error: ", err)
	}
}

// Debug logs key-values at debug level
func (logger *gKLogger) Debug(keyvals ...interface{}) {
	logger.checkError(gklevel.Debug(logger.logger).Log(keyvals...))
}

// Info logs key-values at info level
func (logger *gKLogger) Info(keyvals ...interface{}) {
	logger.checkError(gklevel.Info(logger.logger).Log(keyvals...))
}

// Warn logs key-values at warn level
func (logger *gKLogger) Warn(keyvals ...interface{}) {
	logger.checkError(gklevel.Warn(logger.logger).Log(keyvals...))
}

// Error logs key-values at error level
func (logger *gKLogger) Error(keyvals ...interface{}) {
	logger.checkError(gklevel.Error(logger.logger).Log(keyvals...))
}

// With adds key-values
func (logger *gKLogger) With(keyvals ...interface{}) Logger {
	return NewGKLogger(gklog.With(logger.logger, keyvals...))
}

// NewGKLogger creates a logger based on go-kit logs
func NewGKLogger(logger gklog.Logger) Logger {
	return &gKLogger{
		logger: logger,
	}
}

// NewNopGKLogger instantiates go-kit logger
func NewNopGKLogger() Logger {
	return NewGKLogger(gklog.NewNopLogger())
}
