package language

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// Changelog returns a link to the Golang version
func (l *Language) Changelog() string {

	v, err := semver.NewVersion(l.foundVersion.OriginalVersion)

	if err != nil {
		logrus.Errorf("failing parsing version %q", err)
		return ""
	}

	url := fmt.Sprintf("https://go.dev/doc/go%d.%d.minor", v.Major(), v.Minor())
	if v.Patch() == 0 {
		url = fmt.Sprintf("https://go.dev/doc/go%d.%d", v.Major(), v.Minor())
	}

	return fmt.Sprintf("Golang changelog for version %q is available on %q", l.foundVersion.OriginalVersion, url)
}
