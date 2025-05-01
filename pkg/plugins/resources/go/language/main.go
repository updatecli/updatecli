package language

import (
	"net/http"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Language defines a resource of type "go language"
type Language struct {
	Spec Spec
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	Version       version.Version
	webClient     httpclient.HTTPClient
}

// New returns a reference to a newly initialized Go Module object from a godmodule.Spec
// or an error if the provided Spec triggers a validation error.
func New(spec interface{}) (*Language, error) {

	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	newFilter := newSpec.VersionFilter
	if newFilter.IsZero() {
		// By default, golang versioning uses semantic versioning
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	return &Language{
		Spec:          newSpec,
		versionFilter: newFilter,
		webClient:     &http.Client{},
	}, nil
}

// ReportConfig returns a new configuration without any sensitive information or context specific information.
func (l *Language) ReportConfig() interface{} {
	return Spec{
		Version:       l.Spec.Version,
		VersionFilter: l.Spec.VersionFilter,
	}
}
