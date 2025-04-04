package mobyparser

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile/types"
)

type MobyParser struct {
	// Specifies the Dockerfile instruction such as ENV
	Instruction string `yaml:",omitempty"`
	// Specifies the Dockerfile value for the instruction
	Value string `yaml:",omitempty"`
}

var (
	instructionRegex = regexp.MustCompile(`^.*\[([[:digit:]]*)\]\[[[:digit:]]*\]$`)
	positionRegex    = regexp.MustCompile(`(.*)*\[[[:digit:]]*\]\[[[:digit:]]*\]$`)
	elementRegex     = regexp.MustCompile(`^.*\[[[:digit:]]*\]\[([[:digit:]])*\]$`)
	positionKeyRegex = regexp.MustCompile(`^(.*)\[[[:digit:]]*\]\[[[:digit:]]*\]$`)
)

func (m MobyParser) FindInstruction(dockerfileContent []byte, stage string) bool {

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
		logrus.Infof("%s Instruction %q wasn't found", result.FAILURE, m.Instruction)
		return false
	}

	if val == m.Value {
		logrus.Infof("%s Instruction %q is correctly set to %q",
			result.SUCCESS,
			m.Instruction,
			m.Value)
		return true
	}

	logrus.Infof("%s Instruction %q found but incorrectly set to %q instead of %q",
		result.FAILURE,
		m.Instruction,
		val,
		m.Value)
	return true
}

func (m MobyParser) ReplaceInstructions(dockerfileContent []byte, sourceValue, stage string) ([]byte, types.ChangedLines, error) {

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

	if !valueFound {
		return dockerfileContent, changed, fmt.Errorf("%s cannot find instruction %q", result.FAILURE, m.Instruction)
	}

	if oldVersion == m.Value {
		logrus.Infof("%s Instruction %q already set to %q, nothing else need to be done",
			result.SUCCESS,
			m.Instruction,
			m.Value)
		return dockerfileContent, changed, nil
	}

	logrus.Infof("%s Instruction %q, was updated from %q to %q",
		result.ATTENTION,
		m.Instruction,
		oldVersion,
		m.Value,
	)

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

func (m MobyParser) GetInstruction(dockerfileContent []byte, stage string) string {
	// Not implemented
	logrus.Warningf("Get Instruction is not yet supported for the MobyParser, if needed switch to the simpletextparser")
	return ""
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

		if strings.EqualFold(n.Value, instruction) && i == instructionPosition {
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
		} else if strings.EqualFold(n.Value, instruction) {
			i++
		}

	}
	return false, "", nil
}
