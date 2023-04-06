package gomod

import (
	"github.com/mitchellh/mapstructure"
)

// GoMod defines a resource of type "go language"
type GoMod struct {
	spec     Spec
	filename string
	kind     string
}

var (
	kindGolang string = "golang"
	kindModule string = "module"
)

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

	kind := kindModule
	if newSpec.Kind != "" {
		kind = newSpec.Kind
	}

	err = newSpec.Validate()
	if err != nil {
		return nil, err
	}

	return &GoMod{
		spec:     newSpec,
		filename: filename,
		kind:     kind,
	}, nil
}
