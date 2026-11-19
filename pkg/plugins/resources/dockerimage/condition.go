package dockerimage

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

// Condition checks if a docker image with a specific tag is published
// We assume that if we can't retrieve the docker image digest, then it means
// it doesn't exist.
func (di *DockerImage) Condition(ctx context.Context, source string, scm scm.ScmHandler, pathResolver pathresolver.Resolver) (pass bool, message string, err error) {
	if scm != nil {
		log.Warningf("SCM configuration is not supported for condition of type dockerimage. Remove the `scm` directive from condition to remove this warning message")
	}

	version := source
	if di.spec.Tag != "" {
		version = di.spec.Tag
	}

	if !di.spec.Age.IsZero() {
		return di.conditionAge(ctx, version)
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

	return true, fmt.Sprintf("docker image %s:%s found", di.spec.Image, version), nil
}

// conditionAge checks a tag exists and matches the age filter for every requested platform.
// Dating a tag already checks it exists, so the registry is queried once per platform.
func (di *DockerImage) conditionAge(ctx context.Context, tag string) (bool, string, error) {
	options, err := di.ageOptions(ctx)
	if err != nil {
		return false, "", err
	}

	got, err := di.tagAge(tag, options)
	if err != nil {
		return false, "", err
	}

	if !got.matches {
		return false, got.reason, nil
	}

	return true, fmt.Sprintf("docker image %s:%s found and matching the age filter", di.spec.Image, tag), nil
}
