package file

import (
	"strings"

	"github.com/updatecli/updatecli/pkg/core/text"
)

// FileSpec defines a specification for a "file" resource
// parsed from an updatecli manifest file
type FileSpec struct {
	File        string
	Line        int
	Content     string
	ForceCreate bool
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
	if spec.File == "" {
		return nil, &ErrEmptyFilePath{}
	}
	if spec.Line < 0 {
		return nil, &ErrNegativeLine{}
	}
	if strings.HasPrefix(spec.File, "file://") {
		spec.File = strings.TrimPrefix(spec.File, "file://")
	}
	return &File{
		spec:             spec,
		contentRetriever: &text.Text{},
	}, nil
}

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
