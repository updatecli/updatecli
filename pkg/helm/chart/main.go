package chart

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"
)

// Chart describe helm repository metadata
type Chart struct {
	URL     string
	Name    string
	Version string
}

// Condition check if a specific chart version exist
func (c *Chart) Condition() (bool, error) {
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
