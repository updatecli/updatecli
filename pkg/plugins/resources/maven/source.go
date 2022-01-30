package maven

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source return the latest version
func (m *Maven) Source(workingDir string) (string, error) {

	latestVersion, err := m.metadataHandler.GetLatestVersion()
	if err != nil {
		return "", err
	}

	if latestVersion != "" {
		logrus.Infof(
			"%s Latest version is %s on the Maven repository at %s",
			result.SUCCESS,
			latestVersion,
			m.metadataHandler.GetMetadataURL(),
		)
		return latestVersion, nil
	}

	return "", fmt.Errorf("%s No latest version on the Maven Repository at %s", result.FAILURE, m.metadataHandler.GetMetadataURL())
}
