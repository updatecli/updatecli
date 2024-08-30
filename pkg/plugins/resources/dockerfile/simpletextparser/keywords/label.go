package keywords

import (
	"fmt"
	"strings"
)

type Label struct{}

func (a Label) ReplaceLine(source, originalLine, matcher string) string {

	if !a.IsLineMatching(originalLine, matcher) {
		return originalLine
	}

	// With an ARG instruction, we only need to use the 2nd "word"
	parsedLine := strings.Fields(originalLine)

	parsedLine[1] = matcher + "=" + source
	return strings.Join(parsedLine, " ")
}

func (a Label) IsLineMatching(originalLine, matcher string) bool {
	parsedLine := strings.Fields(originalLine)
	if len(parsedLine) == 0 {
		// Empty line
		return false
	}
	found := strings.ToLower(parsedLine[0]) == "label" && strings.HasPrefix(parsedLine[1], matcher)

	return found
}

func (a Label) GetValue(originalLine, matcher string) (string, error) {
	if a.IsLineMatching(originalLine, matcher) {
		// With an ARG instruction, we just need the rest of the 2nd "word"
		parsedLine := strings.Fields(originalLine)
		splittedArgValue := strings.Split(parsedLine[1], "=")
		if len(splittedArgValue) < 2 {
			// ARG without value
			return "", nil
		} else {
			return splittedArgValue[1], nil
		}
	}
	return "", fmt.Errorf("Value not found in line")
}
