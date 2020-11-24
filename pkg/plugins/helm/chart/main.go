package chart

import (
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"
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
	- {{ $source }}
{{ end }}
{{ end }}
{{ if .URLs }}
URL:
{{ range $index, $url := .URLs }}
	- {{ $url }}
{{ end }}
{{ end }}
`
)

// Chart describe helm repository metadata
type Chart struct {
	URL     string
	Name    string
	Version string
}

// loadIndex loads an index file and does minimal validity checking.
// This will fail if API Version is not set (ErrNoAPIVersion) or if the unmarshal fails.
func loadIndex(data []byte) (repo.IndexFile, error) {
	i := repo.IndexFile{}

	if err := yaml.Unmarshal(data, &i); err != nil {
		return i, err
	}

	i.SortEntries()

	if i.APIVersion == "" {
		return i, repo.ErrNoAPIVersion
	}

	return i, nil
}
