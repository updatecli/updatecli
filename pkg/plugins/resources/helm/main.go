package helm

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

const (
	// AUTO defines automatic version identifier
	AUTO string = "auto"
	// MINORVERSION defines minor version identifier
	MINORVERSION string = "minor"
	// MAJORVERSION defines major version identifier
	MAJORVERSION string = "major"
	// PATCHVERSION defines patch version identifier
	PATCHVERSION string = "patch"
	// NOINCREMENT disables chart version auto increment
	NOINCREMENT string = "none"
)

/*
Spec defines a specification for an "helmchart" resource
parsed from an updatecli manifest file
*/
type Spec struct {
	/*
		file defines the Helm Chart file to update.
		the path must be relative to chart root directory
		the chart name is defined by the parameter "name"

		compatible:
			* source
			* condition
			* target

		default:
			default set to "values.yaml"
	*/
	File string `yaml:",omitempty"`
	/*
		key defines the yamlpath query used for retrieving value from a YAML document

		compatible:
			* target

		example:
			* key: $.image.tag
			* key: $.images[0].tag

		remark:
			* key is a simpler version of yamlpath accepts keys.

	*/
	Key string `yaml:",omitempty"`
	/*
		name defines the Chart name path like 'stable/chart'.

		compatible:
			* source
			* condition
			* target

		example:
			* name: stable/chart

		remark:
			* when used with a scm, the name must be the relative path from the scm repository root directory
			  with such as "stable/chart"
	*/
	Name string `yaml:",omitempty"`
	/*
		skippackaging defines if a Chart should be packaged or not.

		compatible:
			* target

		default: false
	*/
	SkipPackaging bool `yaml:",omitempty"`
	/*
		url defines the Chart location URL.

		compatible:
			* source
			* condition

		example:
			* index.yaml
			* file://./index.yaml
			* https://github.com/updatecli/charts.git
			* oci://ghcr.io/olblak/charts/

	*/
	URL string `yaml:",omitempty"`
	/*
		value is the value associated with a yamlpath query.

		compatible:
			* condition
			* target
	*/
	Value string `yaml:",omitempty"`
	/*
		version defines the Chart version. It is used by condition to check if a version exists on the registry.

		compatible:
			* condition
	*/
	Version string `yaml:",omitempty"`
	/*
		versionIncrement defines if a Chart changes, triggers, or not, a Chart version update, accepted values is a comma separated list of "none,major,minor,patch,auto".

		compatible:
			* target

		default:
			default set to "minor"

		remark:
			when multiple pipelines update the same chart, the versionIncrement will be applied multiple times.
			more information on https://github.com/updatecli/updatecli/issues/693
	*/
	VersionIncrement string `yaml:",omitempty"`
	/*
		AppVersion defines if a Chart changes, triggers, or not, a Chart AppVersion update.
		The value is retrieved from the source input.

		compatible:
			* target

		default
			false
	*/
	AppVersion bool `yaml:",omitempty"`
	/*
		versionfilter provides parameters to specify version pattern and its type like 'regex', 'semver', or just 'latest'.

		compatible:
			* source

		default:
			semver

		remark:
			* Helm chart uses semver by default.
	*/
	VersionFilter version.Filter `yaml:",omitempty"`
	/*
		credentials defines the credentials used to authenticate with OCI registries
	*/
	docker.InlineKeyChain `yaml:",inline" mapstructure:",squash"`
}

// Chart defines a resource of kind helmchart
type Chart struct {
	spec    Spec
	options []remote.Option
	// Holds both parsed version and original version, to allow retrieving metadata such as changelog
	foundVersion version.Version
	// Holds the "valid" version.filter, that might be different than the user-specified filter version filter
	versionFilter version.Filter
}

// New returns a reference to a newly initialized Chart object from a Spec
// or an error if the provided YamlSpec triggers a validation error.
func New(spec interface{}) (*Chart, error) {
	newSpec := Spec{}
	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return &Chart{}, err
	}

	if newSpec.VersionFilter.Kind == "" {
		newSpec.VersionFilter.Kind = "semver"
	}

	newFilter, err := newSpec.VersionFilter.Init()
	if err != nil {
		return &Chart{}, err
	}

	newResource := &Chart{
		spec:          newSpec,
		versionFilter: newFilter,
	}

	err = newSpec.Validate()
	if err != nil {
		return nil, err
	}

	keychains := []authn.Keychain{}

	if !newSpec.Empty() {
		keychains = append(keychains, newSpec.InlineKeyChain)
	}

	keychains = append(keychains, authn.DefaultKeychain)

	newResource.options = append(newResource.options, remote.WithAuthFromKeychain(authn.NewMultiKeychain(keychains...)))

	return newResource, nil
}

// ReportConfig returns a new configuration without any sensitive information
// or context specific information
func (c *Chart) ReportConfig() interface{} {
	return Spec{
		File:             c.spec.File,
		Key:              c.spec.Key,
		Name:             c.spec.Name,
		SkipPackaging:    c.spec.SkipPackaging,
		URL:              redact.URL(c.spec.URL),
		Value:            c.spec.Value,
		Version:          c.spec.Version,
		VersionIncrement: c.spec.VersionIncrement,
		AppVersion:       c.spec.AppVersion,
		VersionFilter:    c.spec.VersionFilter,
	}
}
