package helm

import (
	"bytes"
	"html/template"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
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
func (c Chart) Changelog(from, to string) *result.Changelogs {
	index, err := c.GetRepoIndexFromURL()

	if err != nil {
		logrus.Debugf("failed to get helm repository index: %s", err)
		return nil
	}

	e, err := index.Get(c.spec.Name, c.foundVersion.OriginalVersion)
	if err != nil {
		logrus.Debugf("failed to get helm chart information: %s", err)
		return nil
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
		logrus.Debugf("failed to render helm chart information: %s", err)
		return nil
	}

	changelog := buffer.String()

	return &result.Changelogs{
		{
			Title:       from,
			Body:        changelog,
			PublishedAt: e.Created.String(),
		},
	}

}
