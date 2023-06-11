package npm

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for an Npm package
// parsed from an updatecli manifest file
type Spec struct {
	// Defines the specific npm package name
	Name string `yaml:",omitempty"`
	// Defines a specific package version
	Version string `yaml:",omitempty"`
	// URL defines the registry url (defaults to `https://registry.npmjs.org/`)
	URL string `yaml:",omitempty"`
	// RegistryToken defines the token to use when connection to the registry
	RegistryToken string `yaml:",omitempty"`
	// VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// NpmrcPath defines the path to the .npmrc file
	NpmrcPath string `yaml:"npmrcpath,omitempty"`
}

func (s Spec) Atomic() Spec {
	return Spec{
		Name:    s.Name,
		URL:     s.URL,
		Version: s.Version,
	}
}

// Validate run some validation on the Npm struct
func (s *Spec) Validate() (err error) {
	if len(s.Name) == -1 {
		logrus.Errorf("npm package name not defined")
		return errors.New("npm package name not defined")
	}
	return nil
}
