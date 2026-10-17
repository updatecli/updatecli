// Package log is the only logging entry point for Updatecli.
//
// Every log line goes to the terminal and to the active captures (see Capture),
// so that console output can be attached to resource and pipeline reports.
package log

import (
	"context"
	"fmt"
	"log/slog"
)

// Level is a logging level.
type Level = slog.Level

const (
	ErrorLevel = slog.LevelError
	WarnLevel  = slog.LevelWarn
	InfoLevel  = slog.LevelInfo
	DebugLevel = slog.LevelDebug
)

var (
	// level is the minimum level of the messages to log, info by default
	level  slog.LevelVar
	logger = slog.New(handler{})
)

// handler writes slog records to the terminal and to the active captures.
type handler struct{}

func (handler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= level.Level()
}

func (handler) Handle(_ context.Context, r slog.Record) error {
	l := line{level: r.Level, msg: r.Message}

	r.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case kindKey:
			l.kind = kind(a.Value.Int64())
		case categoryKey:
			l.category = a.Value.String()
		case idKey:
			l.id = a.Value.String()
		case dryRunKey:
			l.dryRun = a.Value.Bool()
		case statusKey:
			l.status = Status(a.Value.Int64())
		case durationKey:
			l.duration = a.Value.Duration()
		case positionKey:
			l.position = int(a.Value.Int64())
		case totalKey:
			l.total = int(a.Value.Int64())
		}
		return true
	})

	return emit(l)
}

// Other attributes and groups aren't rendered, the console output is text.
func (h handler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h handler) WithGroup(string) slog.Handler      { return h }

// SetLevel sets the minimum level of the messages to log.
func SetLevel(l Level) {
	level.Set(l)
	setLogrusLevel(l)
}

// GetLevel returns the minimum level of the messages to log.
func GetLevel() Level {
	return level.Level()
}

// IsDebug reports whether debug messages are logged.
func IsDebug() bool {
	return enabled(DebugLevel)
}

func enabled(l Level) bool {
	return logger.Enabled(context.Background(), l)
}

func write(l Level, msg string) {
	logger.Log(context.Background(), l, msg)
}

// logf formats the message only when it's going to be logged.
func logf(l Level, format string, args ...any) {
	if enabled(l) {
		write(l, fmt.Sprintf(format, args...))
	}
}

func logs(l Level, args ...any) {
	if enabled(l) {
		write(l, fmt.Sprint(args...))
	}
}

// logln behaves like fmt.Sprintln, without the trailing newline.
func logln(l Level, args ...any) {
	if enabled(l) {
		msg := fmt.Sprintln(args...)
		write(l, msg[:len(msg)-1])
	}
}

// Debug logs a message at the debug level.
func Debug(args ...any) { logs(DebugLevel, args...) }

// Debugf logs a formatted message at the debug level.
func Debugf(format string, args ...any) { logf(DebugLevel, format, args...) }

// Debugln logs a message at the debug level.
func Debugln(args ...any) { logln(DebugLevel, args...) }

// Info logs a message at the info level.
func Info(args ...any) { logs(InfoLevel, args...) }

// Infof logs a formatted message at the info level.
func Infof(format string, args ...any) { logf(InfoLevel, format, args...) }

// Infoln logs a message at the info level.
func Infoln(args ...any) { logln(InfoLevel, args...) }

// Printf logs a formatted message at the info level.
func Printf(format string, args ...any) { logf(InfoLevel, format, args...) }

// Warning logs a message at the warning level.
func Warning(args ...any) { logs(WarnLevel, args...) }

// Warningf logs a formatted message at the warning level.
func Warningf(format string, args ...any) { logf(WarnLevel, format, args...) }

// Warningln logs a message at the warning level.
func Warningln(args ...any) { logln(WarnLevel, args...) }

// Warnf logs a formatted message at the warning level.
func Warnf(format string, args ...any) { logf(WarnLevel, format, args...) }

// Error logs a message at the error level.
func Error(args ...any) { logs(ErrorLevel, args...) }

// Errorf logs a formatted message at the error level.
func Errorf(format string, args ...any) { logf(ErrorLevel, format, args...) }

// Errorln logs a message at the error level.
func Errorln(args ...any) { logln(ErrorLevel, args...) }
