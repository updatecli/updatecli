package file

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

// Source return a file content
func (f *File) Source(workingDir string) (string, error) {
	if len(f.spec.ReplacePattern) > 0 {
		validationError := fmt.Errorf("Validation error in source of type 'file': the attribute `spec.replacepattern` is only supported for targets.")
		logrus.Errorf(validationError.Error())
		return "", validationError
	}

	if err := f.Read(); err != nil {
		return "", err
	}

	result := f.CurrentContent
	// If a matchPattern is specified, then retrieve the string matched and returns the (eventually) multi-line string
	if len(f.spec.MatchPattern) > 0 {
		reg, err := regexp.Compile(f.spec.MatchPattern)
		if err != nil {
			logrus.Errorf("Validation error in source of type 'file': Unable to parse the regexp specified at f.spec.MatchPattern (%q)", f.spec.MatchPattern)
			return "", err
		}

		// Check if there is any match in the file
		if !reg.MatchString(f.CurrentContent) {
			return "", fmt.Errorf("No line matched in the file %q for the pattern %q", f.spec.File, f.spec.MatchPattern)
		}
		matchedStrings := reg.FindAllString(f.CurrentContent, -1)

		result = strings.Join(matchedStrings, "\n")
	}

	logrus.Infof("\u2714 Content: found from file %q:\n%v", f.spec.File, result)

	return result, nil
}
