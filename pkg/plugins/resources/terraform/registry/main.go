package registry

import (
	"net/http"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

type TerraformRegistry struct {
	Spec Spec
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter   version.Filter
	Version         version.Version
	scm             string // Source control URL from api
	webClient       httpclient.HTTPClient
	registryAddress registryAddress
}

func New(spec interface{}) (*TerraformRegistry, error) {
	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	newFilter := newSpec.VersionFilter
	if newFilter.IsZero() {
		// By default, use semantic versioning
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	webClient := &http.Client{}

	registryAddress, err := newRegistryAddress(webClient, newSpec)
	if err != nil {
		return nil, err
	}

	return &TerraformRegistry{
		Spec:            newSpec,
		versionFilter:   newFilter,
		webClient:       webClient,
		registryAddress: registryAddress,
	}, nil
}
