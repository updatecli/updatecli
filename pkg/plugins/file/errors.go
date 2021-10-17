package file

// ErrEmptyFilePath is the error when the specification of a file resource has an empty File
type ErrEmptyFilePath struct{}

func (e *ErrEmptyFilePath) Error() string {
	return "Invalid spec for file resource: 'file' is empty."
}

// ErrNegativeLine is the validation error when the provided attribute is a negative number
type ErrNegativeLine struct{}

func (e *ErrNegativeLine) Error() string {
	return "Line cannot be negative for a file resource."
}
