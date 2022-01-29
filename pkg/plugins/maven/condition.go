package maven

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition tests if a specific version exist on the maven repository
func (m *Maven) Condition(source string) (bool, error) {
	if m.spec.Version == "" {
		m.spec.Version = source
	}

	versions, err := m.metadataHandler.GetVersions()
	if err != nil {
		return false, err
	}

	for _, version := range versions {
		if version == m.spec.Version {
			logrus.Infof("%s Version %s is available on the Maven Repository (%s)",
				result.SUCCESS, m.spec.Version, m.metadataHandler.GetMetadataURL())
			return true, nil
		}
	}

	logrus.Infof("%s Version %s is not available on the Maven Repository (%s)",
		result.FAILURE, m.spec.Version, m.metadataHandler.GetMetadataURL())
	return false, nil
}

// ConditionFromSCM returns an error because it's not supported
func (m *Maven) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return false, fmt.Errorf("SCM configuration is not supported for maven condition, aborting")
}
