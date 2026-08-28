package config

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// excerptContextLines is the number of lines displayed before and after the line
// a templating error points at.
const excerptContextLines = 3

// manifestFragment locates one file's contribution inside the concatenated
// content handed to the templating engine.
type manifestFragment struct {
	file string
	// firstLine is the 1-based line number of this fragment's first line within
	// the concatenation.
	firstLine int
	lineCount int
}

// lastLine returns the 1-based line number of this fragment's last line within
// the concatenation.
func (f manifestFragment) lastLine() int {
	return f.firstLine + f.lineCount - 1
}

// appendFragment records where content lands inside the concatenation described
// by fragments, and returns both the updated fragment list and the content to
// append.
//
// The content is normalized to end with a newline: without it the last line of a
// fragment glues onto the first line of the next one, which both corrupts the
// resulting manifest and makes line numbers impossible to attribute.
func appendFragment(fragments []manifestFragment, file string, content []byte) ([]manifestFragment, []byte) {
	if len(content) > 0 && !bytes.HasSuffix(content, []byte("\n")) {
		content = append(content, '\n')
	}

	firstLine := 1
	if len(fragments) > 0 {
		previous := fragments[len(fragments)-1]
		firstLine = previous.firstLine + previous.lineCount
	}

	return append(fragments, manifestFragment{
		file:      file,
		firstLine: firstLine,
		lineCount: bytes.Count(content, []byte("\n")),
	}), content
}

// templateLocation is a templating error resolved back to the file it came from.
type templateLocation struct {
	fragment manifestFragment
	// line is the 1-based line number within the fragment's own file.
	line int
	// concatLine is the 1-based line number within the concatenated content.
	concatLine int
}

// templateError is a templating error whose location has been resolved back to
// the file and line the user actually wrote.
type templateError struct {
	msg string
	err error
}

func (e *templateError) Error() string { return e.msg }
func (e *templateError) Unwrap() error { return e.err }

// resolveFragment returns the fragment owning the given line of the
// concatenation, or nil when no fragment covers it.
func resolveFragment(fragments []manifestFragment, concatLine int) *templateLocation {
	for _, fragment := range fragments {
		if concatLine >= fragment.firstLine && concatLine <= fragment.lastLine() {
			return &templateLocation{
				fragment:   fragment,
				line:       concatLine - fragment.firstLine + 1,
				concatLine: concatLine,
			}
		}
	}

	return nil
}

// locateTemplateError rewrites the "template: <name>:<line>[:<column>]" prefix
// that text/template puts on parse and execution errors so that it points at the
// file the line actually came from, rather than at an offset into the
// concatenation of every partial and the manifest.
//
// The error is returned untouched, with a nil location, when it carries no such
// prefix or when no fragment covers the reported line.
func locateTemplateError(err error, name string, fragments []manifestFragment) (*templateLocation, error) {
	if err == nil {
		return nil, nil
	}

	// The name is quoted so that a Windows path, or any other name containing a
	// colon, cannot swallow the line number.
	prefix := regexp.MustCompile(`^template: ` + regexp.QuoteMeta(name) + `:(\d+)(?::(\d+))?: `)

	message := err.Error()
	match := prefix.FindStringSubmatch(message)
	if match == nil {
		return nil, err
	}

	concatLine, convErr := strconv.Atoi(match[1])
	if convErr != nil {
		return nil, err
	}

	location := resolveFragment(fragments, concatLine)
	if location == nil {
		return nil, err
	}

	position := fmt.Sprintf("%s:%d", location.fragment.file, location.line)
	if match[2] != "" {
		position += ":" + match[2]
	}

	return location, &templateError{
		msg: fmt.Sprintf("%s: %s", position, strings.TrimPrefix(message, match[0])),
		err: err,
	}
}

// renderExcerpt returns the lines of content surrounding location, with the
// offending line marked. Line numbers are those of the file the line came from,
// so that they match what the user sees in their editor, and the excerpt never
// spills into a neighboring fragment.
func renderExcerpt(content []byte, location *templateLocation, contextLines int) string {
	if location == nil {
		return ""
	}

	lines := strings.Split(string(content), "\n")
	// Content ending with a newline yields a trailing empty element which is not
	// a line of its own.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	first := max(location.concatLine-contextLines, location.fragment.firstLine)
	last := min(location.concatLine+contextLines, location.fragment.lastLine())
	last = min(last, len(lines))

	if first > last {
		return ""
	}

	// Width of the widest line number displayed, so that the gutter stays aligned.
	width := len(strconv.Itoa(location.line + last - location.concatLine))

	excerpt := strings.Builder{}
	for i := first; i <= last; i++ {
		marker := "  "
		if i == location.concatLine {
			marker = "> "
		}
		fmt.Fprintf(&excerpt, "%s%*d | %s\n", marker, width, location.line+i-location.concatLine, lines[i-1])
	}

	return excerpt.String()
}
