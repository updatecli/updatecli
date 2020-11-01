package version

import (
	"fmt"
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
	fmt.Printf("\n")
	fmt.Printf("Application:\t%s\n", Version)
	fmt.Printf("%s\n", strings.ReplaceAll(GoVersion, "go version go", "Golang     :\t"))
	fmt.Printf("Build Time :\t%s\n", BuildTime)
	fmt.Printf("\n")
}
