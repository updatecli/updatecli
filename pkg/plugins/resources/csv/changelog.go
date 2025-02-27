package csv

import "github.com/updatecli/updatecli/pkg/core/result"

// Changelog returns the changelog for this resource, or an empty string if not supported
func (c *CSV) Changelog(from, to string) *result.Changelogs {
	return nil
}
