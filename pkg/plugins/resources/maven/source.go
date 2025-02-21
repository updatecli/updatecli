package maven

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
)

// Source return the latest version
func (m *Maven) Source(workingDir string, resultSource *result.Source) error {

	for _, metadataHandler := range m.metadataHandlers {
		// metadataURL contains the URL without username/password
		metadataURL, err := trimUsernamePasswordFromURL(metadataHandler.GetMetadataURL())
		if err != nil {
			logrus.Errorf("Trying to parse Maven metadata url: %s", err)
		}

		latestVersion, err := metadataHandler.GetLatestVersion()
		if err != nil {
			logrus.Warnf("getting latest version: %v", err)
		}

		if latestVersion != "" {
			resultSource.Result = result.SUCCESS
			resultSource.Information = latestVersion
			resultSource.Description = fmt.Sprintf(
				"Latest version is %s on the Maven repository at %s",
				latestVersion,
				redact.URL(metadataURL),
			)
			return nil
		}

	}

	return fmt.Errorf("no latest version for the Maven Artifact %s/%s", m.spec.GroupID, m.spec.ArtifactID)
}
