package file

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/core/text"
)

// TODO: deprecate `spec.file` & `kind: file`, to be replaced by `spec.files` & `kind: files`
// TODO:! tests
// TODO:! update doc
// TODO:? rename `spec.files` to `spec.filePaths`?

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
	CurrentContent   string // TODO:! to be removed, no need if we treat "file" as one occurrence of "fileList" (to be renamed "files")
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
=======
	// TODO:? shouldn't this validation be after the trim prefix?
	// TODO: generalize the Validate + Normalize as an interface to all resources
	err := newResource.Validate()
>>>>>>> 6d73233 (wip: implement 'spec.files' for 'kind: file' (first part: 'targets')):pkg/plugins/file/main.go
	if err != nil {
		return nil, err
	}

	newResource.files = make(map[string]string)
	// TODO:? where does this 'file://' comes from? Is it a common case? Should be tested/striped elsewhere IMO (hlm)
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

// Validate validates the object and returns an error (with all the failed validation messages) if not valid
func (f *File) Validate() error {
	// TODO:! replace by a strings.Builder
	var validationErrors []string

	fileCount := len(f.spec.Files)
	if len(f.spec.File) > 0 {
		fileCount++
	}

	// Check for all validation
	if fileCount == 0 {
		// TODO: to be updated after 'file' deprecation
		validationErrors = append(validationErrors, "Invalid spec for file resource: both 'file' and 'files' are empty.")
	} else {
		// TODO:! check if line & files>1 could be compatible
		if fileCount > 1 && f.spec.Line != 0 {
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
	}

	// Return all the validation errors if any
	if len(validationErrors) > 0 {
		return fmt.Errorf("Validation error: the provided manifest configuration had the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	return nil
}

// TODO: to be superseded by ReadOrForceCreate
// Read puts the content of the file(s) as value of the f.files map
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
			return fmt.Errorf("%s The specified file %q does not exist. If you want to create it, you must set the attribute 'spec.forcecreate' to 'true'.\n", result.FAILURE, f.spec.File)
		}
	}
	return nil
}

// Read puts the content of the file(s) as value of the f.files map if the file(s) exist(s) or else creates the file(s)
func (f *File) ReadOrForceCreate() error {
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
