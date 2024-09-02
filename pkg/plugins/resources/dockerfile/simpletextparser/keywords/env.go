package keywords

import (
	"fmt"
	"strings"
)

type Env struct{}

func (a Env) ReplaceLine(source, originalLine, matcher string) string {

	if !a.IsLineMatching(originalLine, matcher) {
		return originalLine
	}

	// With an ENV instruction, we only need to use the 2nd "word"
	parsedLine := strings.Fields(originalLine)

	// As per https://docs.docker.com/engine/reference/builder/#env
	// syntax without an equal sign is still supported
	// Let's check for presence of equal sign or not on the "parsed" key
	if !strings.Contains(parsedLine[1], "=") {
		// Legacy case: rewrite with new recommended syntax (by Docker)
		return fmt.Sprintf("ENV %s=%s", parsedLine[1], source)
	}

	parsedLine[1] = matcher + "=" + source
	return strings.Join(parsedLine, " ")
}

func (a Env) IsLineMatching(originalLine, matcher string) bool {
	parsedLine := strings.Fields(originalLine)
	if len(parsedLine) == 0 {
		// Empty line
		return false
	}
	found := strings.ToLower(parsedLine[0]) == "env" && strings.HasPrefix(parsedLine[1], matcher)

	return found
}

func (a Env) GetValue(originalLine, matcher string) (string, error) {
	if a.IsLineMatching(originalLine, matcher) {
		// With an ENV instruction, we just need the rest of the 2nd "word"
		parsedLine := strings.Fields(originalLine)
		argValue := parsedLine[1]
		// Legacy Docker allows for env without equals
		if !strings.Contains(argValue, "=") && len(parsedLine) >= 3 {
			argValue = fmt.Sprintf("%s=%s", argValue, parsedLine[2])
		}
		splitArgValue := strings.Split(argValue, "=")
		if len(splitArgValue) < 2 {
			// ENV without value
			return "", nil
		} else {
			return splitArgValue[1], nil
		}
	}
	return "", fmt.Errorf("Value not found in line")
}
