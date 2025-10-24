package simpletextparser

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile/simpletextparser/keywords"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile/types"
)

type SimpleTextDockerfileParser struct {
	Keyword      string `yaml:"keyword,omitempty"`
	Matcher      string `yaml:"matcher,omitempty"`
	Stage        string `yaml:"stage,omitempty"`
	KeywordLogic keywords.Logic
}

func (s SimpleTextDockerfileParser) FindInstruction(dockerfileContent []byte, stage string) bool {
	if found, originalLine := s.hasMatch(dockerfileContent, stage); found {
		logrus.Infof("%s Line %q found, matching the keyword %q and the matcher %q.", result.SUCCESS, originalLine, s.Keyword, s.Matcher)
		return true
	}

	logrus.Infof("%s No line found matching the keyword %q and the matcher %q.", result.FAILURE, s.Keyword, s.Matcher)
	return false
}

func (s SimpleTextDockerfileParser) hasMatch(dockerfileContent []byte, stage string) (bool, string) {
	scanner := bufio.NewScanner(bytes.NewReader(dockerfileContent))

	// For each line of the dockerfile
	// Make sure we handle the stage
	currentStageName := ""
	stageScanner := keywords.From{}
	for scanner.Scan() {
		originalLine := scanner.Text()
		if stageName, err := stageScanner.GetStageName(originalLine); err == nil {
			currentStageName = stageName
		}
		if stage != "" && currentStageName != stage {
			continue
		}
		if s.KeywordLogic.IsLineMatching(originalLine, s.Matcher) {
			// Immediately returns if there is match: the result of a condition is the same with one or multiple matches
			return true, originalLine
		}
	}

	return false, ""
}

