package helm

import (
	"github.com/mitchellh/mapstructure"
)

const (
	// CHANGELOGTEMPLATE contains helm chart changelog information
	CHANGELOGTEMPLATE string = `
Remark: We couldn't identify a way to automatically retrieve changelog information.
Please use following information to take informed decision

{{ if .Name }}Helm Chart: {{ .Name }}{{ end }}
{{ if .Description }}{{ .Description }}{{ end }}
{{ if .Home }}Project Home: {{ .Home }}{{ end }}
{{ if .KubeVersion }}Require Kubernetes Version: {{ .KubeVersion }}{{end}}
{{ if .Created }}Version created on the {{ .Created }}{{ end}}
{{ if .Sources }}
Sources:
{{ range $index, $source := .Sources }}
* {{ $source }}
{{ end }}
{{ end }}
{{ if .URLs }}
URL:
{{ range $index, $url := .URLs }}
* {{ $url }}
{{ end }}
{{ end }}
`
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
	File string
	// [target] Defines the key within the file that need to be updated.
	Key string
	// [target] Defines the Chart name path like "stable/chart".
	Name string
	// [source,condition] Defines the chart location URL.
	URL string
	// [target] Defines the value to set for a key
	Value string
	// [source,condition] Defines the Chart version, default value set based on sourceinput value
	Version string
	// [target] Defines if a Chart changes, triggers (or not) a Chart version update, accepted values is a comma separated list of "none,major,minor,patch"
	VersionIncrement string
	// [target] Defines if AppVersion must be updated as well
	AppVersion bool
}

// Chart defines a resource of kind "helmchart"
type Chart struct {
	spec Spec
}

// New returns a reference to a newly initialized Chart object from a Spec
// or an error if the provided YamlSpec triggers a validation error.
func New(spec interface{}) (*Chart, error) {
	newSpec := Spec{}
	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return &Chart{}, err
	}

	newResource := &Chart{
		spec: newSpec,
	}

	return newResource, nil

}
