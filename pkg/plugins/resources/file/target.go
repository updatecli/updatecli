package file

import (
	"fmt"
	"regexp"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
)

// Target creates or updates a file located locally.
// The default content is the value retrieved from source
func (f *File) Target(source string, dryRun bool) (bool, error) {
	if f.spec.Line > 0 && f.spec.ForceCreate {
		validationError := fmt.Errorf("Validation error in target of type 'file': 'spec.line' and 'spec.forcecreate' are mutually exclusive")
		logrus.Errorf(validationError.Error())
		return false, validationError
	}

	// Test if target reference a file with a prefix like https:// or file://, as we don't know how to update those files.
	// We need to loop the files here before calling ReadOrForceCreate in case one of these file paths is an URL, not supported for a target
	for filePath := range f.files {
		if text.IsURL(filePath) {
			validationError := fmt.Errorf("Validation error in target of type 'file': spec.files item value (%q) is an URL which is not supported for a target.", filePath)
			logrus.Errorf(validationError.Error())
			return false, validationError
		}
	}

	// Retrieving content of file(s) in memory (nothing in case of spec.forceCreate)
	if err := f.Read(); err != nil {
		return false, err
	}

	originalContents := make(map[string]string)

	// Unless there is a content specified, the inputContent source value is used to fill the file
	inputContent := source
	if len(f.spec.Content) > 0 {
		inputContent = f.spec.Content
	}

	// If we're using a regexp for the target
	if len(f.spec.MatchPattern) > 0 {
		// use source (or spec.content) as replace pattern input if no spec.replacepattern is specified
		if len(f.spec.ReplacePattern) > 0 {
			inputContent = f.spec.ReplacePattern
		}

		reg, err := regexp.Compile(f.spec.MatchPattern)
		if err != nil {
			logrus.Errorf("Validation error in target of type 'file': Unable to parse the regexp specified at f.spec.MatchPattern (%q)", f.spec.MatchPattern)
			return false, err
		}

		for filePath := range f.files {
			// Check if there is any match in the file
			if !reg.MatchString(f.files[filePath]) {
				// We allow the possibility to match only some files. In that case, just a warning here
				return false, fmt.Errorf("No line matched in the file %q for the pattern %q", filePath, f.spec.MatchPattern)
			}
			// Keep the original content for later comparison
			originalContents[filePath] = f.files[filePath]
			f.files[filePath] = reg.ReplaceAllString(f.files[filePath], inputContent)
		}
	} else {
		for filePath := range f.files {
			// Keep the original content for later comparison
			originalContents[filePath] = f.files[filePath]
			f.files[filePath] = inputContent
		}
	}

	// Nothing to do if there is no content change
	notChanged := 0
	for filePath := range f.files {
		if f.files[filePath] == originalContents[filePath] {
			notChanged++
			logrus.Infof("%s Content from file %q already up to date", result.SUCCESS, filePath)
		}
	}
	if notChanged == len(f.files) {
		logrus.Infof("%s All contents from 'file' and 'files' combined already up to date", result.SUCCESS)
		return false, nil
	}
	// Otherwise write the new content to the file(s), or nothing but logs if dry run is enabled
	for filePath := range f.files {
		var contentType string
		var err error
		if dryRun {
			contentType = "[dry run] content"
			if f.spec.Line > 0 {
				contentType = fmt.Sprintf("[dry run] line %d", f.spec.Line)
			}
		}
		if f.spec.Line == 0 && !dryRun {
			err = f.contentRetriever.WriteToFile(f.files[filePath], filePath)
			contentType = "content"
		}
		if f.spec.Line > 0 && !dryRun {
			err = f.contentRetriever.WriteLineToFile(f.files[filePath], filePath, f.spec.Line)
			contentType = fmt.Sprintf("line %d", f.spec.Line)
		}
		if err != nil {
			return false, err
		}
		logrus.Infof("%s updated the %s of the file %q\n\n%s\n",
			result.ATTENTION,
			contentType,
			filePath,
			text.Diff(filePath, originalContents[filePath], f.files[filePath]),
		)
	}

	return true, nil
}
