package source

import "github.com/updatecli/updatecli/pkg/plugins/version"

// Changelog is an interface to retrieve changelog description
type Changelog interface {
	Changelog(version.Version) (string, error)
}
