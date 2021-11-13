package file

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/scm"
	"github.com/updatecli/updatecli/pkg/core/text"
)

// Target creates or updates a file located locally.
// The default content is the value retrieved from source
func (f *File) Target(source string, dryRun bool) (bool, error) {
	changed, _, _, err := f.target(source, dryRun)
	return changed, err
}

// TargetFromSCM creates or updates a file from a source control management system.
// The default content is the value retrieved from source
func (f *File) TargetFromSCM(source string, scm scm.Scm, dryRun bool) (bool, []string, string, error) {
	if !filepath.IsAbs(f.spec.File) {
		f.spec.File = filepath.Join(scm.GetDirectory(), f.spec.File)
	}
	return f.target(source, dryRun)
}

func (f *File) target(source string, dryRun bool) (bool, []string, string, error) {
	var files []string
	var message string

	// Test if target reference a file with a prefix like https:// or file://, as we don't know how to update those files.
	if text.IsURL(f.spec.File) {
		validationError := fmt.Errorf("Validation error in target of type 'file': spec.File value (%q) is an URL which is not supported for a target.", f.spec.File)
		logrus.Errorf(validationError.Error())
		return false, files, message, validationError
	}

	// Case 1: The attribute 'spec.line' is specified, we only want to change a specific line of the target file
	if f.spec.Line > 0 {
		if f.spec.ForceCreate {
			validationError := fmt.Errorf("Validation error in target of type 'file': 'spec.line' and 'spec.forcecreate' are mutually exclusive")
			logrus.Errorf(validationError.Error())
			return false, files, message, validationError
		}

		// Check that the specified exists or exit with error
		if !f.contentRetriever.FileExists(f.spec.File) {
			return false, files, message, os.ErrNotExist
		}

		// Retrieve the actual line of the specified file (to check if a change is needed and for report)
		currentLine, err := f.contentRetriever.ReadLine(f.spec.File, f.spec.Line)
		if err != nil {
			return false, files, message, err
		}

		// Unless there is a content specified, the input source value is used to fill the file
		newContent := source
		if len(f.spec.Content) > 0 {
			newContent = f.spec.Content
		}

		// Nothing to do if the line is the same as the input source value
		if currentLine == newContent {
			return false, files, message, err
		}

		// Otherwise, change the line (if not in dry run enabled)
		if !dryRun {
			if err := f.contentRetriever.WriteLineToFile(newContent, f.spec.File, f.spec.Line); err != nil {
				return false, files, message, err
			}
		}

		files = append(files, f.spec.File)
		message = fmt.Sprintf("changed line %d of file %q", f.spec.Line, f.spec.File)

		logrus.Infof("\u2714 The line %d of the file %q was updated: \n\n%s\n", f.spec.Line, f.spec.File, text.Diff(f.spec.File, currentLine, newContent))

		return true, files, message, nil

	}

	//// Default case: change the content of the whole file
	// Check for the existence fo the file:
	// - If the file exists then we retrieve its content in the 'f' object
	// - If it does not exist and it's not specified to be created, then exit on error
	// - If it does not exist and should be created, then continue with "empty string" as initial content
	// - Otherwise exit with the reported error from os.Stat
	if f.contentRetriever.FileExists(f.spec.File) {
		if err := f.Read(); err != nil {
			return false, files, message, err
		}
	} else {
		if !f.spec.ForceCreate {
			return false, files, message, fmt.Errorf("\u2717 The specified file %q does not exist. If you want to create it, you must set the attribute 'spec.forcecreate' to 'true'.\n", f.spec.File)
		}
		logrus.Infof("Creating a new file at %q", f.spec.File)
	}

	// Unless there is a content specified, the input source value is used to fill the file
	newContent := source
	if len(f.spec.Content) > 0 {
		newContent = f.spec.Content
	}
	if len(f.spec.MatchPattern) > 0 {
		// use source (or spec.content) as replace pattern if no spec.replacepattern is specified
		replacePattern := newContent
		if len(f.spec.ReplacePattern) > 0 {
			replacePattern = f.spec.ReplacePattern
		}

		reg, err := regexp.Compile(f.spec.MatchPattern)
		if err != nil {
			logrus.Errorf("Validation error in target of type 'file': Unable to parse the regexp specified at f.spec.MatchPattern (%q)", f.spec.MatchPattern)
			return false, files, message, err
		}
		// Check if there is any match in the file
		if !reg.MatchString(f.CurrentContent) {
			return false, files, message, fmt.Errorf("No line matched in the file %q for the pattern %q", f.spec.File, f.spec.MatchPattern)
		}
		newContent = reg.ReplaceAllString(f.CurrentContent, replacePattern)
	}

	// Nothing to do if the line is the same as the input source value
	if newContent == f.CurrentContent {
		logrus.Infof("\u2714 Content from file %q already up to date", f.spec.File)
		return false, files, message, nil
	}

	// otherwise write the new content to the file (if not in dry run enabled)
	if !dryRun {
		err := f.contentRetriever.WriteToFile(newContent, f.spec.File)
		if err != nil {
			return false, files, message, err
		}
	}

	files = append(files, f.spec.File)
	message = fmt.Sprintf("Updated the file %q\n", f.spec.File)

	logrus.Infof("\u2714 File content for %q, updated. \n\n%s\n", f.spec.File, text.Diff(f.spec.File, f.CurrentContent, newContent))

	return true, files, message, nil
}
