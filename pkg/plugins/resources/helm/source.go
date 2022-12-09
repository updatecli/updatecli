package helm

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"helm.sh/helm/v3/pkg/repo"
)

// Source return the latest version
func (c *Chart) Source(workingDir string) (string, error) {

	var index repo.IndexFile
	var err error

	if strings.HasPrefix(c.spec.URL, "https://") || strings.HasPrefix(c.spec.URL, "http://") {
		index, err = c.GetRepoIndexFromURL()
		if err != nil {
			return "", err
		}
	} else {
		index, err = c.GetRepoIndexFromFile(workingDir)
		if err != nil {
			return "", err
		}
	}

	if err != nil {
		return "", err
	}

	e, err := index.Get(c.spec.Name, c.spec.Version)

	if err != nil {
		return "", err
	}

	if e.Version != "" {
		logrus.Infof("%s Helm Chart '%s' version '%v' is found from repository %s",
			result.SUCCESS,
			c.spec.Name,
			e.Version,
			c.spec.URL)
	}

	return e.Version, nil

}
