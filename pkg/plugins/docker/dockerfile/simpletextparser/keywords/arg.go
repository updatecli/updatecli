package keywords

import (
	"strings"
)

type Arg struct{}

func (a Arg) ReplaceLine(source, originalLine, matcher string) string {

	if !a.IsLineMatching(originalLine, matcher) {
		return originalLine
	}

	// With an ARG instruction, we only need to use the 2nd "word"
	parsedLine := strings.Fields(originalLine)

	parsedLine[1] = matcher + "=" + source
	return strings.Join(parsedLine, " ")
}

func (a Arg) IsLineMatching(originalLine, matcher string) bool {
	parsedLine := strings.Fields(originalLine)
	if len(parsedLine) == 0 {
		// Empty line
		return false
	}
	found := strings.ToLower(parsedLine[0]) == "arg" && strings.HasPrefix(parsedLine[1], matcher)

	return found
}
