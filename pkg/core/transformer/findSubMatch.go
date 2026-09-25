package transformer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

// FindSubMatch defines a regular expression extracting a capture group from the value.
type FindSubMatch struct {
	// "pattern" defines the regular expression applied to the value.
	//
	// remark:
	//   * when nothing matches, the value becomes empty.
	//
	// example:
	//   * pattern: 'v(\d+)\.(\d+)'
	//
	Pattern                string `yaml:",omitempty" jsonschema:"required"`
	DeprecatedCaptureIndex int    `yaml:"captureIndex,omitempty" jsonschema:"-"`
	// "captureindex" defines the capture group to return.
	//
	// default:
	//   0
	//
	// remark:
	//   * 0 returns the whole match, and capture groups start at 1.
	//   * an index without a matching capture group returns an empty value.
	//   * ignored when "capturepattern" is set.
	//
	// example:
	//   * captureindex: 1
	//
	CaptureIndex int
	// "capturepattern" defines a template building the value from the capture groups.
	//
	// remark:
	//   * \0 is replaced by the whole match, \1 by the first capture group, \2 by the second, and so on.
	//   * it takes precedence over "captureindex".
	//
	// example:
	//   * capturepattern: \1.\2
	//
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
			pattern = strings.ReplaceAll(pattern, replace, v)
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
