package temurin

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

/*
Condition tests if the response of the specified HTTP request meets assertion.
If no assertion is specified, it only checks for successful HTTP response code (HTTP/1xx, HTTP/2xx or HTTP/3xx).
*/
func (t *Temurin) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	if len(source) > 0 {
		logrus.Infof("[temurin] using specific version (`spec.SpecificVersion`) from source value %q. Set `disablesourceinput` to true to avoid this behavior.", source)
		if t.spec.SpecificVersion != "" {
			logrus.Warning("[temurin] Overriding specified version (`spec.SpecificVersion`) with the source input value", source)
		}
		t.spec.SpecificVersion = source
	}

	if len(t.spec.Platforms) > 0 {
		pass = true
		for _, platform := range t.spec.Platforms {
			t.spec.OperatingSystem = strings.Split(platform, "/")[0]
			t.spec.Architecture = strings.Split(platform, "/")[1]

			found, message, err := t.checkRelease()
			if err != nil {
				return false, "", err
			}

			if found {
				logrus.Infof(result.SUCCESS+" (Platform %q) "+message, platform)
			} else {
				logrus.Infof(result.FAILURE+" (Platform %q) "+message, platform)
			}

			pass = pass && found
		}

		if pass {
			message = "All releases found."
		} else {
			message = "Some releases were not found."
		}

		return pass, message, nil
	}

	return t.checkRelease()
}

func (t *Temurin) checkRelease() (pass bool, message string, err error) {
	foundReleases, err := t.apiGetReleaseNames()
	if err != nil {
		return false, "", err
	}

	if len(foundReleases) == 0 {
		return false, "No release found with specified attributes.", nil
	}

	if t.spec.SpecificVersion != "" {
		for _, foundRelease := range foundReleases {
			if strings.Contains(foundRelease, t.spec.SpecificVersion) {
				return true, fmt.Sprintf("Found release %q which maps specified %q version.", foundRelease, t.spec.SpecificVersion), nil
			}
		}
		return false, fmt.Sprintf("Release %q (either specified or from source input) not found with specified attributes.", t.spec.SpecificVersion), nil
	}

	return true, "Release exists.", nil
}
