package log

import (
	"bytes"
	"github.com/fatih/color"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	// colorStatus returns a new function that returns status-colorized (cyan) strings for the
	// given arguments with fmt.Sprint().
	colorStatus = color.New(color.FgCyan).SprintFunc()

	// colorWarn returns a new function that returns status-colorized (yellow) strings for the
	// given arguments with fmt.Sprint().
	colorWarn = color.New(color.FgYellow).SprintFunc()

	// colorError returns a new function that returns error-colorized (red) strings for the
	// given arguments with fmt.Sprint().
	colorError = color.New(color.FgRed).SprintFunc()
)

// TextFormat lets use a custom text format.
type TextFormat struct {
	ShowInfoLevel   bool
	ShowTimestamp   bool
	TimestampFormat string
}

// NewTextFormat creates the default text formatter.
func NewTextFormat() *TextFormat {
	return &TextFormat{
		ShowInfoLevel:   false,
		ShowTimestamp:   false,
		TimestampFormat: "2006-01-02 15:04:05",
	}
}

// Format formats the log statement.
func (f *TextFormat) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer

	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	level := strings.ToUpper(entry.Level.String())
	switch level {
	case "INFO":
		if f.ShowInfoLevel {
			b.WriteString(colorStatus(level))
			b.WriteString(": ")
		}
	case "WARNING":
		b.WriteString(colorWarn(level))
		b.WriteString(": ")
	case "DEBUG":
		b.WriteString(colorStatus(level))
		b.WriteString(": ")
	default:
		b.WriteString(colorError(level))
		b.WriteString(": ")
	}
	if f.ShowTimestamp {
		b.WriteString(entry.Time.Format(f.TimestampFormat))
		b.WriteString(" - ")
	}

	b.WriteString(entry.Message)

	if !strings.HasSuffix(entry.Message, "\n") {
		b.WriteByte('\n')
	}
	return b.Bytes(), nil
}
