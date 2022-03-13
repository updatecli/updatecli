package file

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
)

// Target creates or updates a file on the local filesystem
func (f *File) Target(source, workingDir string, dryRun bool) (bool, []string, string, error) {
	// Ensure that we have the absolute path for each file
	absoluteFiles := make(map[string]string)
	for filePath, fileContent := range f.files {
		absoluteFilePath := filePath
		if !filepath.IsAbs(filePath) {
			absoluteFilePath = filepath.Join(workingDir, filePath)
			logrus.Debugf("Relative path detected: changing to absolute path from: %q", absoluteFilePath)
		}
		absoluteFiles[absoluteFilePath] = fileContent
	}
	f.files = absoluteFiles

	var files []string
	var message strings.Builder

	if f.spec.Line > 0 && f.spec.ForceCreate {
		validationError := fmt.Errorf("Validation error in target of type 'file': 'spec.line' and 'spec.forcecreate' are mutually exclusive")
		logrus.Errorf(validationError.Error())
		return false, files, message.String(), validationError
	}

	// Test if target references a file with a prefix like https:// or file://, as we don't know how to update those files.
	// We need to loop the files here before calling ReadOrForceCreate in case one of these file paths is an URL, not supported for a target
	for filePath := range f.files {
		if text.IsURL(filePath) {
			validationError := fmt.Errorf("Validation error in target of type 'file': spec.files item value (%q) is an URL which is not supported for a target.", filePath)
			logrus.Errorf(validationError.Error())
			return false, files, message.String(), validationError
		}
	}

	// Retrieving content of file(s) in memory (nothing in case of spec.forceCreate)
	if err := f.Read(); err != nil {
		return false, files, message.String(), err
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
			return false, files, message.String(), err
		}

		for filePath, fileContent := range f.files {
			// Check if there is any match in the file
			if !reg.MatchString(fileContent) {
				// We allow the possibility to match only some files. In that case, just a warning here
				return false, files, message.String(), fmt.Errorf("No line matched in the file %q for the pattern %q", filePath, f.spec.MatchPattern)
			}
			// Keep the original content for later comparison
			originalContents[filePath] = fileContent
			f.files[filePath] = reg.ReplaceAllString(fileContent, inputContent)
		}
	} else {
		for filePath, fileContent := range f.files {
			// Keep the original content for later comparison
			originalContents[filePath] = fileContent
			f.files[filePath] = inputContent
		}
	}

	// Nothing to do if there is no content change
	notChanged := 0
	for filePath, fileContent := range f.files {
		if fileContent == originalContents[filePath] {
			notChanged++
			logrus.Infof("%s Content from file %q already up to date", result.SUCCESS, filePath)
		} else {
			files = append(files, filePath)
		}
	}
	if notChanged == len(f.files) {
		logrus.Infof("%s All contents from 'file' and 'files' combined already up to date", result.SUCCESS)
		return false, files, message.String(), nil
	}
	sort.Strings(files)
	// Otherwise write the new content to the file(s), or nothing but logs if dry run is enabled
	for filePath, fileContent := range f.files {
		var contentType string
		var err error
		if dryRun {
			contentType = "[dry run] content"
			if f.spec.Line > 0 {
				contentType = fmt.Sprintf("[dry run] line %d", f.spec.Line)
			}
		}
		if f.spec.Line == 0 && !dryRun {
			err = f.contentRetriever.WriteToFile(fileContent, filePath)
			contentType = "content"
		}
		if f.spec.Line > 0 && !dryRun {
			err = f.contentRetriever.WriteLineToFile(fileContent, filePath, f.spec.Line)
			contentType = fmt.Sprintf("line %d", f.spec.Line)
		}
		if err != nil {
			return false, files, message.String(), err
		}
		logrus.Infof("%s updated the %s of the file %q\n\n%s\n",
			result.ATTENTION,
			contentType,
			filePath,
			text.Diff(filePath, originalContents[filePath], fileContent),
		)
		message.WriteString(fmt.Sprintf("Updated the %s of the file %q\n", contentType, filePath))
	}

	return true, files, message.String(), nil
}
