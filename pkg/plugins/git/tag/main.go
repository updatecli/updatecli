package tag

import (
	"github.com/olblak/updateCli/pkg/plugins/version"
)

// Tag contains git tag information
type Tag struct {
	Path          string         // Path contains the git repository path
	VersionFilter version.Filter // //VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
}

// Validate tests that tag struct is correctly configured
func (t *Tag) Validate() error {
	err := t.VersionFilter.Validate()
	if err != nil {
		return err
	}
	return nil
}
