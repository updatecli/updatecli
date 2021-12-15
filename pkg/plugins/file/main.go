package file

import (
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/text"
)

// FileSpec defines a specification for a "file" resource
// parsed from an updatecli manifest file
type FileSpec struct {
	File           string
	Line           int
	Content        string
	ForceCreate    bool
	MatchPattern   string
	ReplacePattern string
}

// File defines a resource of type "file"
type File struct {
	spec             FileSpec
	contentRetriever text.TextRetriever
	CurrentContent   string
}

// New returns a reference to a newly initialized File object from a Filespec
// or an error if the provided Filespec triggers a validation error.
func New(spec FileSpec) (*File, error) {
	newFileResource := &File{
		spec:             spec,
		contentRetriever: &text.Text{},
	}
	// TODO: generalize the Validate + Normalize as an interface to all resources
	err := newFileResource.Validate()
	if err != nil {
		return nil, err
	}

	err = newFileResource.Normalize()
	if err != nil {
		return nil, err
	}

	return newFileResource, nil
}

// Normalize ensures that the attributes of the object are following the expected conventions
func (f *File) Normalize() error {
	f.spec.File = strings.TrimPrefix(f.spec.File, "file://")

	return nil
}

// Validate validates the object and returns an error (with all the failed validation messages) if it is not valid
func (f *File) Validate() error {
	var validationErrors []string

	// Check for all validation
	if f.spec.File == "" {
		validationErrors = append(validationErrors, "Invalid spec for file resource: 'file' is empty.")
	} else {
		if !f.contentRetriever.FileExists(f.spec.File) {
			validationErrors = append(validationErrors, fmt.Sprintf("Invalid spec for yaml resource: the file %q does not exist.", f.spec.File))
		}
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
	// Return all the validation errors if found any
	if len(validationErrors) > 0 {
		return fmt.Errorf("Validation error: the provided manifest configuration had the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	return nil
}

// Read defines CurrentContent to the content of the file which path is specified in spec.File
func (f *File) Read() error {
	// Return the specified line if a positive number is specified by user in its manifest
	if f.spec.Line > 0 {
		line, err := f.contentRetriever.ReadLine(f.spec.File, f.spec.Line)
		if err != nil {
			return err
		}

		f.CurrentContent = line
		return nil
	}

	// Otherwise return the textual content
	textContent, err := f.contentRetriever.ReadAll(f.spec.File)
	if err != nil {
		return err
	}
	f.CurrentContent = textContent

	return nil
}
