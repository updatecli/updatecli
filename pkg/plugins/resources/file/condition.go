package file

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
)

// Condition test if a file content matches the content provided via configuration.
// If the configuration doesn't specify a value then it fall back to the source output
func (f *File) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	workDir := ""
	if scm != nil {
		workDir = scm.GetDirectory()
	}

	if err := f.initFiles(workDir); err != nil {
		return false, "", fmt.Errorf("init files: %w", err)
	}

	files := f.spec.Files
	files = append(files, f.spec.File)

	passing, err := f.condition(source)
	if err != nil {
		return false, "", fmt.Errorf("file condition: %w", err)
	}

	switch passing {
	case true:
		return true, fmt.Sprintf("condition on file %q passed", files), nil

	case false:
		return false, fmt.Sprintf("condition on file %q did not pass", files), nil
	}

	return false, "", fmt.Errorf("unexpected error happened on file. Please report to an issue")
}

func (f *File) condition(source string) (bool, error) {
	var validationErrors []string

	if len(f.spec.ReplacePattern) > 0 {
		validationErrors = append(validationErrors, "Validation error in condition of type 'file': the attribute `spec.replacepattern` is only supported for targets")
	}
	if f.spec.ForceCreate {
		validationErrors = append(validationErrors, "Validation error in condition of type 'file': the attribute `spec.forcecreate` is only supported for targets")
	}
	// Return all the validation errors if found any
	if len(validationErrors) > 0 {
		return false, fmt.Errorf("validation error: the provided manifest configuration had the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	// Start by retrieving the specified file's content
	logrus.Debugf("Reading file(s) %q", f.files)
	if err := f.Read(); err != nil {
		logrus.Debugf("Error while reading file(s): %q", err.Error())
		return false, err
	}

	for filePath, file := range f.files {
		logMessage := fmt.Sprintf("Content of the file %q", file.originalPath)
		if f.spec.Line > 0 {
			logMessage = fmt.Sprintf("Content of the file %q (line %d)", file.originalPath, f.spec.Line)
		}

		// If a matchPattern is specified, then return its result
		if len(f.spec.MatchPattern) > 0 {
			logrus.Debugf("Attribute 'matchpattern' found: %s", f.spec.MatchPattern)
			reg, err := regexp.Compile(f.spec.MatchPattern)
			if err != nil {
				logrus.Errorf("Validation error in condition of type 'file': Unable to parse the regexp specified at f.spec.MatchPattern (%q)", f.spec.MatchPattern)
				return false, err
			}

			if !reg.MatchString(file.content) {
				if f.spec.SearchPattern {
					// When using both a file path pattern AND a content matching regex, then we want to ignore files that don't match the pattern
					// as otherwise we trigger error for files we don't care about.
					logrus.Debugf("No match found for pattern %q in file %q, removing it from the list of files to update", f.spec.MatchPattern, filePath)

					delete(f.files, filePath)

					if len(f.files) == 0 {
						logrus.Infof("no file found matching criteria %q", f.spec.MatchPattern)
						return false, nil
					}
					continue
				}

				logrus.Infof(
					"%s %s did not match the pattern %q",
					result.FAILURE,
					logMessage,
					f.spec.MatchPattern,
				)
				return false, nil
			}

			logrus.Infof("%s %s matched the pattern %q", result.SUCCESS, logMessage, f.spec.MatchPattern)
		}

		// When a source is provided, try to compare the file content with the source
		if len(source) > 0 {
			logrus.Debugf("Using source input value: %q", source)
			if len(f.spec.Content) > 0 {
				validationError := fmt.Errorf("validation error in condition of type 'file': the attributes `sourceid` and `spec.content` are mutually exclusive")
				logrus.Errorln(validationError.Error())
				return false, validationError
			}

			// Compare the content of the file with the source's value
			if file.content != source {
				logrus.Infof(
					"%s %s is different than the input source value:\n%s",
					result.FAILURE,
					logMessage,
					text.Diff(filePath, filePath, file.content, source),
				)

				return false, nil
			}
			logrus.Infof("%s %s is the same as the input source value.", result.SUCCESS, logMessage)
		}

		// No sourceID provided: the specified attribute must be used to determine which content to compare the file with
		logrus.Debug("No source input value (disabled or empty)")
		if len(f.spec.Content) == 0 {
			logrus.Debug("No attribute 'content' provided")
			// No content + no source input values means the user only want to check if the line "exists" (e.g. is not empty) and that's all
			if f.spec.Line > 0 {
				if file.content == "" {
					logrus.Infof("%s %s is empty or the file does not exist.", result.FAILURE, logMessage)
					return false, nil
				}
				logrus.Infof("%s %s is not empty and the file exists.", result.SUCCESS, logMessage)
			}

			// No source, no content, no line: Only check for existence of the file
			return f.contentRetriever.FileExists(file.path), nil
		}

		logrus.Debug("Attribute `content` detected")

		if f.spec.Content != file.content {
			logrus.Infof("%s %s is different than the specified content: \n%s",
				result.FAILURE,
				logMessage,
				text.Diff(filePath, filePath, file.content, f.spec.Content),
			)
			return false, nil
		}
		logrus.Infof("%s %s is the same as the specified content.", result.SUCCESS, logMessage)

		f.files[filePath] = file
	}
	return true, nil
}
