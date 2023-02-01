package helm

import (
	"bytes"
	"html/template"

	"github.com/sirupsen/logrus"
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
)

// Changelog returns a rendered template with this chart version information
func (c Chart) Changelog() string {
	index, err := c.GetRepoIndexFromURL()

	if err != nil {
		return ""
	}

	e, err := index.Get(c.spec.Name, c.spec.Version)
	if err != nil {
		return ""
	}

	t := template.Must(template.New("changelog").Parse(CHANGELOGTEMPLATE))

	buffer := new(bytes.Buffer)

	type params struct {
		Name        string
		Description string
		Home        string
		KubeVersion string
		Created     string
		URLs        []string `json:"url"`
		Sources     []string
	}

	err = t.Execute(buffer, params{
		Name:        e.Name,
		Description: e.Description,
		Home:        e.Home,
		KubeVersion: e.KubeVersion,
		Created:     e.Created.String(),
		URLs:        e.URLs,
		Sources:     e.Sources})

	if err != nil {
		return ""
	}

	changelog := buffer.String()

	logrus.Infof(changelog)

	return changelog

}
