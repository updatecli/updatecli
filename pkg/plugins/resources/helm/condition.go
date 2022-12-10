package helm

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/repo"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition check if a specific chart version exist
func (c *Chart) Condition(source string) (bool, error) {
	return c.ConditionFromSCM(source, nil)
}

// ConditionFromSCM returns an error because it's not supported
func (c *Chart) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {

	if strings.HasPrefix(c.spec.URL, "oci://") {
		return c.OCICondition(source)
	}

	if c.spec.Version != "" {
		logrus.Infof("Version %v, already defined from configuration file", c.spec.Version)
	} else {
		c.spec.Version = source
	}

	var index repo.IndexFile
	var err error

	if strings.HasPrefix(c.spec.URL, "https://") || strings.HasPrefix(c.spec.URL, "http://") {
		index, err = c.GetRepoIndexFromURL()
		if err != nil {
			return false, err
		}
	} else {
		rootDir := ""
		if scm != nil {
			rootDir = scm.GetDirectory()
		}
		index, err = c.GetRepoIndexFromFile(rootDir)
		if err != nil {
			return false, err
		}
	}

	message := ""
	if c.spec.Version != "" {
		message = fmt.Sprintf(" for version '%s'", c.spec.Version)
	}

	if index.Has(c.spec.Name, c.spec.Version) {
		logrus.Infof("%s Helm Chart '%s' is available on %s%s", result.SUCCESS, c.spec.Name, c.spec.URL, message)
		return true, nil
	}

	logrus.Infof("%s Helm Chart '%s' isn't available on %s%s", result.FAILURE, c.spec.Name, c.spec.URL, message)
	return false, nil

}
