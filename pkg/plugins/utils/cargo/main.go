package cargo

import (
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

// InlineKeyChain defines the credentials used to authenticate with a cargo registry.
type InlineKeyChain struct {
	// "token" defines the cargo registry token used for authentication.
	//
	Token string `yaml:",omitempty"`
	// "headerformat" defines the format of the Authorization header sent with "token".
	//
	// default:
	//   Bearer %s
	//
	// remark:
	//   * "%s" is replaced by the token.
	//
	// example:
	//   * headerformat: "Token %s"
	//
	HeaderFormat string `yaml:"headerformat,omitempty"`
}

// Registry defines the cargo registry to query.
type Registry struct {
	// "auth" defines the credentials used to authenticate with the cargo registry.
	//
	Auth InlineKeyChain `yaml:",omitempty"`
	// "url" defines the URL of the cargo registry API.
	//
	// default:
	//   https://crates.io/api/v1/crates, when neither "rootdir" nor a scm is set.
	//
	// remark:
	//   * "url", "rootdir" and "scmid" are mutually exclusive.
	//
	URL string `yaml:",omitempty"`
	// "rootdir" defines the local directory of a cargo registry index, used instead of the registry API.
	//
	// remark:
	//   * "url", "rootdir" and "scmid" are mutually exclusive.
	//
	RootDir string `yaml:",omitempty"`
	// "scmid" defines the scm holding the cargo registry index, used instead of the registry API.
	//
	// remark:
	//   * only used by the cargo autodiscovery.
	//   * "url", "rootdir" and "scmid" are mutually exclusive.
	//
	SCMID string `yaml:",omitempty"`
}

func (r Registry) Validate() bool {
	if r.RootDir != "" && r.SCMID != "" {
		logrus.Errorf("Registry.RootDir is defined and set to %q but would be overridden by the scm %q",
			r.RootDir,
			r.SCMID)
		return false
	}
	if r.URL != "" && r.SCMID != "" {
		logrus.Errorf("Registry.URL is defined and set to %q but would be overridden by the scm %q",
			r.URL,
			r.SCMID)
		return false
	}
	if r.RootDir != "" && r.URL != "" {
		logrus.Errorf("Registry.URL is defined and set to %q but would be overridden by Registry.RootDir %q",
			r.URL,
			r.RootDir)
		return false
	}
	return true
}

func IsCargoInstalled() bool {
	cmd := exec.Command("cargo", "--version")
	err := cmd.Run()
	return err == nil
}

func IsLockFileDetected(lockfile string) bool {
	_, err := os.Stat(lockfile)
	return err == nil
}
