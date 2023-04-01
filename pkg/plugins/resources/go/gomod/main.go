package gomod

import (
	"github.com/mitchellh/mapstructure"
)

// GoMod defines a resource of type "go language"
type GoMod struct {
	spec     Spec
	filename string
}

// New returns a reference to a newly initialized Go Module object from a godmodule.Spec
// or an error if the provided Spec triggers a validation error.
func New(spec interface{}) (*GoMod, error) {

	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	filename := "go.mod"
	if newSpec.File != "" {
		filename = newSpec.File
	}

	return &GoMod{
		spec:     newSpec,
		filename: filename,
	}, nil
}
