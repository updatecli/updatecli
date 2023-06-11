package yaml

import (
	"fmt"
	"strings"
)

/*
"yaml"  defines the specification for manipulating "yaml" files.
It can be used as a "source", a "condition", or a "target".
*/
type Spec struct {
	/*
		"file" defines the yaml file path to interact with.

		compatible:
			* source
			* condition
			* target

		remark:
			* "file" and "files" are mutually exclusive
			* when used as a source or condition, the file path also accept the following protocols
			* protocols "https://", "http://", and "file://" are supported in path for source and condition
	*/
	File string `yaml:",omitempty"`
	/*
		"files" defines the list of yaml files path to interact with.

		compatible:
			* condition
			* target

		remark:
			* file and files are mutually exclusive
			* protocols "https://", "http://", and "file://" are supported in file path for source and condition
	*/
	Files []string `yaml:",omitempty"`
	/*
		"key" defines the yaml keypath.

		compatible:
			* source
			* condition
			* target

		remark:
			* key is a simpler version of yamlpath accepts keys.

		example:
			* key: $.name
			* key: $.agent.name
			* key: $.agents[0].name
			* key: $.agents[*].name
			* key: $.'agents.name'

		remark:
			field path with key/value is not supported at the moment.
			some help would be useful on https://github.com/goccy/go-yaml/issues/290

	*/
	Key string `yaml:",omitempty"`
	/*
		"value" is the value associated with a yaml key.

		compatible:
			* source
			* condition
			* target

		default:
			When used from a condition or a target, the default value is set to linked source output.
	*/
	Value string `yaml:",omitempty"`
	/*
		"keyonly" allows to only check if a key exist and do not return an error otherwise

		compatible:
			* condition

		default:
			false
	*/
	KeyOnly bool `yaml:",omitempty"`
}

func (s Spec) Atomic() Spec {
	return Spec{
		File:  s.File,
		Files: s.Files,
		Key:   s.Key,
		Value: s.Value,
	}
}

// Validate validates the object and returns an error (with all the failed validation messages) if it is not valid
func (s *Spec) Validate() error {
	var validationErrors []string

	// Check for all validation
	if len(s.Files) == 0 && s.File == "" {
		validationErrors = append(validationErrors, "Invalid spec for yaml resource: both 'file' and 'files' are empty.")
	}
	if s.Key == "" {
		validationErrors = append(validationErrors, "Invalid spec for yaml resource: 'key' is empty.")
	}
	if len(s.Files) > 0 && s.File != "" {
		validationErrors = append(validationErrors, "Validation error in target of type 'yaml': the attributes `spec.file` and `spec.files` are mutually exclusive")
	}
	if len(s.Files) > 1 && hasDuplicates(s.Files) {
		validationErrors = append(validationErrors, "Validation error in target of type 'yaml': the attributes `spec.files` contains duplicated values")
	}

	// Return all the validation errors if found any
	if len(validationErrors) > 0 {
		return fmt.Errorf("validation error: the provided manifest configuration had the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	return nil
}
