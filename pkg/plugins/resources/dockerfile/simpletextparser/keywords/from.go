package keywords

import (
	"fmt"
	"strings"
)

type From struct{}

func (f From) ReplaceLine(source, originalLine, matcher string) string {

	if !f.IsLineMatching(originalLine, matcher) {
		return originalLine
	}

	// With a FROM instruction, Only the first word following "FROM" has to be processed
	parsedLine := strings.Fields(originalLine)
	parsedValue := strings.Split(parsedLine[1], ":")

	if len(parsedValue) == 1 {
		// Case with no tags specified: append the new tag to the current value
		parsedLine[1] = fmt.Sprintf("%s:%s", parsedLine[1], source)
	} else {
		// Case with a tag already specified: replace only the tag's value
		parsedLine[1] = fmt.Sprintf("%s:%s", parsedValue[0], source)
	}

	return strings.Join(parsedLine, " ")

}

func (f From) IsLineMatching(originalLine, matcher string) bool {
	parsedLine := strings.Fields(originalLine)
	if len(parsedLine) == 0 {
		// Empty line
		return false
	}
	found := strings.ToLower(parsedLine[0]) == "from" && strings.HasPrefix(parsedLine[1], matcher)

	return found
}
