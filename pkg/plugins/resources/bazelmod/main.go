package bazelmod

import (
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
)

// Bazelmod stores configuration about the MODULE.bazel file and the module to update
type Bazelmod struct {
	spec             Spec
	contentRetriever text.TextRetriever
}

// New returns a reference to a newly initialized Bazelmod object from a Spec
// or an error if the provided Spec triggers a validation error.
func New(spec interface{}) (*Bazelmod, error) {
	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	err = newSpec.Validate()
	if err != nil {
		return nil, err
	}

	newSpec.File = strings.TrimPrefix(newSpec.File, "file://")

	b := Bazelmod{
		spec:             newSpec,
		contentRetriever: &text.Text{},
	}

	return &b, nil
}

// Validate tests that the Bazelmod struct is correctly configured
func (b *Bazelmod) Validate() error {
	return b.spec.Validate()
}

// Changelog returns the changelog for this resource, or nil if not supported
func (b *Bazelmod) Changelog(from, to string) *result.Changelogs {
	return nil
}

// ReportConfig returns a new configuration object with only the necessary fields
// to identify the resource without any sensitive information or context specific data.
func (b *Bazelmod) ReportConfig() interface{} {
	return Spec{
		File:   b.spec.File,
		Module: b.spec.Module,
	}
}
