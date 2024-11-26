package helm

import (
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/result"
	"helm.sh/helm/v3/pkg/repo"
)

// Source return the latest version
func (c *Chart) Source(workingDir string, resultSource *result.Source) error {

	if strings.HasPrefix(c.spec.URL, "oci://") {
		return c.OCISource(workingDir, resultSource)
	}

	var index repo.IndexFile
	var err error

	if strings.HasPrefix(c.spec.URL, "https://") || strings.HasPrefix(c.spec.URL, "http://") {
		index, err = c.GetRepoIndexFromURL()
		if err != nil {
			return fmt.Errorf("getting repo index from url: %w", err)
		}
	} else {
		index, err = c.GetRepoIndexFromFile(workingDir)
		if err != nil {
			return fmt.Errorf("getting repo index from file: %w", err)
		}
	}

	entriesVersion, found := index.Entries[c.spec.Name]
	if !found {
		return fmt.Errorf("helm chart %q not found from Helm Chart repository %q", c.spec.Name, c.spec.URL)
	}

	versions := []string{}
	for i := len(entriesVersion) - 1; i >= 0; i-- {
		versions = append(versions, entriesVersion[i].Version)
	}

	c.foundVersion, err = c.versionFilter.Search(versions)
	if err != nil {
		return fmt.Errorf("filtering version: %w", err)
	}

	version := c.foundVersion.GetVersion()
	if version == "" {
		return fmt.Errorf("no Helm Chart version found for %q", c.spec.Name)
	}
	resultSource.Information = []result.SourceInformation{{
		Key:   resultSource.ID,
		Value: version,
	}}

	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("Helm Chart %q version %q is found from repository %q",
		c.spec.Name,
		version,
		c.spec.URL)

	return nil
}
