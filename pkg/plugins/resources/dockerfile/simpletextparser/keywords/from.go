package keywords

import (
	"fmt"
	"strings"
)

type From struct {
	Alias bool
}
type fromToken struct {
	keyword  string
	platform string
	image    string
	tag      string
	digest   string
	alias    string
	aliasKw  string
	comment  string
}

func (ft fromToken) getImageRef() string {
	// We can now reconstruct the image, we first start with the tag
	imageWithTagTokens := []string{ft.image}
	if ft.tag != "" {
		imageWithTagTokens = append(imageWithTagTokens, ft.tag)
	}
	imageWithTag := strings.Join(imageWithTagTokens, ":")
	// We now try to add the digest
	imageWithDigestTokens := []string{imageWithTag}
	if ft.digest != "" {
		imageWithDigestTokens = append(imageWithDigestTokens, ft.digest)
	}
	return strings.Join(imageWithDigestTokens, "@")
}

func (f From) parseTokens(originalLine string) (fromToken, error) {
	tokens := fromToken{}
	parsedLine := strings.Fields(originalLine)
	lineLength := len(parsedLine)
	if len(parsedLine) < 2 {
		// Empty or malformed line
		return tokens, fmt.Errorf("Got an empty or malformed line")
	}
	tokens.keyword = parsedLine[0]
	if strings.ToLower(tokens.keyword) != "from" {
		return tokens, fmt.Errorf("Not a FROM line: %q", tokens.keyword)
	}
	currentTokenIndex := 1

	// We may have a platform token
	if strings.HasPrefix(parsedLine[currentTokenIndex], "--platform") {
		tokens.platform = parsedLine[currentTokenIndex]
		currentTokenIndex += 1
	}

	// The next token is always the image, a from line without image is not valid
	if currentTokenIndex >= lineLength {
		return tokens, fmt.Errorf("No image in line")
	}
	rawImage := parsedLine[currentTokenIndex]
	// Test for digest
	imageInfo := strings.Split(rawImage, "@")
	if len(imageInfo) == 2 {
		// image as a digest
		tokens.digest = imageInfo[1]
		rawImage = imageInfo[0]
	}
	// Test for tag
	imageInfo = strings.Split(rawImage, ":")
	if len(imageInfo) == 2 {
		// image as a tag
		tokens.tag = imageInfo[1]
		rawImage = imageInfo[0]
	}
	// Whatever remains is the raw image
	tokens.image = rawImage
	currentTokenIndex += 1
	// We may have reach the end of the from directive
	if currentTokenIndex >= lineLength {
		return tokens, nil
	}
	currentToken := parsedLine[currentTokenIndex]
	if strings.ToLower(currentToken) == "as" {
		// We have an alias
		tokens.aliasKw = currentToken
		currentTokenIndex += 1
		if currentTokenIndex >= lineLength {
			// Invalid alias, missing alias
			return tokens, fmt.Errorf("Malformed FROM line, AS keyword but no value for it")
		}
		tokens.alias = parsedLine[currentTokenIndex]
		currentTokenIndex += 1
	}
	if currentTokenIndex >= lineLength {
		// We reach the end of a non commented line
		return tokens, nil
	}
	// Whatever remains should be a comment and start with #
	if !strings.HasPrefix(parsedLine[currentTokenIndex], "#") {
		return tokens, fmt.Errorf("Remaining token in line that should be a comment but doesn't start with \"#\"")
	}
	tokens.comment = strings.Join(parsedLine[currentTokenIndex:], " ")

	return tokens, nil

}

func (f From) ReplaceLine(source, originalLine, matcher string) string {
	if !f.IsLineMatching(originalLine, matcher) {
		return originalLine
	}
	tokens, err := f.parseTokens(originalLine)
	if err != nil {
		// Could not process as a from line, not touching
		return originalLine
	}
	// Let's construct the line back
	lineTokens := []string{
		tokens.keyword,
	}
	// Add platform if present
	if tokens.platform != "" {
		lineTokens = append(lineTokens, tokens.platform)
	}
	// Reconstruct the image
	// And image can be in the form <name>[:<tag>][@<digest>]
	// Source can be in the form [<tag>][@<digest>]
	if !f.Alias {
		tag := source
		digest := ""
		splitSource := strings.Split(source, "@")
		if len(splitSource) == 2 {
			// We have a digest
			tag = splitSource[0]
			digest = splitSource[1]
		}
		tokens.tag = tag
		tokens.digest = digest
	}
	lineTokens = append(lineTokens, tokens.getImageRef())
	// Add alias if present
	if tokens.aliasKw != "" {
		alias := tokens.alias
		if f.Alias {
			alias = source
		}
		lineTokens = append(lineTokens, tokens.aliasKw, alias)
	}
	// Add comment if present
	if tokens.comment != "" {
		lineTokens = append(lineTokens, tokens.comment)
	}

	return strings.Join(lineTokens, " ")

}

func (f From) IsLineMatching(originalLine, matcher string) bool {
	tokens, err := f.parseTokens(originalLine)
	if err != nil {
		return false
	}

	toMatch := tokens.image
	if f.Alias {
		toMatch = tokens.alias
	}
	if toMatch == "" {
		return false
	}
	return strings.HasPrefix(toMatch, matcher)
}

func (f From) GetStageName(originalLine string) (string, error) {
	tokens, err := f.parseTokens(originalLine)
	if err != nil {
		return "", err
	}
	return tokens.alias, nil
}

func (f From) GetValue(originalLine, matcher string) (string, error) {
	tokens, err := f.parseTokens(originalLine)
	if err != nil {
		return "", err
	}
	if !f.IsLineMatching(originalLine, matcher) {
		return "", fmt.Errorf("FROM line not matching")
	}
	if f.Alias {
		return tokens.alias, nil
	}
	return tokens.getImageRef(), nil
}
