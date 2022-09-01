package maven

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source return the latest version
func (m *Maven) Source(workingDir string) (string, error) {

	for _, metadataHandler := range m.metadataHandlers {
		// metadataURL contains the URL without username/password
		metadataURL, err := trimUsernamePasswordFromURL(metadataHandler.GetMetadataURL())
		if err != nil {
			logrus.Errorf("Trying to parse Maven metadatal url: %s", err)
		}

		latestVersion, err := metadataHandler.GetLatestVersion()
		if err != nil {
			return "", err
		}

		if latestVersion != "" {
			logrus.Infof(
				"%s Latest version is %s on the Maven repository at %s",
				result.SUCCESS,
				latestVersion,
				metadataURL,
			)
			return latestVersion, nil
		}

	}

	return "", fmt.Errorf("%s No latest version for the Maven Artifact %s/%s", result.FAILURE, m.spec.GroupID, m.spec.ArtifactID)
}
