package mobyparser

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/olblak/updateCli/pkg/plugins/docker/dockerfile/types"
	"github.com/sirupsen/logrus"
)

type MobyParser struct {
	Instruction string
	Value       string
}

var (
	instructionRegex = regexp.MustCompile(`^.*\[([[:digit:]]*)\]\[[[:digit:]]*\]$`)
	positionRegex    = regexp.MustCompile(`(.*)*\[[[:digit:]]*\]\[[[:digit:]]*\]$`)
	elementRegex     = regexp.MustCompile(`^.*\[[[:digit:]]*\]\[([[:digit:]])*\]$`)
	positionKeyRegex = regexp.MustCompile(`^(.*)\[[[:digit:]]*\]\[[[:digit:]]*\]$`)
)

func (m MobyParser) FindInstruction(dockerfileContent []byte) bool {

	data, err := parser.Parse(bytes.NewReader(dockerfileContent))
	if err != nil {
		logrus.Error(err.Error())
		return false
	}

	found, val, err := m.replace(data.AST)
	if err != nil {
		logrus.Error(err.Error())
		return false
	}

	if !found {
		logrus.Infof("\u2717 Instruction %q wasn't found", m.Instruction)
		return false
	}

	if val == m.Value {
		logrus.Infof("\u2714 Instruction %q is correctly set to %q",
			m.Instruction,
			m.Value)
		return true
	}

	logrus.Infof("\u2717 Instruction %q found but incorrectly set to %q instead of %q",
		m.Instruction,
		val,
		m.Value)
	return true
}

func (m MobyParser) ReplaceInstructions(dockerfileContent []byte, sourceValue string) ([]byte, types.ChangedLines, error) {

	changed := make(types.ChangedLines)

	data, err := parser.Parse(bytes.NewReader(dockerfileContent))
	if err != nil {
		return dockerfileContent, changed, err
	}

	m.Value = sourceValue

	valueFound, oldVersion, err := m.replace(data.AST)
	if err != nil {
		return dockerfileContent, changed, err
	}

	if valueFound {
		if oldVersion == m.Value {
			logrus.Infof("\u2714 Instruction %q already set to %q, nothing else need to be done",
				m.Instruction,
				m.Value)
			return dockerfileContent, changed, nil
		}

		logrus.Infof("\u2714 Instruction %q, was updated from %q to %q",
			m.Instruction,
			oldVersion,
			m.Value,
		)
	} else {
		return dockerfileContent, changed, fmt.Errorf("\u2717 cannot find instruction %q", m.Instruction)
	}

	newDockerfileContent := ""
	err = Marshal(data, &newDockerfileContent)
	if err != nil {
		return dockerfileContent, changed, err
	}

	originalDockerfileContentLines := strings.Split(string(dockerfileContent), "\n")
	newDockerfileContentLines := strings.Split(newDockerfileContent, "\n")
	for index, currentLine := range newDockerfileContentLines {
		// If the line is not a comment, then we expect it to be a change
		if !strings.HasPrefix(strings.TrimSpace(currentLine), "#") {
			// Check that the originalDockerfile has the original line or ignore the change
			if len(originalDockerfileContentLines) > index && currentLine != originalDockerfileContentLines[index] {
				changed[index] = types.LineDiff{Original: originalDockerfileContentLines[index], New: currentLine}
			}
		}
	}

	return []byte(newDockerfileContent), changed, nil
}

func (m MobyParser) String() string {
	return fmt.Sprintf("%s %s", m.Instruction, m.Value)
}

// isPositionKey checks if key use the array position like instruction[0][0].
// where first [0] references the instruction position
// and the  second [0] references the element inside instruction line
func isPositionKeys(key string) bool {
	matched := positionKeyRegex.MatchString(key)

	return matched
}

func getPositionKeys(k string) (
	key string,
	instructionPosition int,
	elementPosition int,
	err error) {

	if isPositionKeys(k) {
		keys := positionRegex.FindStringSubmatch(k)
		key = keys[1]

		positions := instructionRegex.FindStringSubmatch(k)

		instructionPosition, err = strconv.Atoi(positions[1])

		if err != nil {
			logrus.Errorf("err - %s", err)
			return "", -1, -1, err
		}

		positions = elementRegex.FindStringSubmatch(k)

		elementPosition, err = strconv.Atoi(positions[1])

		if err != nil {
			logrus.Errorf("err - %s", err)
			return "", -1, -1, err
		}

	} else {
		key = k
		instructionPosition = 0
		elementPosition = 0
		err = nil
	}

	return key, instructionPosition, elementPosition, err

}

// Search for both a Dockerfile instruction and its value.
// While the dockerfile instruction is not case sensitive, its value is
func (m MobyParser) replace(node *parser.Node) (bool, string, error) {
	instruction, instructionPosition, elementPosition, err := getPositionKeys(m.Instruction)
	if err != nil {
		return false, "", err
	}

	i := 0
	for _, n := range node.Children {

		if strings.ToUpper(n.Value) == strings.ToUpper(instruction) && i == instructionPosition {

			if n.Next != nil {
				j := 0
				for nod := n.Next; nod != nil && j <= elementPosition; nod = nod.Next {
					if elementPosition == j {
						val := nod.Value
						nod.Value = m.Value

						return true, val, nil
					}
					j++
				}

			}
		} else if strings.ToUpper(n.Value) == strings.ToUpper(instruction) {
			i++
		}

	}
	return false, "", nil
}
