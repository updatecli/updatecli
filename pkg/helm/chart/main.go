package chart

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/olblak/updateCli/pkg/scm"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"
)

const (
	CHANGELOGTEMPLATE string = `
{{ if .Name }}Helm Chart: {{ .Name }}{{ end }}
{{ if .Description }}{{ .Description }}{{ end }}
{{ if .Home }}Project Home: {{ .Home }}{{ end }}
{{ if .KubeVersion }}Require Kubernetes Version: {{ .KubeVersion }}{{end}}
{{ if .Created }}Version created on the {{ .Created }}{{ end}}
{{ if .URL }}
Various URL:
	{{ .URL }}
{{ end }}
`
)

// Chart describe helm repository metadata
type Chart struct {
	URL     string
	Name    string
	Version string
}

// Changelog return any information available for a helm chart
func (c *Chart) Changelog(name string) (string, error) {
	URL := fmt.Sprintf("%s/index.yaml", c.URL)

	req, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		return "", err
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	index, err := loadIndex(body)

	if err != nil {
		return "", err
	}

	e, err := index.Get(c.Name, c.Version)

	if err != nil {
		return "", err
	}

	t := &template.Template{}
	t = template.Must(template.New("changelog").Parse(CHANGELOGTEMPLATE))
	buffer := new(bytes.Buffer)

	type params struct {
		Name        string
		Description string
		Home        string
		KubeVersion string
		Created     string
		URL         []string
		SOURCES     []string
	}

	err = t.Execute(buffer, params{
		Name:        e.Name,
		Description: e.Description,
		Home:        e.Home,
		KubeVersion: e.KubeVersion,
		Created:     e.Created.String(),
		URL:         e.URLs,
		SOURCES:     e.Sources})

	if err != nil {
		return "", err
	}

	changelog := buffer.String()

	return changelog, nil

}

// Condition check if a specific chart version exist
func (c *Chart) Condition(source string) (bool, error) {

	if c.Version != "" {
		fmt.Printf("Version %v, already defined from configuration file\n", c.Version)
	} else {
		c.Version = source
	}
	URL := fmt.Sprintf("%s/index.yaml", c.URL)

	req, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		return false, err
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return false, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return false, err
	}

	index, err := loadIndex(body)

	if err != nil {
		return false, err
	}

	message := ""
	if c.Version != "" {
		message = fmt.Sprintf(" for version '%s'", c.Version)
	}

	if index.Has(c.Name, c.Version) {
		fmt.Printf("\u2714 Helm Chart '%s' is available on %s%s\n", c.Name, c.URL, message)
		return true, nil
	}

	fmt.Printf("\u2717 Helm Chart '%s' isn't available on %s%s\n", c.Name, c.URL, message)
	return false, nil

}

// ConditionFromSCM returns an error because it's not supported
func (c *Chart) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {
	return false, fmt.Errorf("SCM configuration is not supported for Helm chart condition, aborting")
}

// Source return the latest version
func (c *Chart) Source() (string, error) {
	URL := fmt.Sprintf("%s/index.yaml", c.URL)

	req, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		return "", err
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	index, err := loadIndex(body)

	if err != nil {
		return "", err
	}

	e, err := index.Get(c.Name, c.Version)

	if err != nil {
		return "", err
	}

	if e.Version != "" {
		fmt.Printf("\u2714 Helm Chart '%s' version '%v' is founded from repository %s\n",
			c.Name,
			e.Version,
			c.URL)
	}

	return e.Version, nil

}

// loadIndex loads an index file and does minimal validity checking.
//
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
