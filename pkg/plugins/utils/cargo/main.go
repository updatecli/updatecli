package cargo

import "github.com/sirupsen/logrus"

type InlineKeyChain struct {
	// [A][S][C] Token specifies the cargo registry token to use for authentication.
	Token string `yaml:",omitempty"`
	// [A][S][C] HeaderFormat specifies the cargo registry header format to use for authentication (defaults to `Bearer`).
	HeaderFormat string `yaml:"headerFormat,omitempty"`
}

type Registry struct {
	// [A][S][C] Auth specifies the cargo registry auth to use for authentication.
	Auth InlineKeyChain `yaml:",omitempty"`
	// [A][S][C] URL specifies the cargo registry URL to use for authentication.
	URL string `yaml:",omitempty"`
	// [A][S][C] RootDir specifies the cargo registry root directory to use as FS index.
	RootDir string `yaml:",omitempty"`
	// [A] SCMID specifies the cargo registry scmId to use as FS index.
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
