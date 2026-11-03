package yaml

import "errors"

var (
	// ErrKeyNotFound is returned when a key is not found in a yaml file
	ErrKeyNotFound = errors.New("key not found")
)

// noYamlFileFoundMatchingCriteria reports a searchpattern that matched no file or no key
const noYamlFileFoundMatchingCriteria = "no yaml file found matching criteria"
