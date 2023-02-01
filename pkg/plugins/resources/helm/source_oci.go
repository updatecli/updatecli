package helm

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// OCISource return a Helm Chart version hosted on a OCI registry
func (c *Chart) OCISource(workingDir string) (string, error) {

	refName := filepath.Join(strings.TrimPrefix(c.spec.URL, "oci://"), c.spec.Name)

	repo, err := name.NewRepository(refName)
	if err != nil {
		return "", fmt.Errorf("invalid OCI Helm chart %s: %w", refName, err)
	}

	logrus.Debugf(
		"Searching versions for Helm chart %q",
		repo,
	)

	versions, err := remote.List(repo, c.options...)
	if err != nil {
		return "", fmt.Errorf("unable to list versions for OCI Helm chart %s: %w", repo, err)
	}

	c.foundVersion, err = c.versionFilter.Search(versions)
	if err != nil {
		return "", err
	}
	version := c.foundVersion.GetVersion()

	if len(version) == 0 {
		logrus.Infof("%s No OCI Helm chart version found matching pattern %q", result.FAILURE, c.versionFilter.Pattern)
		return version, fmt.Errorf("no OCI Helm chart version found matching pattern %q", c.versionFilter.Pattern)
	} else if len(version) > 0 {
		logrus.Infof("%s OCI version %q found matching pattern %q", result.SUCCESS, version, c.versionFilter.Pattern)
	} else {
		logrus.Errorf("Something unexpected happened in OCI Helm chart source")
	}

	return version, nil
}
