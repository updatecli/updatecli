package language

import (
	"net/http"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Language defines a resource of type "go language"
type Language struct {
	spec Spec
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
		spec:          newSpec,
		versionFilter: newFilter,
		webClient:     http.DefaultClient,
	}, nil
}
