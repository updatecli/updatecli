package keywords

import (
	"fmt"
	"strings"
)

type SimpleKeyword struct {
	Keyword string
}

type SimpleTokens struct {
	Keyword string
	Name    string
	Value   string
	Comment string
}

func (t SimpleTokens) String() string {
	var builder strings.Builder

	builder.WriteString(t.Keyword + " " + t.Name)
	if t.Value == "" {
		return builder.String()
	}
	builder.WriteString("=" + t.Value)
	if t.Comment != "" {
		builder.WriteString(t.Comment)
	}
	return builder.String()
}

func (s SimpleKeyword) IsLineMatching(originalLine, matcher string) bool {
	tokens, err := s.getTokens(originalLine)
	if err != nil {
		return false
	}
	return strings.HasPrefix(tokens.Name, matcher)
}

func (s SimpleKeyword) ReplaceLine(source, originalLine, matcher string) string {
	if !s.IsLineMatching(originalLine, matcher) {
		return originalLine
	}
	tokens, err := s.getTokens(originalLine)
	if err != nil {
		return originalLine
	}
	tokens.Value = source
	return tokens.String()
}

func (s SimpleKeyword) GetValue(originalLine, matcher string) (string, error) {
	if s.IsLineMatching(originalLine, matcher) {
		tokens, err := s.getTokens(originalLine)
		if err == nil {
			return tokens.Value, nil
		}
	}
	return "", fmt.Errorf("value not found in line")
}

func (s SimpleKeyword) getTokens(originalLine string) (SimpleTokens, error) {
	tokens := SimpleTokens{}
	parsedLine := strings.Fields(originalLine)
	lineLength := len(parsedLine)
	if lineLength < 2 {
		// Empty or malformed line
		return tokens, fmt.Errorf("got an empty or malformed line")
	}
	if !strings.EqualFold(parsedLine[0], s.Keyword) {
		return tokens, fmt.Errorf("expected %q keyword: %q", s.Keyword, parsedLine[0])
	}
	// Save token case
	tokens.Keyword = parsedLine[0]
	commentIndex := 2
	// Legacy Docker allows for env without equals
	if !strings.Contains(parsedLine[1], "=") && len(parsedLine) >= 3 {
		tokens.Name = parsedLine[1]
		tokens.Value = parsedLine[2]
		commentIndex = 3
	} else {
		splitArgValue := strings.Split(parsedLine[1], "=")
		tokens.Name = splitArgValue[0]
		if len(splitArgValue) == 2 {
			tokens.Value = splitArgValue[1]
		}
	}
	// Ensure we don't store the quotes
	tokens.Value = strings.Trim(tokens.Value, "\"")

	if lineLength > commentIndex {
		// We may have a comment
		if !strings.HasPrefix(parsedLine[2], "#") {
			return tokens, fmt.Errorf("remaining token in line that should be a comment but doesn't start with \"#\"")
		}
		tokens.Comment = strings.Join(parsedLine[commentIndex:], " ")
	}
	return tokens, nil
}

func (s SimpleKeyword) GetTokens(originalLine string) (Tokens, error) {
	return s.getTokens(originalLine)
}
