package keywords

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
)

type From struct {
}

const (
	extractVariableRegex string = `\$\{([a-zA-Z.+-_]*)\}`
)

var (
	// The compiled version of the regex created at init() is cached here so it
	// only needs to be created once.
	regexVariableName *regexp.Regexp
)

func init() {
	regexVariableName = regexp.MustCompile(extractVariableRegex)
}

type FromTokenArgs struct {
	Prefix string
	Suffix string
	Name   string
}

type FromToken struct {
	Keyword  string
	Platform string
	Image    string
	Tag      string
	Digest   string
	Alias    string
	AliasKw  string
	Comment  string
	Args     map[string]*FromTokenArgs
}

func (ft FromToken) getImageRef() string {
	// We can now reconstruct the image, we first start with the tag
	imageWithTagTokens := []string{ft.Image}
	if ft.Tag != "" {
		imageWithTagTokens = append(imageWithTagTokens, ft.Tag)
	}
	imageWithTag := strings.Join(imageWithTagTokens, ":")
	// We now try to add the digest
	imageWithDigestTokens := []string{imageWithTag}
	if ft.Digest != "" {
		imageWithDigestTokens = append(imageWithDigestTokens, ft.Digest)
	}
	return strings.Join(imageWithDigestTokens, "@")
}

func parseArg(token string) (*FromTokenArgs, error) {
	founds := regexVariableName.FindAllStringSubmatch(token, -1)

	numFounds := len(founds)

	if numFounds == 0 {
		return nil, nil
	}

	if numFounds > 1 {
		err := errors.New("too many arguments")
		return nil, err
	}

	args := FromTokenArgs{}
	for _, found := range founds {
		startingIndex := strings.Index(token, "${"+found[1]+"}")
		endingIndex := startingIndex + len("${"+found[1]+"}")

		args.Name = found[1]

		if startingIndex > 0 {
			args.Prefix = token[:startingIndex]
		}

		if len(token) > endingIndex {
			args.Suffix = token[endingIndex:]
		}
	}

	return &args, nil
}

func (f From) GetTokens(originalLine string) (Tokens, error) {
	return f.getTokens(originalLine)
}
func (f From) getTokens(originalLine string) (FromToken, error) {
	tokens := FromToken{}
	parsedLine := strings.Fields(originalLine)
	args := make(map[string]*FromTokenArgs)
	lineLength := len(parsedLine)
	if len(parsedLine) < 2 {
		// Empty or malformed line
		return tokens, fmt.Errorf("got an empty or malformed line")
	}
	tokens.Keyword = parsedLine[0]
	if strings.ToLower(tokens.Keyword) != "from" {
		return tokens, fmt.Errorf("not a FROM line: %q", tokens.Keyword)
	}
	currentTokenIndex := 1

	// We may have a platform token
	if strings.HasPrefix(parsedLine[currentTokenIndex], "--platform=") {
		tokens.Platform = strings.TrimPrefix(parsedLine[currentTokenIndex], "--platform=")
		// Platform may have an arg in it
		arg, err := parseArg(tokens.Platform)
		if arg != nil && err == nil {
			args["platform"] = arg
		}
		currentTokenIndex += 1
	}

	// The next token is always the image, a from line without image is not valid
	if currentTokenIndex >= lineLength {
		return tokens, fmt.Errorf("no image in line")
	}

	imageName, imageTag, imageDigest, err := dockerimage.ParseOCIReferenceInfo(parsedLine[currentTokenIndex])
	if err != nil {
		return tokens, fmt.Errorf("could not parse image %q: %s", parsedLine[currentTokenIndex], err)
	}
	tokens.Image = imageName
	tokens.Tag = imageTag
	tokens.Digest = strings.TrimPrefix(imageDigest, "@")
	// Let' s extract args from image, tag or digest
	if tokens.Image != "" {
		arg, _ := parseArg(tokens.Image)
		if arg != nil {
			args["image"] = arg
		}
	}
	if tokens.Tag != "" {
		arg, _ := parseArg(tokens.Tag)
		if arg != nil {
			args["tag"] = arg
		}
	}
	if tokens.Digest != "" {
		arg, _ := parseArg(tokens.Digest)
		if arg != nil {
			args["digest"] = arg
		}
	}
	if len(args) > 0 {
		tokens.Args = args
	}
	currentTokenIndex += 1
	// We may have reach the end of the from directive
	if currentTokenIndex >= lineLength {
		return tokens, nil
	}
	currentToken := parsedLine[currentTokenIndex]
	if strings.ToLower(currentToken) == "as" {
		// We have an alias
		tokens.AliasKw = currentToken
		currentTokenIndex += 1
		if currentTokenIndex >= lineLength {
			// Invalid alias, missing alias
			return tokens, fmt.Errorf("malformed FROM line, AS keyword but no value for it")
		}
		tokens.Alias = parsedLine[currentTokenIndex]
		currentTokenIndex += 1
	}
	if currentTokenIndex >= lineLength {
		// We reach the end of a non commented line
		return tokens, nil
	}
	// Whatever remains should be a comment and start with #
	if !strings.HasPrefix(parsedLine[currentTokenIndex], "#") {
		return tokens, fmt.Errorf("remaining token in line that should be a comment but doesn't start with \"#\"")
	}
	tokens.Comment = strings.Join(parsedLine[currentTokenIndex:], " ")

	return tokens, nil

}

func (f From) ReplaceLine(source, originalLine, matcher string) string {
	if !f.IsLineMatching(originalLine, matcher) {
		return originalLine
	}
	tokens, err := f.getTokens(originalLine)

	if err != nil {
		// Could not process as a from line, not touching
		return originalLine
	}

	// Let's construct the line back
	lineTokens := []string{
		tokens.Keyword,
	}
	// Add platform if present
	if tokens.Platform != "" {
		lineTokens = append(lineTokens, fmt.Sprintf("--platform=%s", tokens.Platform))
	}
	// Reconstruct the image
	// And image can be in the form <name>[:<tag>][@<digest>]
	// Source can be in the form [<tag>][@<digest>]
	tag := source
	digest := ""
	splitSource := strings.Split(source, "@")
	if len(splitSource) == 2 {
		// We have a digest
		tag = splitSource[0]
		digest = splitSource[1]
	}
	tokens.Tag = tag
	tokens.Digest = digest
	lineTokens = append(lineTokens, tokens.getImageRef())
	// Add alias if present
	if tokens.AliasKw != "" {
		lineTokens = append(lineTokens, tokens.AliasKw, tokens.Alias)
	}
	// Add comment if present
	if tokens.Comment != "" {
		lineTokens = append(lineTokens, tokens.Comment)
	}

	return strings.Join(lineTokens, " ")

}

func (f From) IsLineMatching(originalLine, matcher string) bool {
	tokens, err := f.getTokens(originalLine)
	if err != nil {
		return false
	}
	return tokens.Image == matcher
}

func (f From) GetStageName(originalLine string) (string, error) {
	tokens, err := f.getTokens(originalLine)
	if err != nil {
		return "", err
	}
	return tokens.Alias, nil
}

func (f From) GetValue(originalLine, matcher string) (string, error) {
	tokens, err := f.getTokens(originalLine)
	if err != nil {
		return "", err
	}
	if !f.IsLineMatching(originalLine, matcher) {
		return "", fmt.Errorf("FROM line not matching")
	}
	return tokens.getImageRef(), nil
}
