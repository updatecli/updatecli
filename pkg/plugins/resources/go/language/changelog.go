package language

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
)

// Changelog returns a link to the Golang version
func (l *Language) Changelog() string {

	v, err := semver.NewVersion(l.Version.GetVersion())

	if err != nil {
		logrus.Errorf("failing parsing version %q - %q",
			l.Version.GetVersion(), err)
		return ""
	}

	url := fmt.Sprintf("https://go.dev/doc/devel/release#go%d.%d.minor", v.Major(), v.Minor())
	if v.Patch() == 0 {
		url = fmt.Sprintf("https://go.dev/doc/go%d.%d", v.Major(), v.Minor())
	}

	return fmt.Sprintf("Golang changelog for version %q is available on %q", l.Version.OriginalVersion, redact.URL(url))
}
