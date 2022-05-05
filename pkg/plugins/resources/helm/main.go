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
)

// Spec defines a specification for an "helmchart" resource
// parsed from an updatecli manifest file
type Spec struct {
	File             string // [target] Define file to update
	Key              string // [target] Define Key to update
	Name             string // [source][condition][target] Define Chart name path like "stable/chart"
	URL              string // [source][condition] Define the chart location
	Value            string // [target] Define value to set
	Version          string // [source][condition]
	VersionIncrement string // [target] Define the rule to incremental the Chart version, accept a list of rules
	AppVersion       bool   // [target] Boolean that define we must update the App Version
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
