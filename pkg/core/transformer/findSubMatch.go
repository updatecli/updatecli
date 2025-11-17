package transformer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

// FindSubMatch is a struct used to feed regexp.findSubMatch
type FindSubMatch struct {
	// Pattern defines regular expression to use for retrieving a submatch
	Pattern                string `yaml:",omitempty" jsonschema:"required"`
	DeprecatedCaptureIndex int    `yaml:"captureIndex,omitempty" jsonschema:"-"`
	// CaptureIndex defines which substring occurrence to retrieve. Note also that a value of `0` for `captureIndex` returns all submatches, and individual submatch indexes start at `1`.
	CaptureIndex int
	// Uses the match group(s) to generate the output using \0, \1, \2, etc
	CapturePattern string `yaml:",omitempty"`
}

func (f *FindSubMatch) Apply(input string) (string, error) {

	output := input
	findSubMatch := f

	if len(findSubMatch.Pattern) == 0 {
		return "", fmt.Errorf("no regex provided")
	}

	// Check if the regular expression can be compiled
	re, err := regexp.Compile(findSubMatch.Pattern)
	if err != nil {
		return "", err
	}

	found := re.FindStringSubmatch(output)

	// Log if no match is found
	if len(found) == 0 {
		logrus.Debugf("No result found after applying regex %q to %q", findSubMatch.Pattern, output)
		return "", nil
	}

	if findSubMatch.CapturePattern != "" {
		pattern := findSubMatch.CapturePattern
		for i, v := range found {
			replace := fmt.Sprintf("\\%d", i)
			pattern = strings.Replace(pattern, replace, v, -1)
		}

		return pattern, nil
	} else {
		// Log if there can't be a submatch corresponding to the captureIndex
		if len(found) <= findSubMatch.CaptureIndex {
			logrus.Debugf("No capture found at position %v after applying regex %q to %q, full result with CaptureIndex 0 would be %v", findSubMatch.CaptureIndex, findSubMatch.Pattern, output, found)
			return "", nil
		}

		// Output the submatch corresponding to the captureIndex
		return found[findSubMatch.CaptureIndex], nil
	}
}

func (f *FindSubMatch) Validate() error {
	if f.DeprecatedCaptureIndex != 0 {
		logrus.Warningln("captureIndex is deprecated in favor of captureindex")

		switch f.CaptureIndex {
		case 0:
			f.CaptureIndex = f.DeprecatedCaptureIndex
		default:
			logrus.Warningf("Both captureIndex and captureindex are defined, ignoring the first one")
		}

		f.DeprecatedCaptureIndex = 0
	}

	return nil
}
