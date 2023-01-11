package cargopackage

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for a "dockerimage" resource
// parsed from an updatecli manifest file
type Spec struct {
	// [S][C] IndexUrl specifies the URL of the index to use to check version
	IndexUrl string `yaml:",omitempty"`
	// [S][C] Package specifies the name of the package
	Package string `yaml:",omitempty" jsonschema:"required"`
	// [C] Defines a specific package version
	Version string `yaml:",omitempty"`
	// [S][C] Username specifies the git username to access the index
	Username string `yaml:",omitempty"`
	// [S][C] Password specifies the git password to access the index
	Password string `yaml:",omitempty"`
	// [S][C] PrivateKeyFile specifies the path of the ssh private key
	PrivateKey string `yaml:",omitempty"`
	// [S][C] PrivateKeyUser specifies the user of the ssh private key (default to `git`)
	PrivateKeyUser string `yaml:",omitempty"`
	// [S][C] PrivateKeyPassword specifies the password of the ssh private key
	PrivateKeyPassword string `yaml:",omitempty"`
	// [S] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
}

func (s *Spec) Validate() (err error) {
	if s.Username != "" && s.Password == "" {
		logrus.Errorf("Password should be specified when using user/password authentication")
		return errors.New("password not defined while using user/password auth")
	}
	if s.Username != "" && s.PrivateKey != "" {
		logrus.Errorf("Username and PrivateKey configured at the same time, choose one")
		return errors.New("user/password should not be used at the same time")
	}
	return nil
}
