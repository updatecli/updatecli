package file

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
)

// TODO: deprecate `spec.file` & `kind: file`, to be replaced by `spec.files` & `kind: files`
// TODO:! tests
// TODO:! update doc

/* [Meta to be removed]
 * "TODO:": to be kept in code (?)
 * "TODO:!": personal todo for preparing the PR (to be removed from draft)
 * "TODO:?": questions for maitainers (to be removed after review)
 */

// Spec defines a specification for a "file" resource
// parsed from an updatecli manifest file
type Spec struct {
	File           string
	Files          []string
	Line           int
	Content        string
	ForceCreate    bool
	MatchPattern   string
	ReplacePattern string
}

// File defines a resource of kind "file"
type File struct {
	spec             Spec
	contentRetriever text.TextRetriever
	files            map[string]string
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

	err = newResource.Validate()
	if err != nil {
		return nil, err
	}

	newResource.files = make(map[string]string)
	// file as simple item of files
	if len(newResource.spec.File) > 0 {
		newResource.files[strings.TrimPrefix(newResource.spec.File, "file://")] = ""
	}
	// files
	for _, file := range newResource.spec.Files {
		// TODO:? warn if already in? (duplicates)
		// TODO:! only add if not already in
		newResource.files[strings.TrimPrefix(file, "file://")] = ""
	}

	return newResource, nil
}

// TODO:! change sign by (s *Spec)
// Validate validates the object and returns an error (with all the failed validation messages) if not valid
func (f *File) Validate() error {
	// TODO:! replace by a strings.Builder
	var validationErrors []string

	// Check for all validation
	if len(f.spec.Files) == 0 && len(f.spec.File) == 0 {
		// TODO: to be updated after 'file' deprecation
		validationErrors = append(validationErrors, "Invalid spec for file resource: both 'file' and 'files' are empty.")
	}
	if len(f.spec.Files) > 0 && len(f.spec.File) > 0 {
		validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.file` and `spec.files` are mutually exclusive")
	}
	if len(f.spec.Files) > 1 && f.spec.Line != 0 {
		validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.files` and `spec.line` are mutually exclusive if there is more than one file")
	}
	if f.spec.Line < 0 {
		validationErrors = append(validationErrors, "Line cannot be negative for a file resource.")
	}
	if f.spec.Line > 0 {
		if f.spec.ForceCreate {
			validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.forcecreate` and `spec.line` are mutually exclusive")
		}

		if len(f.spec.MatchPattern) > 0 {
			validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.matchpattern` and `spec.line` are mutually exclusive")
		}

		if len(f.spec.ReplacePattern) > 0 {
			validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.replacepattern` and `spec.line` are mutually exclusive")
		}
	}
	if len(f.spec.Content) > 0 && len(f.spec.ReplacePattern) > 0 {
		validationErrors = append(validationErrors, "Validation error in target of type 'file': the attributes `spec.replacepattern` and `spec.line` are mutually exclusive")
	}

	// Return all the validation errors if any
	if len(validationErrors) > 0 {
		return fmt.Errorf("Validation error: the provided manifest configuration had the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	return nil
}

// Read puts the content of the file(s) as value of the f.files map if the file(s) exist(s) or log the non existence of the file
func (f *File) Read() error {
	var err error

	// Retrieve files content
	for filePath := range f.files {
		if f.contentRetriever.FileExists(filePath) {
			// Return the specified line if a positive number is specified by user in its manifest
			// Note that in this case we're with a fileCount of 1 (as other cases wouldn't pass validation)
			if f.spec.Line > 0 {
				f.files[filePath], err = f.contentRetriever.ReadLine(filePath, f.spec.Line)
				if err != nil {
					return err
				}
			}

			// Otherwise return the textual content
			f.files[filePath], err = f.contentRetriever.ReadAll(filePath)
			if err != nil {
				return err
			}
		} else {
			if f.spec.ForceCreate {
				// f.files[filePath] is already set to "", no need for more except logging
				logrus.Infof("Creating a new file at %q", filePath)
			} else {
				return fmt.Errorf("%s The specified file %q does not exist. If you want to create it, you must set the attribute 'spec.forcecreate' to 'true'.\n", result.FAILURE, filePath)
			}
		}
	}
	return nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (f *File) Changelog() string {
	return ""
}
