package gittag

import (
	"context"
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

// Source returns the latest git tag based on create time
func (gt *GitTag) Source(_ context.Context, pathResolver pathresolver.Resolver, resultSource *result.Source) error {
	var err error

	gt.directory = pathResolver.RepositoryDir()

	err = gt.Validate()
	if err != nil {
		return fmt.Errorf("validate git tag: %w", err)
	}

	var tags map[string]string
	// tagsList holds the tags within the age window, published every tag
	var tagsList, published []string

	switch gt.lsRemote {
	case true:
		// Remote tags are listed without their date, so the age filter never applies to them
		tagsList, tags, err = gt.listRemoteURLTags()
		if err != nil {
			return err
		}
		published = tagsList

	case false:
		tagsList, tags, published, err = gt.listRemoteDirectoryTags(gt.directory, gt.spec.Age, pathResolver)
		if err != nil {
			return fmt.Errorf("listing local tags: %w", err)
		}
	}

	if len(published) == 0 {
		return fmt.Errorf("no tags found")
	}

	gt.foundVersion, err = age.Search(&gt.versionFilter, published, tagsList)
	if err != nil {
		/*
			Every tag matching the version filter is still cooling down, which is an expected
			state of the age filter rather than a failure, so the source is skipped instead.
		*/
		if errors.Is(err, age.ErrNoVersionMatchingAge) {
			resultSource.Result = result.SKIPPED
			resultSource.Description = "no git tag matching the version filter matches the age filter yet"
			return nil
		}

		return fmt.Errorf("filtering tags: %w", err)
	}

	name := gt.foundVersion.GetVersion()

	var hash string
	if _, ok := tags[name]; ok {
		hash = tags[name]
	}

	resultSource.Information = name
	if gt.spec.Key == "hash" {
		resultSource.Information = hash
	}

	if len(resultSource.Information) == 0 {
		return fmt.Errorf("no Git tag found matching pattern %q of kind %q",
			gt.versionFilter.Pattern,
			gt.versionFilter.Kind,
		)
	}

	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("Git tag %q found matching pattern %q of kind %q",
		resultSource.Information,
		gt.versionFilter.Pattern,
		gt.versionFilter.Kind,
	)

	return nil
}
