package gomodule

import (
	"net/http"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

/*
	https://go.dev/ref/mod#goproxy-protocol
*/

const (
	goModuleDefaultProxy string = "https://proxy.golang.org"
)

// GoModule defines a resource of type "gomodule"
type GoModule struct {
	spec Spec
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	foundVersion  version.Version
	webClient     httpclient.HTTPClient
}

// New returns a reference to a newly initialized Go Module object from a godmodule.Spec
// or an error if the provided Spec triggers a validation error.
func New(spec interface{}, isSCM bool) (*GoModule, error) {
	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	newFilter, err := newSpec.VersionFilter.Init()
	if err != nil {
		return nil, err
	}

	return &GoModule{
		spec:          newSpec,
		versionFilter: newFilter,
		webClient:     http.DefaultClient,
	}, nil
}
