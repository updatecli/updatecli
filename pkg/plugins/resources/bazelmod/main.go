package bazelmod

import (
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Bazelmod stores configuration about the MODULE.bazel file and the module to update
type Bazelmod struct {
	spec             Spec
	contentRetriever text.TextRetriever
	// Holds both parsed version and original version (to allow retrieving metadata such as changelog)
	foundVersion version.Version
	// Holds the "valid" version.filter, that might be different than the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
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

	newFilter, err := newSpec.VersionFilter.Init()
	if err != nil {
		return nil, err
	}

	b := Bazelmod{
		spec:             newSpec,
		versionFilter:    newFilter,
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
		File:          b.spec.File,
		Module:        b.spec.Module,
		VersionFilter: b.spec.VersionFilter,
	}
}
