package osv

import (
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Changelog is not supported for the osv plugin.
func (o *Osv) Changelog(from, to string) *result.Changelogs {
	return nil
}
