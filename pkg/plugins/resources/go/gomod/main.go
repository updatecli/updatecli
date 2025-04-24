package gomod

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/core/text"
)

// GoMod defines a resource of type "go language"
type GoMod struct {
	spec             Spec
	filename         string
	kind             string
	foundVersion     string
	contentRetriever text.TextRetriever
	currentContent   string
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
		filename = strings.TrimPrefix(newSpec.File, "file://")
	}

	kind := kindModule
	if newSpec.Module == "" {
		kind = kindGolang

	}

	return &GoMod{
		spec:             newSpec,
		filename:         filename,
		kind:             kind,
		contentRetriever: &text.Text{},
	}, nil
}

// Read reads the file content
func (g *GoMod) Read(filename string) error {
	textContent, err := g.contentRetriever.ReadAll(filename)
	if err != nil {
		return err
	}
	g.currentContent = textContent
	return nil
}

// CleanConfig returns a new configuration for this resource without any sensitive information or context specific information.
func (g *GoMod) CleanConfig() interface{} {
	return Spec{
		Module:  g.spec.Module,
		File:    g.spec.File,
		Version: g.spec.Version,
	}
}
