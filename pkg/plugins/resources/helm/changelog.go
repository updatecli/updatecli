package helm

import (
	"bytes"
	"html/template"

	"github.com/sirupsen/logrus"
)

// Changelog returns a rendered template with this chart version informations
func (c Chart) Changelog() string {
	index, err := c.GetRepoIndexFile()

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
