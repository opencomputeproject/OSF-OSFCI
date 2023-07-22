/*Package base logger notes: same as zap.go*/
package base

import "errors"

// Zlog A global variable so that log functions can be directly accessed
var Zlog Logger

// Fields Type to pass when we want to call WithFields for structured logging
type Fields map[string]interface{}

const (
	//Debug for verbose logging
	Debug = "debug"
	//Info for info logging
	Info = "info"
	//Warn for warning logs
	Warn = "warn"
	//Error for error logs
	Error = "error"
	//Fatal is for logging fatal messages. The system shutsdown after logging the message.
	Fatal = "fatal"
)

const (
	//InstanceZapLogger for instance
	InstanceZapLogger int = iota
)

var (
	errInvalidLoggerInstance = errors.New("Invalid logger instance")
)

// Logger interface
type Logger interface {
	//Debugf level interface
	Debugf(format string, args ...interface{})
	//Infof level interface
	Infof(format string, args ...interface{})
	//Warnf level interface
	Warnf(format string, args ...interface{})
	//Errorf level interface
	Errorf(format string, args ...interface{})
	//Fatalf level interface
	Fatalf(format string, args ...interface{})
	//Panic level interface
	Panicf(format string, args ...interface{})
}

// Configuration stores the config for the logger
type Configuration struct {
	EnableConsole     bool
	ConsoleJSONFormat bool
	ConsoleLevel      string
	EnableFile        bool
	FileJSONFormat    bool
	FileLevel         string
	FileLocation      string
}

// NewLogger returns an instance of logger
func NewLogger(config Configuration) error {
	logger, err := newZapLogger(config)
	if err != nil {
		return err
	}
	Zlog = logger
	return nil
}

// Debugf interface impl
func Debugf(format string, args ...interface{}) {
	Zlog.Debugf(format, args...)
}

// Infof interface impl
func Infof(format string, args ...interface{}) {
	Zlog.Infof(format, args...)
}

// Warnf interface impl
func Warnf(format string, args ...interface{}) {
	Zlog.Warnf(format, args...)
}

// Errorf interface impl
func Errorf(format string, args ...interface{}) {
	Zlog.Errorf(format, args...)
}

// Fatalf interface impl
func Fatalf(format string, args ...interface{}) {
	Zlog.Fatalf(format, args...)
}

// Panicf interface impl
func Panicf(format string, args ...interface{}) {
	Zlog.Panicf(format, args...)
}
