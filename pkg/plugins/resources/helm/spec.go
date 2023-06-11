package helm

import (
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for an "helmchart" resource
// parsed from an updatecli manifest file
type Spec struct {
	// [target] Defines the Helm Chart file to update.
	File string `yaml:",omitempty"`
	// [target] Defines the key to update within the file.
	Key string `yaml:",omitempty"`
	// [target] Defines the Chart name path like 'stable/chart'.
	Name string `yaml:",omitempty"`
	// [source,condition] Defines the chart location URL.
	URL string `yaml:",omitempty"`
	// [target] Defines the value to set for a key
	Value string `yaml:",omitempty"`
	// [source,condition] Defines the Chart version, default value set based on a source input value
	Version string `yaml:",omitempty"`
	// [target] Defines if a Chart changes, triggers, or not, a Chart version update, accepted values is a comma separated list of "none,major,minor,patch"
	VersionIncrement string `yaml:",omitempty"`
	// [target] Enable AppVersion update based in source input.
	AppVersion bool `yaml:",omitempty"`
	// VersionFilter provides parameters to specify version pattern and its type like 'regex', 'semver', or just 'latest'.
	VersionFilter version.Filter `yaml:",omitempty"`
	// Credentials used to authenticate with OCI registries
	docker.InlineKeyChain `yaml:",inline" mapstructure:",squash"`
}

func (s Spec) Atomic() Spec {
	return Spec{
		File:    s.File,
		Key:     s.Key,
		Name:    s.Name,
		Value:   s.Value,
		Version: s.Version,
		URL:     s.URL,
	}
}
