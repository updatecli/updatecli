package json

import "github.com/updatecli/updatecli/pkg/core/result"

// Changelog returns the changelog for this resource, or an empty string if not supported
func (j *Json) Changelog(from, to string) *result.Changelogs {
	return nil
}
