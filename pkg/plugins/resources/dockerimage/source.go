package dockerimage

import (
	"fmt"
	"regexp"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (di *DockerImage) Source(workingDir string, resultSource *result.Source) error {
	repo, err := name.NewRepository(di.spec.Image)
	if err != nil {
		return fmt.Errorf("invalid repository %s: %w", di.spec.Image, err)
	}
	logrus.Debugf(
		"Searching tags for the image %q",
		repo,
	)

	tags, err := remote.List(repo, di.options...)
	if err != nil {
		return fmt.Errorf("unable to list tags for repository %s: %w", repo, err)
	}

	// apply tagFilter

	logrus.Debugf("%d Docker image tag(s) found", len(tags))

	if di.spec.TagFilter != "" {
		tags = di.filterTags(tags)
	}

	di.foundVersion, err = di.versionFilter.Search(tags)
	if err != nil {
		return fmt.Errorf("filtering tags: %w", err)
	}
	tag := di.foundVersion.GetVersion()

	if len(tag) == 0 {
		return fmt.Errorf("no Docker Image Tag tag found matching pattern %q", di.versionFilter.Pattern)
	}

	ref, err := di.createRef(tag)
	if err != nil {
		return err
	}

	architecture := ""

	if len(di.spec.Architectures) > 0 {
		architecture = di.spec.Architectures[0]
	}

	found, err := di.checkImage(ref, architecture)
	if err != nil {
		return err
	}

	if !found {
		return fmt.Errorf("no Docker Image for architecture %s", di.spec.Architectures[0])
	}

	resultSource.Result = result.SUCCESS
	resultSource.Information = []result.SourceInformation{{
		Key:   resultSource.ID,
		Value: tag,
	}}
	resultSource.Description = fmt.Sprintf("Docker Image Tag %q found matching pattern %q", tag, di.versionFilter.Pattern)

	return nil
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
