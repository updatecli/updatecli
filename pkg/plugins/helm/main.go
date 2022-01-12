package helm

import (
	"helm.sh/helm/v3/pkg/repo"
	yml "sigs.k8s.io/yaml"
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

// Chart describe helm repository metadata
type Chart struct {
	spec Spec
}

// loadIndex loads an index file and does minimal validity checking.
// This will fail if API Version is not set (ErrNoAPIVersion) or if the unmarshal fails.
func loadIndex(data []byte) (repo.IndexFile, error) {
	i := repo.IndexFile{}

	if err := yml.Unmarshal(data, &i); err != nil {
		return i, err
	}

	i.SortEntries()

	if i.APIVersion == "" {
		return i, repo.ErrNoAPIVersion
	}

	return i, nil
}
