package language

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
)

// Changelog returns a link to the Golang version
func (l *Language) Changelog(from, to string) *result.Changelogs {

	var err error

	if from == "" && to == "" {
		return nil
	}

	// Parse the target version (to) for changelog generation
	toVersion, err := semver.NewVersion(to)
	if err != nil {
		logrus.Errorf("failing parsing to version %q - %q",
			to, err)
		return nil
	}

	title := toVersion.String()
	url := fmt.Sprintf("https://go.dev/doc/devel/release#go%d.%d.minor", toVersion.Major(), toVersion.Minor())
	body := fmt.Sprintf("Golang changelog for version %q is available on %q", toVersion.String(), redact.URL(url))

	if toVersion.Patch() == 0 {
		title = fmt.Sprintf("%d.%d", toVersion.Major(), toVersion.Minor())
		url = fmt.Sprintf("https://go.dev/doc/go%d.%d", toVersion.Major(), toVersion.Minor())
		body = fmt.Sprintf(`Golang changelog for version "%d.%d" is available on %q`,
			toVersion.Major(),
			toVersion.Minor(),
			redact.URL(url))
	}

	return &result.Changelogs{
		{
			Title: title,
			Body:  body,
			URL:   url,
		},
	}
}
