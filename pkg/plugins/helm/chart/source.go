package chart

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// Source return the latest version
func (c *Chart) Source(workingDir string) (string, error) {

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
