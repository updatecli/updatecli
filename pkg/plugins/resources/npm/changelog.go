package npm

import "fmt"

// Changelog returns the link to the found npm package version's deprecated info
func (n Npm) Changelog() string {
	if n.foundVersion.GetVersion() == "" {
		return ""
	}

	changelog := fmt.Sprintf(
		"Deprecated Warning: %s\n",
		n.data.Versions[n.foundVersion.GetVersion()].Deprecated,
	)

	return changelog
}
