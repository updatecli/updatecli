package helm

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
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
"helmchart" defines the specification for manipulating Helm charts.
It can be used as a "source", a "condition", or a "target".
*/
type Spec struct {
	// "file" defines the chart file to update.
	//
	// compatible:
	//   * target
	//
	// default:
	//   values.yaml
	//
	// remark:
	//   * the path is relative to the chart root directory.
	//   * the chart directory is defined by "name".
	//
	File string `yaml:",omitempty"`
	// "key" defines the yamlpath query used to retrieve the value from the yaml file.
	//
	// compatible:
	//   * target
	//
	// remark:
	//   * "key" is required in a target.
	//   * "key" is a simpler version of yamlpath.
	//
	// example:
	//   * key: $.image.tag
	//   * key: $.images[0].tag
	//
	Key string `yaml:",omitempty"`
	// "name" defines the chart name, or the chart path such as "stable/chart".
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * in a target, "name" is the chart directory path. When used with an scm,
	//     it is relative to the scm repository root directory, such as "stable/chart".
	//
	// example:
	//   * name: stable/chart
	//
	Name string `yaml:",omitempty"`
	// "skippackaging" defines whether the chart dependencies update is skipped.
	//
	// compatible:
	//   * target
	//
	// default:
	//   false
	//
	SkipPackaging bool `yaml:",omitempty"`
	// "url" defines the chart repository location.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * the schemes "https://", "http://", "oci://" and "file://" are supported.
	//   * a url without scheme is read as a local path.
	//   * "index.yaml" is appended when the url does not end with it.
	//
	// example:
	//   * url: index.yaml
	//   * url: file://./index.yaml
	//   * url: https://github.com/updatecli/charts.git
	//   * url: oci://ghcr.io/olblak/charts/
	//
	URL string `yaml:",omitempty"`
	// "value" defines the value associated with the yamlpath query.
	//
	// compatible:
	//   * target
	//
	// default:
	//   the output of the associated source.
	//
	Value string `yaml:",omitempty"`
	// "version" defines the chart version to check on the registry.
	//
	// compatible:
	//   * condition
	//
	// default:
	//   the output of the associated source.
	//
	Version string `yaml:",omitempty"`
	// "versionincrement" defines how the chart version is bumped when the chart changes.
	//
	// compatible:
	//   * target
	//
	// default:
	//   minor
	//
	// remark:
	//   * accepted values are a comma separated list of "major", "minor" and "patch",
	//     or one of "auto" or "none" on its own.
	//   * "none" disables the chart version update.
	//   * "auto" bumps the part of the chart version matching the part that changed in the updated value.
	//   * when several pipelines update the same chart, the increment is applied several times.
	//     More information on https://github.com/updatecli/updatecli/issues/693
	//
	// example:
	//   * versionincrement: patch
	//   * versionincrement: major,minor
	//
	VersionIncrement string `yaml:",omitempty"`
	// "appversion" defines whether the chart "appVersion" is updated.
	//
	// compatible:
	//   * target
	//
	// default:
	//   false
	//
	// remark:
	//   * the value is retrieved from the source output.
	//   * the "appVersion" field is only updated when the chart metadata already holds it.
	//
	AppVersion bool `yaml:",omitempty"`
	// "versionfilter" defines the version pattern and its kind, such as "regex", "semver" or "latest".
	//
	// compatible:
	//   * source
	//
	// default:
	//   semver
	//
	VersionFilter version.Filter `yaml:",omitempty"`
	// InlineKeyChain defines the credentials used to authenticate with the chart repository.
	//
	// remark:
	//   * they are used with OCI registries.
	//   * "username" and "password" are also used for basic authentication with an http or https repository.
	//
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
	newResource.options = append(newResource.options, remote.WithTransport(httpclient.ProxyOnlyTransport()))

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
