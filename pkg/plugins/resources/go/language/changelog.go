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
	var froundVersion *semver.Version

	if from == "" && to == "" {
		return nil
	}

	versions, err := l.versions()
	if err != nil {
		logrus.Errorf("failed to retrieve golang version: %s", err)
		return nil
	}

	if len(versions) == 0 {
		logrus.Errorf("no Golang version found")
		return nil
	}

	fromVersion, err := semver.NewVersion(from)
	if err != nil {
		logrus.Errorf("failing parsing from version %q - %q",
			from, err)
		return nil
	}

	for i := range versions {
		froundVersion, err = semver.NewVersion(versions[i])
		if err != nil {
			logrus.Debugf("failing parsing from version %q - %q",
				versions[i], err)
			return nil
		}

		if fromVersion.Equal(froundVersion) {
			break
		}
	}

	title := froundVersion.String()
	url := fmt.Sprintf("https://go.dev/doc/devel/release#go%d.%d.minor", froundVersion.Major(), froundVersion.Minor())
	body := fmt.Sprintf("Golang changelog for version %q is available on %q", froundVersion.String(), redact.URL(url))

	if froundVersion.Patch() == 0 {
		title = fmt.Sprintf("%d.%d", froundVersion.Major(), froundVersion.Minor())
		url = fmt.Sprintf("https://go.dev/doc/go%d.%d", froundVersion.Major(), froundVersion.Minor())
		body = fmt.Sprintf(`Golang changelog for version "%d.%d" is available on %q`,
			froundVersion.Major(),
			froundVersion.Minor(),
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
