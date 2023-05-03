package helm

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/repo"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition checks if a specific chart version exist
func (c *Chart) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {

	if strings.HasPrefix(c.spec.URL, "oci://") {
		return c.OCICondition(source, scm, resultCondition)
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
			return err
		}
	} else {
		rootDir := ""
		if scm != nil {
			rootDir = scm.GetDirectory()
		}
		index, err = c.GetRepoIndexFromFile(rootDir)
		if err != nil {
			return err
		}
	}

	message := ""
	if c.spec.Version != "" {
		message = fmt.Sprintf(" for version '%s'", c.spec.Version)
	}

	if index.Has(c.spec.Name, c.spec.Version) {
		resultCondition.Pass = true
		resultCondition.Result = result.SUCCESS
		resultCondition.Description = fmt.Sprintf("Helm Chart %q is available on %s%s", c.spec.Name, c.spec.URL, message)
		return nil
	}

	return fmt.Errorf("the Helm chart %q isn't available on %s%s", c.spec.Name, c.spec.URL, message)

}
