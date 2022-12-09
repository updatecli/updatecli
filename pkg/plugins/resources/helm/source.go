package helm

import (
	"fmt"
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

	entriesVersion, found := index.Entries[c.spec.Name]
	if !found {
		return "", fmt.Errorf("helm chart %q not found from Helm Chart repository %q", c.spec.Name, c.spec.URL)
	}

	versions := []string{}

	for _, entry := range entriesVersion {
		versions = append(versions, entry.Version)
	}

	c.foundVersion, err = c.versionFilter.Search(versions)
	if err != nil {
		return "", err
	}
	value := c.foundVersion.GetVersion()

	if value != "" {
		logrus.Infof("%s Helm Chart '%s' version '%v' is found from repository %s",
			result.SUCCESS,
			c.spec.Name,
			value,
			c.spec.URL)
	}

	return value, nil

}
