package dockerimage

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

func (di *DockerImage) Source(_ context.Context, pathResolver pathresolver.Resolver, resultSource *result.Source) error {
	repo, err := name.NewRepository(di.spec.Image)
	if err != nil {
		return fmt.Errorf("invalid repository %s: %w", di.spec.Image, err)
	}
	log.Debugf(
		"Searching tags for the image %q",
		repo,
	)

	tags, err := remote.List(repo, di.options...)
	if err != nil {
		return fmt.Errorf("unable to list tags for repository %s: %w", repo, err)
	}

	// apply tagFilter

	log.Debugf("%d Docker image tag(s) found", len(tags))

	if di.spec.TagFilter != "" {
		tags = di.filterTags(tags)
	}

	architecture := ""

	if len(di.spec.Architectures) > 0 {
		architecture = di.spec.Architectures[0]
	}

	if di.spec.Age.IsZero() {
		di.foundVersion, err = di.versionFilter.Search(tags)
	} else {
		di.foundVersion, err = di.tagMatchingAge(tags, architecture)

		switch {
		case errors.Is(err, age.ErrNoVersionMatchingAge):
			resultSource.Result = result.SKIPPED
			resultSource.Description = fmt.Sprintf("no tag of the Docker image %s matching pattern %q matches the age filter yet", di.spec.Image, di.versionFilter.Pattern)
			return nil
		case errors.Is(err, errNoReleaseDate):
			resultSource.Result = result.SKIPPED
			resultSource.Description = fmt.Sprintf("the age filter can't be applied to the Docker image %s, as none of the tags checked reports a creation date: remove it", di.spec.Image)
			return nil
		}
	}
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

	found, err := di.checkImage(ref, architecture)
	if err != nil {
		return err
	}

	if !found {
		if architecture == "" {
			return fmt.Errorf("no Docker Image found for tag %s", tag)
		}
		return fmt.Errorf("no Docker Image for architecture %s", architecture)
	}

	resultSource.Result = result.SUCCESS
	resultSource.Information = tag
	resultSource.Description = fmt.Sprintf("Docker Image Tag %q found matching pattern %q", tag, di.versionFilter.Pattern)
	if !di.spec.Age.IsZero() {
		resultSource.Description += " and the age filter"
	}

	return nil
}

func (di *DockerImage) filterTags(tags []string) []string {
	var results []string
	re, err := regexp.Compile(di.spec.TagFilter)
	if err != nil {
		log.Errorln(err)
		log.Debugln("=> something went wrong, falling back to latest versioning")
		return []string{}
	}
	for _, tag := range tags {
		if re.MatchString(tag) {
			results = append(results, tag)
		}
	}

	return results
}
