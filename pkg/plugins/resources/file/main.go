package file

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
)

// Spec defines a specification for a "file" resource
// parsed from an updatecli manifest file
type Spec struct {
	// File contains the file path(s) to take in account and is incompatible with Files
	File string `yaml:",omitempty"`
	// Files contains the file path(s) to take in account and is incompatible with File
	Files []string `yaml:",omitempty"`
	// Line contains the line of the file(s) to take in account
	Line int `yaml:",omitempty"`
	// Content specifies the content to take in account instead of the file content
	Content string `yaml:",omitempty"`
	// ForceCreate specifies if nonexistent file(s) should be created if they are targets
	ForceCreate bool `yaml:",omitempty"`
	// MatchPattern specifies the regexp pattern to match on the file(s)
	MatchPattern string `yaml:",omitempty"`
	// ReplacePattern specifies the regexp replace pattern to apply on the file(s) content
	ReplacePattern string `yaml:",omitempty"`
}

// File defines a resource of kind "file"
type File struct {
	spec             Spec
	contentRetriever text.TextRetriever
	files            map[string]fileMetadata // map of file
}

type fileMetadata struct {
	originalPath string
	path         string
	content      string
}

// New returns a reference to a newly initialized File object from a Spec
// or an error if the provided Filespec triggers a validation error.
func New(spec interface{}) (*File, error) {
	newSpec := Spec{}
	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	newResource := &File{
		spec:             newSpec,
		contentRetriever: &text.Text{},
	}

	err = newResource.spec.Validate()
	if err != nil {
		return nil, err
	}

	newResource.files = make(map[string]fileMetadata)
	// File as unique element of newResource.files
	if len(newResource.spec.File) > 0 {
		f := fileMetadata{
			path:         strings.TrimPrefix(newResource.spec.File, "file://"),
			originalPath: strings.TrimPrefix(newResource.spec.File, "file://"),
		}
		newResource.files[newResource.spec.File] = f
	}

	for _, filePath := range newResource.spec.Files {
		f := fileMetadata{
			path:         strings.TrimPrefix(filePath, "file://"),
			originalPath: strings.TrimPrefix(filePath, "file://"),
		}
		newResource.files[filePath] = f
	}

	return newResource, nil
}

func hasDuplicates(values []string) bool {
	uniqueValues := make(map[string]string)
	for _, v := range values {
		uniqueValues[v] = ""
	}

	return len(values) != len(uniqueValues)
}

func (f *File) UpdateAbsoluteFilePath(workDir string) {
	for filePath := range f.files {
		if workDir != "" {
			file := f.files[filePath]
			file.path = joinPathWithWorkingDirectoryPath(file.originalPath, workDir)

			logrus.Debugf("Relative path detected: changing from %q to absolute path from SCM: %q", file.originalPath, file.path)
			f.files[filePath] = file
		}
	}
}

// Validate validates the object and returns an error (with all the failed validation messages) if not valid
func (s *Spec) Validate() error {
	var validationErrors []string

	// Check for all validation
	if len(s.Files) == 0 && len(s.File) == 0 {
		validationErrors = append(validationErrors, "Invalid spec for file resource: both 'file' and 'files' are empty.")
	}
	if len(s.Files) > 0 && len(s.File) > 0 {
		validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.file` and `spec.files` are mutually exclusive")
	}
	if len(s.Files) > 1 && s.Line != 0 {
		validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.files` and `spec.line` are mutually exclusive if there is more than one file")
	}
	if len(s.Files) > 1 && hasDuplicates(s.Files) {
		validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.files` contains duplicated values")
	}
	if s.Line < 0 {
		validationErrors = append(validationErrors, "Line cannot be negative for a file resource.")
	}
	if s.Line > 0 {
		if s.ForceCreate {
			validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.forcecreate` and `spec.line` are mutually exclusive")
		}

		if len(s.MatchPattern) > 0 {
			validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.matchpattern` and `spec.line` are mutually exclusive")
		}

		if len(s.ReplacePattern) > 0 {
			validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.replacepattern` and `spec.line` are mutually exclusive")
		}
	}
	if len(s.Content) > 0 && len(s.ReplacePattern) > 0 {
		validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.replacepattern` and `spec.line` are mutually exclusive")
	}

	// Return all the validation errors if any
	if len(validationErrors) > 0 {
		return fmt.Errorf("validation error: the provided manifest configuration had the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	return nil
}

// Read puts the content of the file(s) as value of the f.files map if the file(s) exist(s) or log the non existence of the file
func (f *File) Read() error {
	var err error

	// Retrieve files content
	for filePath := range f.files {
		file := f.files[filePath]
		if f.contentRetriever.FileExists(file.path) {
			// Return the specified line if a positive number is specified by user in its manifest
			// Note that in this case we're with a fileCount of 1 (as other cases wouldn't pass validation)
			if f.spec.Line > 0 {
				file.content, err = f.contentRetriever.ReadLine(file.path, f.spec.Line)
				if err != nil {
					return err
				}
			}

			// Otherwise return the textual content
			if f.spec.Line == 0 {
				file.content, err = f.contentRetriever.ReadAll(file.path)
				if err != nil {
					return err
				}
			}
		} else {
			if f.spec.ForceCreate {
				// f.files[filePath] is already set to "", no need for more except logging
				logrus.Infof("Creating a new file at %q", file.originalPath)
			} else {
				if f.spec.Line > 0 {
					return fmt.Errorf("%s The specified line %d of the file %q does not exist", result.FAILURE, f.spec.Line, file.originalPath)
				}
				return fmt.Errorf("%s The specified file %q does not exist. If you want to create it, you must set the attribute 'spec.forcecreate' to 'true'", result.FAILURE, filePath)
			}
		}
		f.files[filePath] = file
	}
	return nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (f *File) Changelog() string {
	return ""
}
