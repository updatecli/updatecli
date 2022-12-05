package dockerimage

import (
	"fmt"
	"regexp"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (di *DockerImage) Source(workingDir string) (string, error) {

	repo, err := name.NewRepository(di.spec.Image)
	if err != nil {
		return "", fmt.Errorf("invalid repository %s: %w", di.spec.Image, err)
	}
	logrus.Debugf(
		"Searching tags for the image %q",
		repo,
	)

	tags, err := remote.List(repo, di.options...)
	if err != nil {
		return "", fmt.Errorf("unable to list tags for repository %s: %w", repo, err)
	}

	// apply tagFilter

	logrus.Debugf("%d Docker image tag(s) found", len(tags))

	if di.spec.TagFilter != "" {
		tags = di.filterTags(tags)
	}

	di.foundVersion, err = di.versionFilter.Search(tags)
	if err != nil {
		return "", err
	}
	tag := di.foundVersion.GetVersion()

	if len(tag) == 0 {
		logrus.Infof("%s No Docker Image Tag found matching pattern %q", result.FAILURE, di.versionFilter.Pattern)
		return tag, fmt.Errorf("no Docker Image Tag tag found matching pattern %q", di.versionFilter.Pattern)
	} else if len(tag) > 0 {
		logrus.Infof("%s Docker Image Tag %q found matching pattern %q", result.SUCCESS, tag, di.versionFilter.Pattern)
	} else {
		logrus.Errorf("Something unexpected happened in dockerimage source")
	}

	return tag, nil
}

func (di *DockerImage) filterTags(tags []string) []string {
	var results []string
	re, err := regexp.Compile(di.spec.TagFilter)
	if err != nil {
		logrus.Errorln(err)
		logrus.Debugln("=> something went wrong, falling back to latest versioning")
		return []string{}
	}
	for _, tag := range tags {
		if re.MatchString(tag) {
			results = append(results, tag)
		}
	}

	fmt.Println()

	return results
}
