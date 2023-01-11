package cargopackage

import (
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for a "dockerimage" resource
// parsed from an updatecli manifest file
type Spec struct {
	// [S][C] IndexDir specifies the directory of the index to use to check version
	IndexDir string `yaml:",omitempty"`
	// [S][C] Package specifies the name of the package
	Package string `yaml:",omitempty" jsonschema:"required"`
	// [C] Defines a specific package version
	Version string `yaml:",omitempty"`
	// [S] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
}

// Validate tests that tag struct is correctly configured
func (s *Spec) Validate() error {
	validationErrors := []string{}
	if s.IndexDir == "" {
		validationErrors = append(validationErrors, "Index directory path is empty while it must be specified. Did you specify an `scmID` or a `spec.indexDIR`?")
	}

	// Return all the validation errors if found any
	if len(validationErrors) > 0 {
		return fmt.Errorf("validation error: the provided manifest configuration has the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	return nil
}
