package simpletextparser

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/olblak/updateCli/pkg/plugins/docker/dockerfile/simpletextparser/keywords"
	"github.com/olblak/updateCli/pkg/plugins/docker/dockerfile/types"
	"github.com/sirupsen/logrus"
)

type SimpleTextDockerfileParser struct {
	Keyword      string `yaml:"keyword"`
	Matcher      string `yaml:"matcher"`
	KeywordLogic keywords.Logic
}

func (s SimpleTextDockerfileParser) FindInstruction(dockerfileContent []byte) bool {
	scanner := bufio.NewScanner(bytes.NewReader(dockerfileContent))

	// For each line of the dockerfile
	for scanner.Scan() {
		originalLine := scanner.Text()

		if s.KeywordLogic.IsLineMatching(originalLine, s.Matcher) {
			// Immediately return if there is match: the result of a condition is the same with one or multiple matches
			return true
		}
	}
	return false
}

func (s SimpleTextDockerfileParser) ReplaceInstructions(dockerfileContent []byte, sourceValue string) ([]byte, types.ChangedLines, error) {

	changedLines := make(types.ChangedLines)

	if !s.FindInstruction(dockerfileContent) {
		// TODO: Dockerfile path (put on parent caller?)
		return dockerfileContent, changedLines, fmt.Errorf(
			"\u2717 No instruction found for the keyword %q and the matcher %q from Dockerfile '%s'",
			s.Keyword,
			s.Matcher,
			"FOO",
		)
	}

	var newDockerfile bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(dockerfileContent))
	linePosition := 0

	// For each line of the dockerfile
	for scanner.Scan() {
		originalLine := scanner.Text()
		newLine := originalLine
		linePosition += 1

		if s.KeywordLogic.IsLineMatching(originalLine, s.Matcher) {
			// if strings.HasPrefix(strings.ToLower(originalLine), strings.ToLower(d.Instruction.Keyword)) {
			newLine = s.KeywordLogic.ReplaceLine(sourceValue, originalLine, s.Matcher)

			if newLine != originalLine {
				changedLines[linePosition] = types.LineDiff{Original: originalLine, New: newLine}
			} else {
				// TODO: Dockerfile path (put on parent caller?)
				logrus.Infof("\u2714 Instruction '%s' from Dockerfile '%s', is correctly set to '%s'", s.Keyword, "FOO", originalLine)
			}
		}

		// Write the line to the new content (error is always nil)
		_, _ = newDockerfile.WriteString(newLine + "\n")
	}
	return newDockerfile.Bytes(), changedLines, nil
}

func NewSimpleTextDockerfileParser(input map[string]string) (SimpleTextDockerfileParser, error) {

	var newParser SimpleTextDockerfileParser
	var err error

	keyword, ok := input["keyword"]
	if !ok {
		return SimpleTextDockerfileParser{}, fmt.Errorf("Cannot parse instruction: No key 'keyword' found in the instruction %v.", input)
	}
	newParser.Keyword = keyword
	err = newParser.setKeywordLogic()
	if err != nil {
		return SimpleTextDockerfileParser{}, err
	}
	delete(input, "keyword")

	matcher, ok := input["matcher"]
	if !ok {
		return SimpleTextDockerfileParser{}, fmt.Errorf("Cannot parse instruction: No key 'matcher' found in the instruction %v.", input)
	}
	newParser.Matcher = matcher
	delete(input, "matcher")

	if len(input) > 0 {
		logrus.Warnf("Unused directives found for instruction: %v.", input)
	}

	return newParser, nil
}

func (s *SimpleTextDockerfileParser) setKeywordLogic() error {
	switch strings.ToLower(s.Keyword) {
	case "from":
		s.KeywordLogic = keywords.From{}
	case "arg":
		s.KeywordLogic = keywords.Arg{}
	default:
		return fmt.Errorf("\u2717 Unknown keyword %q provided for Dockerfile's instruction.", s.Keyword)
	}

	return nil
}

func (s SimpleTextDockerfileParser) String() string {
	return fmt.Sprintf("%s %s", strings.ToUpper(s.Keyword), s.Matcher)
}
