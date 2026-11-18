package dockerimage

import (
	"context"
	"fmt"
	"time"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

// Condition checks if a docker image with a specific tag is published
// We assume that if we can't retrieve the docker image digest, then it means
// it doesn't exist.
func (di *DockerImage) Condition(_ context.Context, source string, scm scm.ScmHandler, pathResolver pathresolver.Resolver) (pass bool, message string, err error) {
	if scm != nil {
		log.Warningf("SCM configuration is not supported for condition of type dockerimage. Remove the `scm` directive from condition to remove this warning message")
	}

	version := source
	if di.spec.Tag != "" {
		version = di.spec.Tag
	}

	ref, err := di.createRef(version)
	if err != nil {
		return false, "", err
	}

	found := true

	if len(di.spec.Architectures) == 0 {
		found, err = di.checkImage(ref, "")
		if err != nil {
			return false, "", err
		}
	} else {
		for _, arch := range di.spec.Architectures {
			foundArchitecture, err := di.checkImage(ref, arch)
			if err != nil {
				return false, "", err
			}
			if !foundArchitecture {
				found = false
				break
			}
		}
	}

	if !found {
		return false, fmt.Sprintf("docker image %s:%s not found", di.spec.Image, version), nil
	}

	if !di.spec.Age.IsZero() {
		return di.conditionAge(version)
	}

	return true, fmt.Sprintf("docker image %s:%s found", di.spec.Image, version), nil
}

// conditionAge checks the creation date of a tag against the age filter, for every requested platform.
func (di *DockerImage) conditionAge(tag string) (bool, string, error) {
	architectures := di.spec.Architectures
	if len(architectures) == 0 {
		architectures = []string{""}
	}

	for _, arch := range architectures {
		image := fmt.Sprintf("%s:%s", di.spec.Image, tag)
		if arch != "" {
			image += " (" + arch + ")"
		}

		date, dated, err := di.releaseDate(tag, arch)
		if err != nil {
			return false, "", err
		}

		if !dated {
			return false, fmt.Sprintf("docker image %s has no creation date, so the age filter can't be applied", image), nil
		}

		if !di.spec.Age.Matches(date) {
			return false, fmt.Sprintf("docker image %s created on %s is outside the age filter", image, date.Format(time.RFC3339)), nil
		}
	}

	return true, fmt.Sprintf("docker image %s:%s found and matching the age filter", di.spec.Image, tag), nil
}
