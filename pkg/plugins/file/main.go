package file

// FileSpec defines a specification for a "file" resource
// parsed from an updatecli manifest file
type FileSpec struct {
	File    string
	Line    Line
	Content string
}

// Files defines a resource of type "file"
type File struct {
	spec FileSpec
}

// New returns a reference to a newly initialized File object from a Filespec
// or an error if the provided Filespec triggers a validation error.
func New(spec FileSpec) (*File, error) {
	if spec.File == "" {
		return nil, &ErrEmptyFilePath{}
	}
	return &File{
		spec: spec,
	}, nil
}
