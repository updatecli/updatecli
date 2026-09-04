package gittag

import (
	"context"
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
)

// Source returns the latest git tag based on create time
func (gt *GitTag) Source(_ context.Context, workingDir string, resultSource *result.Source) error {
	var err error

	gt.directory = workingDir

	err = gt.Validate()
	if err != nil {
		return fmt.Errorf("validate git tag: %w", err)
	}

	var tags map[string]string
	var tagsList []string

	switch gt.lsRemote {
	case true:
		tagsList, tags, err = gt.listRemoteURLTags()
		if err != nil {
			return err
		}

	case false:
		tagsList, tags, err = gt.listRemoteDirectoryTags(workingDir, gt.spec.Age)
		if err != nil {
			/*
				Every published tag is still cooling down, which is an expected state of
				the age filter rather than a failure, so the source is skipped instead.
			*/
			if errors.Is(err, age.ErrNoVersionMatchingAge) {
				resultSource.Result = result.SKIPPED
				resultSource.Description = "no git tag matches the age filter yet"
				return nil
			}

			return fmt.Errorf("listing local tags: %w", err)
		}
	}

	if len(tagsList) == 0 {
		return fmt.Errorf("no tags found")
	}

	gt.foundVersion, err = gt.versionFilter.Search(tagsList)
	if err != nil {
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
