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
	MINORVERSION = "minor"
	// MAJORVERSION defines major version identifier
	MAJORVERSION = "major"
	// PATCHVERSION defines patch version identifier
	PATCHVERSION = "patch"
	// NOINCREMENT defines if a chart version need to be updated
	NOINCREMENT = "none"
)

// Spec defines a specification for an "helmchart" resource
// parsed from an updatecli manifest file
type Spec struct {

	// Defines the Helm Chart file to update, only for "target"
	File string
	// Defines the key within the file that need to be updated, only for "target"
	Key string
	// Defines the Chart name path like "stable/chart", only for "target"
	Name string
	// Defines the chart location URL, only for source and condition
	URL string
	// Defines the value to set for a key, only for "target"
	Value string
	// Defines the Chart version, only for "source" and "condition", default value set based on sourceinput value
	Version string // [source][condition]
	// Defines if a Chart change, triggers a Chart version update, target only, accept values is a comma separated list of "none,major,minor,patch"
	VersionIncrement string
	// Defines if AppVersion must be updated as well, only for "target"
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
