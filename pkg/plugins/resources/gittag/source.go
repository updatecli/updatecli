package gittag

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest git tag based on create time
func (gt *GitTag) Source(workingDir string, resultSource *result.Source) error {
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
		tagsList, tags, err = gt.listRemoteDirectoryTags(workingDir)
		if err != nil {
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
