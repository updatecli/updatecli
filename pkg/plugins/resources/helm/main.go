package helm

import (
	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

const (
	// MINORVERSION defines minor version identifier
	MINORVERSION string = "minor"
	// MAJORVERSION defines major version identifier
	MAJORVERSION string = "major"
	// PATCHVERSION defines patch version identifier
	PATCHVERSION string = "patch"
	// NOINCREMENT defines if a chart version doesn't need to be incremented
	NOINCREMENT string = "none"
)

// Spec defines a specification for an "helmchart" resource
// parsed from an updatecli manifest file
type Spec struct {
	// [target] Defines the Helm Chart file to update.
	File string `yaml:",omitempty"`
	// [target] Defines the key within the file that need to be updated.
	Key string `yaml:",omitempty"`
	// [target] Defines the Chart name path like "stable/chart".
	Name string `yaml:",omitempty"`
	// [source,condition] Defines the chart location URL.
	URL string `yaml:",omitempty"`
	// [target] Defines the value to set for a key
	Value string `yaml:",omitempty"`
	// [source,condition] Defines the Chart version, default value set based on sourceinput value
	Version string `yaml:",omitempty"`
	// [target] Defines if a Chart changes, triggers (or not) a Chart version update, accepted values is a comma separated list of "none,major,minor,patch"
	VersionIncrement string `yaml:",omitempty"`
	// [target] Defines if AppVersion must be updated as well
	AppVersion bool `yaml:",omitempty"`
	// VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
}

// Chart defines a resource of kind "helmchart"
type Chart struct {
	spec Spec
	// Holds both parsed version and original version (to allow retrieving metadata such as changelog)
	foundVersion version.Version
	// Holds the "valid" version.filter, that might be different than the user-specified filter (Spec.VersionFilter)
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

	newFilter, err := newSpec.VersionFilter.Init()
	if err != nil {
		return &Chart{}, err
	}

	newResource := &Chart{
		spec:          newSpec,
		versionFilter: newFilter,
	}

	return newResource, nil

}
