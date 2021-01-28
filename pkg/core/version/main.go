package version

import (
	"github.com/sirupsen/logrus"
	"strings"
)

var (
	// Version contains application version
	Version string

	// BuildTime contains application build time
	BuildTime string

	// GoVersion contains the golang version uses to build this binary
	GoVersion string
)

// Show displays various version information
func Show() {
	strings.ReplaceAll(GoVersion, "go version go", "Golang     :")
	logrus.Infof("")
	logrus.Infof("Application:\t%s", Version)
	logrus.Infof("%s", strings.ReplaceAll(GoVersion, "go version go", "Golang     :\t"))
	logrus.Infof("Build Time :\t%s", BuildTime)
	logrus.Infof("")
}
