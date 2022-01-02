package helmchart

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
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
		logrus.Infof("%s Helm Chart '%s' version '%v' is found from repository %s",
			result.SUCCESS,
			c.Name,
			e.Version,
			c.URL)
	}

	return e.Version, nil

}
