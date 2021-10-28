package file

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/scm"
	"github.com/updatecli/updatecli/pkg/core/text"
)

// Condition test if a file content match the content provided via configuration.
// If the configuration doesn't specify a value then it fall back to the source output
func (f *File) Condition(source string) (bool, error) {
	return f.checkFileCondition(source)
}

// ConditionFromSCM test if a file content from SCM match the content provided via configuration.
// If the configuration doesn't specify a value then it fall back to the source output
func (f *File) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {
	if !filepath.IsAbs(f.spec.File) {
		f.spec.File = filepath.Join(scm.GetDirectory(), f.spec.File)
	}
	return f.checkFileCondition(source)
}

func (f *File) checkFileCondition(source string) (bool, error) {
	// Start by retrieving the specified file's content
	if err := f.Read(); err != nil {
		return false, err
	}

	logMessage := fmt.Sprintf("Content of the file %q", f.spec.File)
	if f.spec.Line > 0 {
		logMessage = fmt.Sprintf("Content of the file %q (line %d)", f.spec.File, f.spec.Line)
	}

	if len(f.spec.ReplacePattern) > 0 {
		validationError := fmt.Errorf("Validation error in condition of type 'file': the attribute `spec.replacepattern` is only supported for targets.")
		logrus.Errorf(validationError.Error())
		return false, validationError
	}

	// If a matchPattern is specified, then return its result
	if len(f.spec.MatchPattern) > 0 {
		reg, err := regexp.Compile(f.spec.MatchPattern)
		if err != nil {
			logrus.Errorf("Validation error in condition of type 'file': Unable to parse the regexp specified at f.spec.MatchPattern (%q)", f.spec.MatchPattern)
			return false, err
		}

		return reg.MatchString(f.CurrentContent), nil
	}

	// When a source is provided, try to compare the file content with the source
	if len(source) > 0 {
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
				text.Diff(f.CurrentContent, source),
			)

			return false, nil
		}

		logrus.Infof("\u2714 %s is the same as the input source value.", logMessage)

		return true, nil

	}

	// No sourceID provided: the specified attribute must be used to determine which content to compare the file with
	if len(f.spec.Content) == 0 && f.spec.Line > 0 {
		// No content + no source input values means the user only want to check if the line "exists" (e.g. is not empty) and that's all
		if f.CurrentContent == "" {
			logrus.Infof("\u2717 %s is empty or the file does not exist.", logMessage)
			return false, nil
		}

		logrus.Infof("\u2714 %s is not empty and the file exists.", logMessage)
		return true, nil
	}

	if f.spec.Content != f.CurrentContent {
		logrus.Infof("\u2717 %s is different than the specified content: \n%s",
			logMessage, text.Diff(f.CurrentContent, f.spec.Content))

		return false, nil
	}

	logrus.Infof("\u2714 %s is the same as the specified content.", logMessage)
	return true, nil
}
