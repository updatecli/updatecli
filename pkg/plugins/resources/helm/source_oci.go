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
func (c *Chart) OCISource(workingDir string, resultSource *result.Source) error {

	refName := filepath.Join(strings.TrimPrefix(c.spec.URL, "oci://"), c.spec.Name)

	repo, err := name.NewRepository(refName)
	if err != nil {
		return fmt.Errorf("invalid OCI Helm chart %s: %w", refName, err)
	}

	logrus.Debugf("Searching versions for Helm chart %q", repo)

	versions, err := remote.List(repo, c.options...)
	if err != nil {
		return fmt.Errorf("unable to list versions for OCI Helm chart %s: %w", repo, err)
	}

	c.foundVersion, err = c.versionFilter.Search(versions)
	if err != nil {
		return fmt.Errorf("filtering OCI helm chart version: %w", err)
	}
	version := c.foundVersion.GetVersion()

	if len(version) == 0 {
		return fmt.Errorf("no OCI Helm chart version found matching pattern %q", c.versionFilter.Pattern)
	}

	resultSource.Information = []result.SourceInformation{{
		Key:   resultSource.ID,
		Value: version,
	}}
	resultSource.Description = fmt.Sprintf("OCI version %q found matching pattern %q", version, c.versionFilter.Pattern)
	resultSource.Result = result.SUCCESS

	return nil
}
