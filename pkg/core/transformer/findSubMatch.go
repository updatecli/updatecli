package transformer

import (
	"fmt"
	"regexp"

	"github.com/sirupsen/logrus"
)

// FindSubMatch is a struct used to feed regexp.findSubMatch
type FindSubMatch struct {
	// Pattern defines regular expression to use for retrieving a submatch
	Pattern                string
	DeprecatedCaptureIndex int `yaml:"captureIndex"`
	// CaptureIndex defines which substring occurrence to retrieve. Note also that a value of `0` for `captureIndex` returns all submatches, and individual submatch indexes start at `1`.
	CaptureIndex int
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

	// Log if there can't be a submatch corresponding to the captureIndex
	if len(found) <= findSubMatch.CaptureIndex {
		logrus.Debugf("No capture found at position %v after applying regex %q to %q, full result with CaptureIndex 0 would be %v", findSubMatch.CaptureIndex, findSubMatch.Pattern, output, found)
		return "", nil
	}

	// Output the submatch corresponding to the captureIndex
	return found[findSubMatch.CaptureIndex], nil
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
