package helm

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

// Changelog returns a rendered template with the informations of this version of the chart
func (c Chart) Changelog() string {
	URL := fmt.Sprintf("%s/index.yaml", c.spec.URL)

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return ""
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ""
	}

	index, err := loadIndex(body)
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
