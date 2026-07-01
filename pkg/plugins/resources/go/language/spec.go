package language

import (
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for a "Golang" resource parsed from an updatecli manifest file
type Spec struct {
	// Version defines a specific golang version
	//
	// Compatible:
	//   * condition
	//
	Version string `yaml:",omitempty"`
	// versionfilter provides parameters to specify version pattern and
	// its type like regex, semver, or just latest.
	//
	// Compatible:
	//   * source
	//
	VersionFilter version.Filter `yaml:",omitempty"`
	// age defines the minimum or maximum age of a release to be considered valid. It accepts a duration string (e.g., "24h", "7d").
	//
	// Compatible:
	//   * source
	//   * condition
	//
	// Remarks:
	//   If age is specified, the Updatecli retrieves the release date of each Golang version from a git repository.
	//   This has significant performance implications, as it requires fetching git tags and their associated commit dates.
	//   Use wisely.
	Age age.Spec `yaml:",omitempty"`
}