func (s SimpleTextDockerfileParser) ReplaceInstructions(dockerfileContent []byte, sourceValue, stage string) ([]byte, types.ChangedLines, error) {

	changedLines := make(types.ChangedLines)

	if found, _ := s.hasMatch(dockerfileContent, stage); !found {
		return dockerfileContent, changedLines, fmt.Errorf("%s No line found matching the keyword %q and the matcher %q", result.FAILURE, s.Keyword, s.Matcher)
	}

	var newDockerfile bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(dockerfileContent))
	linePosition := 0

	// Make sure we handle the stage
	currentStageName := ""
	stageScanner := keywords.From{}
	// For each line of the dockerfile
	for scanner.Scan() {
		originalLine := scanner.Text()
		newLine := originalLine
		linePosition += 1

		if stageName, err := stageScanner.GetStageName(originalLine); err == nil {
			currentStageName = stageName
		}

		if ((stage != "" && stage == currentStageName) || stage == "") && s.KeywordLogic.IsLineMatching(originalLine, s.Matcher) {
			// if strings.HasPrefix(strings.ToLower(originalLine), strings.ToLower(d.spec.Instruction.Keyword)) {
			newLine = s.KeywordLogic.ReplaceLine(sourceValue, originalLine, s.Matcher)

			if newLine != originalLine {
				logrus.Infof("%s The line #%d, matched by the keyword %q and the matcher %q, is changed from %q to %q.",
					result.ATTENTION,
					linePosition,
					s.Keyword,
					s.Matcher,
					originalLine,
					newLine,
				)
				changedLines[linePosition] = types.LineDiff{Original: originalLine, New: newLine}
			} else {
				logrus.Infof("%s The line #%d, matched by the keyword %q and the matcher %q, is correctly set to %q.",
					result.SUCCESS,
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

func (s SimpleTextDockerfileParser) GetInstructionTokens(dockerfileContent []byte) []keywords.Tokens {
	var instructions []keywords.Tokens
	scanner := bufio.NewScanner(bytes.NewReader(dockerfileContent))
	for scanner.Scan() {
		line := scanner.Text()
		if s.KeywordLogic.IsLineMatching(line, s.Matcher) {
			token, err := s.KeywordLogic.GetTokens(line)
			if err != nil {
				continue
			}
			instructions = append(instructions, token)
		}
	}
	return instructions
}

// GetImages will return all images in a dockerfile
func (s SimpleTextDockerfileParser) GetImages(dockerfileContent []byte) []keywords.FromToken {
	var images []keywords.FromToken
	stages := map[string]bool{}
	scanner := bufio.NewScanner(bytes.NewReader(dockerfileContent))
	fromScanner := keywords.From{}
	for scanner.Scan() {
		originalLine := scanner.Text()
		tokens, err := fromScanner.GetTokens(originalLine)
		if err != nil {
			continue
		}
		fromTokens := tokens.(keywords.FromToken)
		if _, ok := stages[fromTokens.Image]; ok {
			// We are reusing a stage
			continue
		}
		if fromTokens.Alias != "" {
			stages[fromTokens.Alias] = true
		}
		images = append(images, fromTokens)
	}
	return images
}

func (s SimpleTextDockerfileParser) GetInstruction(dockerfileContent []byte, stage string) string {
	type stageValue struct {
		StageName string
		Value     string
	}
	var stagesInstruction []stageValue
	scanner := bufio.NewScanner(bytes.NewReader(dockerfileContent))
	stageCounter := -1
	currentStageName := ""
	stageScanner := keywords.From{}
	for scanner.Scan() {
		originalLine := scanner.Text()
		if stageName, err := stageScanner.GetStageName(originalLine); err == nil {
			currentStageName = stageName
			stageCounter += 1
		}
		if stage != "" && stage != currentStageName {
			continue
		}
		if s.KeywordLogic.IsLineMatching(originalLine, s.Matcher) {
			stageName := currentStageName
			if stageName == "" {
				// Use stage counter
				stageName = fmt.Sprintf("%d", stageCounter)
			}
			if value, err := s.KeywordLogic.GetValue(originalLine, s.Matcher); err == nil {
				stageValue := stageValue{
					StageName: stageName,
					Value:     value,
				}
				if len(stagesInstruction) > 0 && stagesInstruction[len(stagesInstruction)-1].StageName == stageName {
					// Value can be overwritten
					stagesInstruction = stagesInstruction[:len(stagesInstruction)-1]
				}
				stagesInstruction = append(stagesInstruction, stageValue)
			}
		}

	}
	var instruction string
	if len(stagesInstruction) == 0 {
		// Found nothing
		return ""
	}
	if stage == "" {
		// No Stage selected, return info for last one
		return stagesInstruction[len(stagesInstruction)-1].Value
	}
	// We have a stage name, let's try to find it
	for _, currentStage := range stagesInstruction {
		if currentStage.StageName == stage {
			instruction = currentStage.Value
		}
	}
	return instruction
}

func NewSimpleTextDockerfileParser(input map[string]string) (SimpleTextDockerfileParser, error) {

	var newParser SimpleTextDockerfileParser
	var err error

	keyword, ok := input["keyword"]
	if !ok {
		return SimpleTextDockerfileParser{}, fmt.Errorf("cannot parse instruction: No key 'keyword' found in the instruction %v", input)
	}
	newParser.Keyword = keyword
	err = newParser.setKeywordLogic()
	if err != nil {
		return SimpleTextDockerfileParser{}, err
	}
	delete(input, "keyword")

	matcher, ok := input["matcher"]
	if !ok {
		return SimpleTextDockerfileParser{}, fmt.Errorf("cannot parse instruction: No key 'matcher' found in the instruction %v", input)
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
	"label":       keywords.SimpleKeyword{Keyword: "label"},
	"maintainer":  nil,
	"expose":      nil,
	"add":         nil,
	"copy":        nil,
	"entrypoint":  nil,
	"volume":      nil,
	"user":        nil,
	"workdir":     nil,
	"arg":         keywords.SimpleKeyword{Keyword: "arg"},
	"env":         keywords.SimpleKeyword{Keyword: "env"},
	"onbuild":     nil,
	"stopsignal":  nil,
	"healthcheck": nil,
	"shell":       nil,
}

func (s *SimpleTextDockerfileParser) setKeywordLogic() error {
	keywordLogic, found := supportedKeywordsInitializers[strings.ToLower(s.Keyword)]
	if !found {
		return fmt.Errorf("%s Unknown keyword %q provided for Dockerfile's instruction", result.FAILURE, s.Keyword)
	}

	if keywordLogic == nil {
		return fmt.Errorf("%s Provided keyword %q not supported (yet). Feel free to open an issue explaining your use-case to help adding the implementation", result.FAILURE, s.Keyword)
	}

	s.KeywordLogic = keywordLogic

	return nil
}
