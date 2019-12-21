package log

// https://github.com/go-kit/kit/issues/164
// As much as I love the general go-kit effort of creating a generic library for everyone to use,
// they really don't seem to care about making developers' life easier.
// The whole reason why everyone liked their logging framework is that it was simple, but as soon
// as everyone started to take care about error handling and linting their code it started to make
// a lot less sense.

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
