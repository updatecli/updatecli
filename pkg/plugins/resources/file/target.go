package file

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
)

// Target creates or updates a file from a source control management system.
// The default content is the value retrieved from source
func (f *File) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {
	absoluteFiles := make(map[string]string)
	for filePath := range f.files {
		absoluteFilePath := filePath
		if !filepath.IsAbs(filePath) && scm != nil {
			absoluteFilePath = filepath.Join(scm.GetDirectory(), filePath)
			logrus.Debugf("Relative path detected: changing to absolute path from SCM: %q", absoluteFilePath)
		}
		absoluteFiles[absoluteFilePath] = f.files[filePath]
	}
	f.files = absoluteFiles

	return f.target(source, dryRun, resultTarget)
}

func (f *File) target(source string, dryRun bool, resultTarget *result.Target) error {
	var files []string

	if f.spec.Line > 0 && f.spec.ForceCreate {
		validationError := fmt.Errorf("validation error in target of type 'file': 'spec.line' and 'spec.forcecreate' are mutually exclusive")
		logrus.Errorf(validationError.Error())
		return validationError
	}

	// Test if target reference a file with a prefix like https:// or file://, as we don't know how to update those files.
	// We need to loop the files here before calling ReadOrForceCreate in case one of these file paths is an URL, not supported for a target
	for filePath := range f.files {
		if text.IsURL(filePath) {
			validationError := fmt.Errorf("validation error in target of type 'file': spec.files item value (%q) is an URL which is not supported for a target", filePath)
			logrus.Errorf(validationError.Error())
			return validationError
		}
	}

	// Retrieving content of file(s) in memory (nothing in case of spec.forceCreate)
	if err := f.Read(); err != nil {
		return err
	}

	originalContents := make(map[string]string)

	// Unless there is a content specified, the inputContent source value is used to fill the file
	inputContent := source
	if len(f.spec.Content) > 0 {
		inputContent = f.spec.Content
	}

	resultTarget.NewInformation = inputContent
	/*
		At the moment, we don't have an easy to identify that precise information
		that would be updated without considering the new file content.

		It's doable but out of the current scope of the effort.

		With a valid usecase we can improve the situation.

		Especially considering that we may have multiple files to update
	*/
	resultTarget.OldInformation = "unknown"

	// If we're using a regexp for the target
	if len(f.spec.MatchPattern) > 0 {
		// use source (or spec.content) as replace pattern input if no spec.replacepattern is specified
		if len(f.spec.ReplacePattern) > 0 {
			inputContent = f.spec.ReplacePattern
		}

		reg, err := regexp.Compile(f.spec.MatchPattern)
		if err != nil {
			logrus.Errorf("Validation error in target of type 'file': Unable to parse the regexp specified at f.spec.MatchPattern (%q)", f.spec.MatchPattern)
			return err
		}

		for filePath := range f.files {
			// Check if there is any match in the file
			if !reg.MatchString(f.files[filePath]) {
				// We allow the possibility to match only some files. In that case, just a warning here
				return fmt.Errorf("no line matched in file %q for pattern %q", filePath, f.spec.MatchPattern)
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
			logrus.Debugf("content from file %q already up to date", filePath)
		} else {
			files = append(files, filePath)
		}
	}
	if notChanged == len(f.files) {
		resultTarget.Description = "all contents from 'file' and 'files' combined already up to date"
		resultTarget.Files = files
		resultTarget.Changed = false
		resultTarget.Result = result.SUCCESS

		logrus.Infoln(resultTarget.Description)

		return nil
	}

	sort.Strings(files)

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
			return err
		}

		resultTarget.Description = fmt.Sprintf("%s\nUpdated %s of file %q\n",
			resultTarget.Description,
			contentType,
			filePath)

		logrus.Debugf("%s updated %s of file %q\n\n%s\n",
			contentType,
			filePath,
			text.Diff(filePath, originalContents[filePath], f.files[filePath]),
		)
	}

	resultTarget.Description = strings.TrimPrefix(resultTarget.Description, "\n")

	resultTarget.Result = result.ATTENTION
	resultTarget.Changed = true
	resultTarget.Files = files

	return nil
}
