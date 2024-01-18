package dockerfile

import "errors"

var (
	// ErrTooManyVariables is returned when more than one variable is detected in the Dockerfile instruction.
	ErrTooManyVariables = errors.New("too many arguments")
)
