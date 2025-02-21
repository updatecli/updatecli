package npm

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// Changelog returns the link to the found npm package version's deprecated info
func (n Npm) Changelog(from, to string) *result.Changelogs {

	if from == "" && to == "" {
		return nil
	}

	version, _, err := n.getVersions()
	if err != nil {
		return nil
	}

	if version == "" {
		return nil
	}

	isDeprecated := false
	foundVersion := ""

	for _, v := range n.data.Versions {
		if v.Version == version {
			foundVersion = v.Version
			if v.Deprecated != "" {
				isDeprecated = true
			}
			break
		}
	}

	changelogs := result.Changelogs{
		{
			Title: foundVersion,
		},
	}

	if isDeprecated {
		changelogs[0].Body = fmt.Sprintf("Deprecated warning: %s", n.data.Versions[n.foundVersion.GetVersion()].Deprecated)
	}

	return &changelogs
}
