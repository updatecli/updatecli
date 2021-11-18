package file

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/scm"
	"github.com/updatecli/updatecli/pkg/core/text"
)

// Condition test if a file content match the content provided via configuration.
// If the configuration doesn't specify a value then it fall back to the source output
func (f *File) Condition(source string) (bool, error) {
	return f.checkCondition(source)
}

// ConditionFromSCM test if a file content from SCM match the content provided via configuration.
// If the configuration doesn't specify a value then it fall back to the source output
func (f *File) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {
	if !filepath.IsAbs(f.spec.File) {
		f.spec.File = filepath.Join(scm.GetDirectory(), f.spec.File)
		logrus.Debugf("Relative path detected: changing to absolute path from SCM: %q", f.spec.File)
	}
	return f.checkCondition(source)
}

func (f *File) checkCondition(source string) (bool, error) {
	passing, err := f.condition(source)
	if err != nil {
		logrus.Infof("%s Condition on file %q errored", result.FAILURE, f.spec.File)
	} else {
		if passing {
			logrus.Infof("%s Condition on file %q passed", result.SUCCESS, f.spec.File)
		} else {
			logrus.Infof("%s Condition on file %q did not pass", result.FAILURE, f.spec.File)
		}
	}

	return passing, err
}

func (f *File) condition(source string) (bool, error) {
	var validationErrors []string

	if len(f.spec.ReplacePattern) > 0 {
		validationErrors = append(validationErrors, "Validation error in condition of type 'file': the attribute `spec.replacepattern` is only supported for targets.")
	}
	if f.spec.ForceCreate {
		validationErrors = append(validationErrors, "Validation error in condition of type 'file': the attribute `spec.forcecreate` is only supported for targets.")
	}
	// Return all the validation errors if found any
	if len(validationErrors) > 0 {
		return false, fmt.Errorf("Validation error: the provided manifest configuration had the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	// Start by retrieving the specified file's content
	logrus.Debugf("Reading file %q", f.spec.File)
	if err := f.Read(); err != nil {
		logrus.Debugf("Error while reading file: %q", err.Error())
		return false, err
	}

	logMessage := fmt.Sprintf("Content of the file %q", f.spec.File)
	if f.spec.Line > 0 {
		logMessage = fmt.Sprintf("Content of the file %q (line %d)", f.spec.File, f.spec.Line)
	}

	// If a matchPattern is specified, then return its result
	if len(f.spec.MatchPattern) > 0 {
		logrus.Debugf("Attribute 'matchpattern' found: %s", f.spec.MatchPattern)
		reg, err := regexp.Compile(f.spec.MatchPattern)
		if err != nil {
			logrus.Errorf("Validation error in condition of type 'file': Unable to parse the regexp specified at f.spec.MatchPattern (%q)", f.spec.MatchPattern)
			return false, err
		}

		if !reg.MatchString(f.CurrentContent) {
			logrus.Infof(
				"\u2717 %s did not match the pattern %q",
				logMessage,
				f.spec.MatchPattern,
			)
			return false, nil
		}
		logrus.Infof("\u2714 %s matched the pattern %q", logMessage, f.spec.MatchPattern)
		return true, nil
	}

	// When a source is provided, try to compare the file content with the source
	if len(source) > 0 {
		logrus.Debugf("Using source input value: %q", source)
		if len(f.spec.Content) > 0 {
			validationError := fmt.Errorf("Validation error in condition of type 'file': the attributes `sourceID` and `spec.content` are mutually exclusive")
			logrus.Errorf(validationError.Error())
			return false, validationError
		}

		// Compare the content of the file with the source's value
		if f.CurrentContent != source {
			logrus.Infof(
				"\u2717 %s is different than the input source value:\n%s",
				logMessage,
				text.Diff(f.spec.File, f.CurrentContent, source),
			)

			return false, nil
		}

		logrus.Infof("\u2714 %s is the same as the input source value.", logMessage)

		return true, nil

	}

	// No sourceID provided: the specified attribute must be used to determine which content to compare the file with
	logrus.Debug("No source input value (disabled or empty)")
	if len(f.spec.Content) == 0 {
		logrus.Debug("No attribute 'content' provided")
		// No content + no source input values means the user only want to check if the line "exists" (e.g. is not empty) and that's all
		if f.spec.Line > 0 {
			if f.CurrentContent == "" {
				logrus.Infof("\u2717 %s is empty or the file does not exist.", logMessage)
				return false, nil
			}

			logrus.Infof("\u2714 %s is not empty and the file exists.", logMessage)
			return true, nil
		}

		// No source, no content, no line: Only check for existence of the file
		return f.contentRetriever.FileExists(f.spec.File), nil
	}

	logrus.Debug("Attribute `content` detected")

	if f.spec.Content != f.CurrentContent {
		logrus.Infof("\u2717 %s is different than the specified content: \n%s",
			logMessage, text.Diff(f.spec.File, f.CurrentContent, f.spec.Content))

		return false, nil
	}

	logrus.Infof("\u2714 %s is the same as the specified content.", logMessage)
	return true, nil
}
