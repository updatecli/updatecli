package github

import (
	"fmt"
	"strings"
)

// ReleaseType specifies accepts a list of GitHub Release type ["draft","prerelease","release","latest"]
type ReleaseType []string

var (
	ValidReleaseType []string = []string{
		"draft",
		"prerelease",
		"release",
		"latest",
	}
)

func (r *ReleaseType) Init() {
	release := *r

	if len(release) == 0 {
		release = append(release, "release")
	}
	r = &release
}

// IsZero checks if all release type are set to disable
func (r ReleaseType) IsZero() bool {
	return len(r) == 0
}

func (r ReleaseType) IsEqual(release string) bool {
	for i := range r {
		if strings.ToLower(r[i]) == release {
			return true
		}
	}
	return false
}

func (r ReleaseType) Validate() error {
	var notValidReleaseType []string

	for i := range r {
		valid := false
		for j := range ValidReleaseType {
			if strings.ToLower(r[i]) == ValidReleaseType[j] {
				valid = true
				break
			}
		}
		if !valid {
			notValidReleaseType = append(notValidReleaseType, r[i])
		}
	}

	if len(notValidReleaseType) > 0 {
		return fmt.Errorf("release type [%q] not valid", strings.Join(notValidReleaseType, "\",\""))
	}
	return nil
}
