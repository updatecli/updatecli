package file

// ErrEmptyFilePath is the error when the specification of a file resource has an empty File
type ErrEmptyFilePath struct{}

func (e *ErrEmptyFilePath) Error() string {
	return "Invalid spec for file resource: 'file' is empty."
}

//ErrLineNotFound is the error message when no matching line found
type ErrLineNotFound struct{}

func (e *ErrLineNotFound) Error() string {
	return "line not found in the specified file."
}
