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

	// With an LABEL instruction, we only need to use the 2nd "word"
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
		// With an LABEL instruction, we just need the rest of the 2nd "word"
		parsedLine := strings.Fields(originalLine)
		splitArgValue := strings.Split(parsedLine[1], "=")
		if len(splitArgValue) < 2 {
			// LABEL without value
			return "", nil
		} else {
			return splitArgValue[1], nil
		}
	}
	return "", fmt.Errorf("Value not found in line")
}
