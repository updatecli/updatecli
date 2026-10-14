// Package log is the only logging entry point for Updatecli.
//
// Every log line goes to the terminal and to the active captures (see Capture),
// so that console output can be attached to resource and pipeline reports.
package log

import (
	"github.com/sirupsen/logrus"
)

// Level is a logging level.
type Level = logrus.Level

const (
	ErrorLevel = logrus.ErrorLevel
	WarnLevel  = logrus.WarnLevel
	InfoLevel  = logrus.InfoLevel
	DebugLevel = logrus.DebugLevel
)

func init() {
	// Dependencies such as helm or sops log through the global logrus logger,
	// routing it through our output keeps their messages in captures.
	logrus.SetOutput(fanout{})
	logrus.SetFormatter(NewTextFormat())
}

// SetLevel sets the minimum level of the messages to log.
func SetLevel(level Level) {
	logrus.SetLevel(level)
}

// GetLevel returns the minimum level of the messages to log.
func GetLevel() Level {
	return logrus.GetLevel()
}

// IsDebug reports whether debug messages are logged.
func IsDebug() bool {
	return logrus.IsLevelEnabled(logrus.DebugLevel)
}

// Debug logs a message at the debug level.
func Debug(args ...any) { logrus.Debug(args...) }

// Debugf logs a formatted message at the debug level.
func Debugf(format string, args ...any) { logrus.Debugf(format, args...) }

// Debugln logs a message at the debug level.
func Debugln(args ...any) { logrus.Debugln(args...) }

// Info logs a message at the info level.
func Info(args ...any) { logrus.Info(args...) }

// Infof logs a formatted message at the info level.
func Infof(format string, args ...any) { logrus.Infof(format, args...) }

// Infoln logs a message at the info level.
func Infoln(args ...any) { logrus.Infoln(args...) }

// Printf logs a formatted message at the info level.
func Printf(format string, args ...any) { logrus.Printf(format, args...) }

// Warning logs a message at the warning level.
func Warning(args ...any) { logrus.Warning(args...) }

// Warningf logs a formatted message at the warning level.
func Warningf(format string, args ...any) { logrus.Warningf(format, args...) }

// Warningln logs a message at the warning level.
func Warningln(args ...any) { logrus.Warningln(args...) }

// Warnf logs a formatted message at the warning level.
func Warnf(format string, args ...any) { logrus.Warnf(format, args...) }

// Error logs a message at the error level.
func Error(args ...any) { logrus.Error(args...) }

// Errorf logs a formatted message at the error level.
func Errorf(format string, args ...any) { logrus.Errorf(format, args...) }

// Errorln logs a message at the error level.
func Errorln(args ...any) { logrus.Errorln(args...) }
