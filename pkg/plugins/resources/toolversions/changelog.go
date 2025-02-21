package toolversions

import "github.com/updatecli/updatecli/pkg/core/result"

// Changelog returns the changelog for this resource, or an empty string if not supported
func (t *ToolVersions) Changelog(from, to string) *result.Changelogs {
	return nil
}
