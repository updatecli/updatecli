package file

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
)

// Source return a file content
func (f *File) Source(workingDir string) (string, error) {
	var validationErrors []string
	var foundContent string
	var oldFilePath string
	var newFilePath string

	if len(f.spec.Files) > 1 {
		validationErrors = append(validationErrors, "validation error in source of type 'file': the attributes `spec.files` can't contain more than one element for sources")
	}
	if len(f.spec.ReplacePattern) > 0 {
		validationErrors = append(validationErrors, "validation error in source of type 'file': the attribute `spec.replacepattern` is only supported for targets")
	}
	if len(f.spec.Content) > 0 {
		validationErrors = append(validationErrors, "validation error in source of type 'file': the attribute `spec.content` is only supported for conditions and targets")
	}
	if f.spec.ForceCreate {
		validationErrors = append(validationErrors, "validation error in source of type 'file': the attribute `spec.forcecreate` is only supported for targets")
	}
	// Return all the validation errors if found any
	if len(validationErrors) > 0 {
		return "", fmt.Errorf("validation error: the provided manifest configuration had the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	// Looping on the only filePath in 'files'
	for filePath := range f.files {
		// Relative path is used when an SCM is associated with the file resource: means the file is on a remote SCM (hence relative path)
		if !text.IsURL(filePath) && !filepath.IsAbs(filePath) {
			newFilePath = filepath.Join(workingDir, filePath)
			oldFilePath = filePath
			logrus.Debugf("relative path detected: changing to absolute path from working directory: %q", filePath)
		}
	}
	// Replace old file path
	if newFilePath != "" {
		delete(f.files, oldFilePath)
		file := f.files[newFilePath]
		file.content = ""
		f.files[newFilePath] = file
	}

	if err := f.Read(); err != nil {
		return "", err
	}

	// Looping on the only filePath in 'files'
	for filePath := range f.files {
		foundContent = f.files[filePath].content
		// If a matchPattern is specified, then retrieve the string matched and returns the (eventually) multi-line string
		if len(f.spec.MatchPattern) > 0 {
			reg, err := regexp.Compile(f.spec.MatchPattern)
			if err != nil {
				logrus.Errorf("validation error in source of type 'file': Unable to parse the regexp specified at f.spec.MatchPattern (%q)", f.spec.MatchPattern)
				return "", err
			}

			// Check if there is any match in the file
			if !reg.MatchString(f.files[filePath].content) {
				return "", fmt.Errorf("no line matched in the file %q for the pattern %q", filePath, f.spec.MatchPattern)
			}
			matchedStrings := reg.FindAllString(f.files[filePath].content, -1)

			foundContent = strings.Join(matchedStrings, "\n")
		}

		logrus.Infof("%s Content: found from file %q:\n%v", result.SUCCESS, filePath, foundContent)

	}
	return foundContent, nil
}
