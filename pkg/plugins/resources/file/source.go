package file

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source return a file content
func (f *File) Source(workingDir string) (string, error) {
	var validationErrors []string

	if len(f.spec.ReplacePattern) > 0 {
		validationErrors = append(validationErrors, "Validation error in source of type 'file': the attribute `spec.replacepattern` is only supported for targets.")
	}
	if len(f.spec.Content) > 0 {
		validationErrors = append(validationErrors, "Validation error in source of type 'file': the attribute `spec.content` is only supported for conditions and targets.")
	}
	if f.spec.ForceCreate {
		validationErrors = append(validationErrors, "Validation error in source of type 'file': the attribute `spec.forcecreate` is only supported for targets.")
	}
	// Return all the validation errors if found any
	if len(validationErrors) > 0 {
		return "", fmt.Errorf("Validation error: the provided manifest configuration had the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	if !filepath.IsAbs(f.spec.File) {
		f.spec.File = filepath.Join(workingDir, f.spec.File)
		logrus.Debugf("Relative path detected: changing to absolute path from working directory: %q", f.spec.File)
	}

	if err := f.Read(); err != nil {
		return "", err
	}

	foundContent := f.CurrentContent
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

		foundContent = strings.Join(matchedStrings, "\n")
	}

	logrus.Infof("%s Content: found from file %q:\n%v", result.SUCCESS, f.spec.File, foundContent)

	return foundContent, nil
}
