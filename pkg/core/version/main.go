package version

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	sv "github.com/Masterminds/semver/v3"
)

var (
	// Version contains application version
	Version string

	// BuildTime contains application build time
	BuildTime string

	// GoVersion contains the golang version uses to build this binary
	GoVersion string

	// DisableDevWarning is used to identify if we already notify that we use a dev version
	isDevWarningDisabled bool
)

// Show displays various version information
func Show() {
	logrus.Infof("")
	logrus.Infof("Application:\t%s", Version)
	logrus.Infof("%s", strings.ReplaceAll(GoVersion, "go version go", "Golang     :\t"))
	logrus.Infof("Build Time :\t%s", BuildTime)
	logrus.Infof("")
}

// IsGreaterThan test if an updatecli manifest required version is greater or equal to the current updatecli binary version
// Please not that empty version are set to 0.0.0
// A prerelease binary version such as 1.0.0-rc.1 is compared on its release line, so it satisfies a manifest requiring 1.0.0
func IsGreaterThan(binaryVersion, manifestVersion string) (bool, error) {
	if len(manifestVersion) == 0 {
		manifestVersion = "0.0.0"
	}

	if len(binaryVersion) == 0 {
		if !isDevWarningDisabled {
			logrus.Warningf("Updatecli binary version is unset. This means you are using a development version that ignores manifest version constraint.\n")
			isDevWarningDisabled = true
		}

		return true, nil
	}

	mv, err := sv.NewVersion(manifestVersion)
	if err != nil {
		return false, fmt.Errorf("can't parse Updatecli manifest version %q - %q", manifestVersion, err)
	}

	bv, err := sv.NewVersion(binaryVersion)
	if err != nil {
		return false, fmt.Errorf("can't parse Updatecli binary version %q - %q", binaryVersion, err)
	}

	// Semver ranks a prerelease below the version it leads to, so a 1.0.0-rc.1
	// binary would reject a manifest requiring 1.0.0. A release candidate must
	// be able to run the manifests of the version it is a candidate for, so the
	// comparison is made on the binary's release line and ignores the prerelease
	// and build identifiers.
	if bv.Prerelease() != "" || bv.Metadata() != "" {
		bv = sv.New(bv.Major(), bv.Minor(), bv.Patch(), "", "")
	}

	return bv.GreaterThan(mv) || bv.Equal(mv), nil
}
