package gittag

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Target creates a tag if needed from a local git repository, without pushing the tag
func (gt *GitTag) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {
	if scm != nil {
		if len(gt.spec.Path) > 0 {
			logrus.Warningf("Path setting value %q overridden by the scm configuration (value %q)",
				gt.spec.Path,
				scm.GetDirectory())
		}
		gt.spec.Path = scm.GetDirectory()
	}

	if err := gt.target(source, dryRun, resultTarget); err != nil {
		return err
	}

	if dryRun || !resultTarget.Changed {
		return nil
	}

	if scm != nil {
		if err := scm.PushTag(source); err != nil {
			logrus.Errorf("Git push tag error: %s", err)
			return err
		}
	}

	resultTarget.Description = fmt.Sprintf("git tag %q successfully created and pushed", source)
	if gt.spec.Message != "" {
		resultTarget.Description = gt.spec.Message
	}

	return nil
}

func (gt *GitTag) target(source string, dryRun bool, resultTarget *result.Target) error {
	// Ensure that a git message is present to annotate the tag to create
	if len(gt.spec.Message) == 0 {
		// absence of a message is not blocking: warn the user and continue
		gt.spec.Message = "Generated by updatecli"
		logrus.Warningf("No specified message for gittag target. Using default value %q", gt.spec.Message)
	}

	// cfr https://github.com/updatecli/updatecli/issues/1126
	// to know why the following line is needed at the moment
	resultTarget.Files = []string{""}

	// Fail if a pattern is specified
	if gt.spec.VersionFilter.Pattern != "" {
		return fmt.Errorf("target validation error: spec.versionfilter.pattern is not allowed for targets of type gittag")
	}

	// Fail if the git tag resource cannot be validated
	err := gt.Validate()
	if err != nil {
		return err
	}

	// Check if the provided tag (from source input value) already exists
	gt.versionFilter.Pattern = source
	tags, err := gt.nativeGitHandler.Tags(gt.spec.Path)
	if err != nil {
		return err
	}

	gt.foundVersion, err = gt.versionFilter.Search(tags)
	notFoundError := &version.ErrNoVersionFoundForPattern{Pattern: source}
	if err != nil && err.Error() != notFoundError.Error() {
		return err
	}

	if gt.foundVersion.GetVersion() == source {
		resultTarget.Information = source
		resultTarget.NewInformation = source
		resultTarget.Result = result.SUCCESS
		resultTarget.Description = fmt.Sprintf("git tag %q already exists", source)
		return nil
	}

	resultTarget.NewInformation = source
	resultTarget.Result = result.ATTENTION
	resultTarget.Changed = true

	if dryRun {
		// Dry run: no changes to apply.
		// Return early without creating tag but notify that a change should be made.
		resultTarget.Description = fmt.Sprintf("git tag %q should be created", source)
		return nil
	}

	_, err = gt.nativeGitHandler.NewTag(source, gt.spec.Message, gt.spec.Path)
	if err != nil {
		return err
	}

	resultTarget.Description = fmt.Sprintf("git tag %q created", source)
	if gt.spec.Message != "" {
		resultTarget.Description = gt.spec.Message
	}

	return nil
}
