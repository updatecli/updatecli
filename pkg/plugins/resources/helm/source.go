package helm

import (
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source return the latest version
func (c *Chart) Source(workingDir string) (string, error) {

	index, err := c.GetRepoIndexFile()

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
