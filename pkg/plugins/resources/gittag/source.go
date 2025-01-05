package gittag

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest git tag based on create time
func (gt *GitTag) Source(workingDir string, resultSource *result.Source) error {
	var err error

	gt.directory = workingDir

	if gt.spec.URL != "" {
		gt.directory, err = gt.clone()
		if err != nil {
			return err
		}

	} else if gt.spec.Path != "" {
		gt.directory = gt.spec.Path
	}

	err = gt.Validate()
	if err != nil {
		return fmt.Errorf("validate git tag: %w", err)
	}

	if gt.directory == "" {
		return fmt.Errorf("Unkownn Git working directory. Did you specify one of `URL`, `scmID`, or `spec.path`?")
	}

	refs, err := gt.nativeGitHandler.TagRefs(gt.directory)
	if err != nil {
		return fmt.Errorf("retrieving tag refs: %w", err)
	}

	if len(refs) == 0 {
		return fmt.Errorf("no tags found at path %q", gt.directory)
	}

	var tags []string
	for i := range refs {
		tags = append(tags, refs[i].Name)
	}

	gt.foundVersion, err = gt.versionFilter.Search(tags)
	if err != nil {
		return fmt.Errorf("filtering tags: %w", err)
	}

	name := gt.foundVersion.GetVersion()
	var hash string

	for i := range refs {
		if refs[i].Name == name {
			hash = refs[i].Hash
		}
	}
	resultSource.Information = name
	if gt.spec.Key == "hash" {
		resultSource.Information = hash
	}

	if len(resultSource.Information) == 0 {
		return fmt.Errorf("no git tag found matching pattern %q of kind %q",
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
