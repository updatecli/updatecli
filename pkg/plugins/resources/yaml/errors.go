package yaml

import "errors"

var (
	// ErrKeyNotFound is returned when a key is not found in a yaml file
	ErrKeyNotFound = errors.New("key not found")
)
