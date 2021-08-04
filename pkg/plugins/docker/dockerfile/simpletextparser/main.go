package simpletextparser

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/plugins/docker/dockerfile/simpletextparser/keywords"
	"github.com/updatecli/updatecli/pkg/plugins/docker/dockerfile/types"
	"github.com/sirupsen/logrus"
)

type SimpleTextDockerfileParser struct {
	Keyword      string `yaml:"keyword"`
	Matcher      string `yaml:"matcher"`
	KeywordLogic keywords.Logic
}

func (s SimpleTextDockerfileParser) FindInstruction(dockerfileContent []byte) bool {
	if found, originalLine := s.hasMatch(dockerfileContent); found {
		logrus.Infof("\u2714 Line %q found, matching the keyword %q and the matcher %q.", originalLine, s.Keyword, s.Matcher)
		return true
	}

	logrus.Infof("\u2717 No line found matching the keyword %q and the matcher %q.", s.Keyword, s.Matcher)
	return false
}

func (s SimpleTextDockerfileParser) hasMatch(dockerfileContent []byte) (bool, string) {
	scanner := bufio.NewScanner(bytes.NewReader(dockerfileContent))

	// For each line of the dockerfile
	for scanner.Scan() {
		originalLine := scanner.Text()

		if s.KeywordLogic.IsLineMatching(originalLine, s.Matcher) {
			// Immediately returns if there is match: the result of a condition is the same with one or multiple matches
			return true, originalLine
		}
	}

	return false, ""
}

func (s SimpleTextDockerfileParser) ReplaceInstructions(dockerfileContent []byte, sourceValue string) ([]byte, types.ChangedLines, error) {

	changedLines := make(types.ChangedLines)

	if found, _ := s.hasMatch(dockerfileContent); !found {
		return dockerfileContent, changedLines, fmt.Errorf("\u2717 No line found matching the keyword %q and the matcher %q.", s.Keyword, s.Matcher)
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
				logrus.Infof("\u2714 The line #%d, matched by the keyword %q and the matcher %q, is changed from %q to %q.",
					linePosition,
					s.Keyword,
					s.Matcher,
					originalLine,
					newLine,
				)
				changedLines[linePosition] = types.LineDiff{Original: originalLine, New: newLine}
			} else {
				logrus.Infof("\u2714 The line #%d, matched by the keyword %q and the matcher %q, is correctly set to %q.",
					linePosition,
					s.Keyword,
					s.Matcher,
					originalLine,
				)
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

var supportedKeywordsInitializers = map[string]keywords.Logic{
	"from":        keywords.From{},
	"run":         nil,
	"cmd":         nil,
	"label":       nil,
	"maintainer":  nil,
	"expose":      nil,
	"env":         nil,
	"add":         nil,
	"copy":        nil,
	"entrypoint":  nil,
	"volume":      nil,
	"user":        nil,
	"workdir":     nil,
	"arg":         keywords.Arg{},
	"onbuild":     nil,
	"stopsignal":  nil,
	"healthcheck": nil,
	"shell":       nil,
}

func (s *SimpleTextDockerfileParser) setKeywordLogic() error {
	keywordLogic, found := supportedKeywordsInitializers[strings.ToLower(s.Keyword)]
	if !found {
		return fmt.Errorf("\u2717 Unknown keyword %q provided for Dockerfile's instruction.", s.Keyword)
	}

	if keywordLogic == nil {
		return fmt.Errorf("\u2717 Provided keyword %q not supported (yet). Feel free to open an issue explaining your use-case to help adding the implementation.", s.Keyword)
	}

	s.KeywordLogic = keywordLogic

	return nil
}
