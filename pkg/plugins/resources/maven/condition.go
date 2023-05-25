package maven

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition tests if a specific version exist on the maven repository
func (m *Maven) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {

	if scm != nil {
		logrus.Warningf("SCM configuration is not supported for maven condition, aborting")
	}

	if m.spec.Version == "" {
		m.spec.Version = source
	}

	for _, metadataHandler := range m.metadataHandlers {
		// metadataURL contains the URL without username/password
		metadataURL, err := trimUsernamePasswordFromURL(metadataHandler.GetMetadataURL())
		if err != nil {
			return fmt.Errorf("trying to parse Maven metadata url: %s", err)
		}

		versions, err := metadataHandler.GetVersions()
		if err != nil {
			return err
		}

		for _, version := range versions {
			if version == m.spec.Version {
				resultCondition.Pass = true
				resultCondition.Result = result.SUCCESS
				resultCondition.Description = fmt.Sprintf("Version %s is available on the Maven Repository (%s)",
					m.spec.Version, metadataURL)
				return nil
			}
		}
	}

	resultCondition.Result = result.FAILURE
	resultCondition.Pass = false
	resultCondition.Description = fmt.Sprintf("Version %s is not found for Maven artifact (%s/%s)",
		m.spec.Version, m.spec.GroupID, m.spec.ArtifactID)
	return nil
}
