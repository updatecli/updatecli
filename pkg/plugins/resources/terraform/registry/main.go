package registry

import (
	"net/http"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
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

// CleanConfig returns a new configuration object with only the necessary fields
// to identify the resource without any sensitive information or context specific data.
func (t *TerraformRegistry) CleanConfig() interface{} {
	return Spec{
		Type:          t.Spec.Type,
		Hostname:      redact.URL(t.Spec.Hostname),
		Namespace:     t.Spec.Namespace,
		Name:          t.Spec.Name,
		TargetSystem:  t.Spec.TargetSystem,
		RawString:     t.Spec.RawString,
		Version:       t.Spec.Version,
		VersionFilter: t.Spec.VersionFilter,
	}
}
