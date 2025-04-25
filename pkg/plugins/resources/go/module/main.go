package gomodule

import (
	"net/http"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
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
	Spec Spec
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	Version       version.Version
	webClient     httpclient.HTTPClient
}

// New returns a reference to a newly initialized Go Module object from a godmodule.Spec
// or an error if the provided Spec triggers a validation error.
func New(spec interface{}) (*GoModule, error) {
	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	newFilter := newSpec.VersionFilter
	if newSpec.VersionFilter.IsZero() {
		logrus.Debugln("no versioning filtering specified, fallback to semantic versioning")
		// By default, golang versioning uses semantic versioning
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	return &GoModule{
		Spec:          newSpec,
		versionFilter: newFilter,
		webClient:     &http.Client{},
	}, nil
}

// ReportConfig returns a new configuration without any sensitive information or context specific information.
func (g *GoModule) ReportConfig() interface{} {
	return Spec{
		Proxy:         redact.URL(g.Spec.Proxy),
		Module:        g.Spec.Module,
		Version:       g.Spec.Version,
		VersionFilter: g.Spec.VersionFilter,
	}
}
