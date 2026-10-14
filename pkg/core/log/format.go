package log

import (
	"bytes"
	"strings"

	"github.com/fatih/color"
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

// render formats a log line: the level, except for info messages, then the message.
func render(level Level, msg string) []byte {
	var b bytes.Buffer

	switch {
	case level >= ErrorLevel:
		b.WriteString(colorError("ERROR"))
		b.WriteString(": ")
	case level >= WarnLevel:
		b.WriteString(colorWarn("WARNING"))
		b.WriteString(": ")
	case level >= InfoLevel:
	default:
		b.WriteString(colorStatus("DEBUG"))
		b.WriteString(": ")
	}

	b.WriteString(msg)

	if !strings.HasSuffix(msg, "\n") {
		b.WriteByte('\n')
	}
	return b.Bytes()
}
