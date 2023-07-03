package gittag

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Target creates a tag if needed from a local git repository, without pushing the tag
func (gt *GitTag) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {
	var err error

	if gt.spec.Path != "" && scm != nil {
		logrus.Warningf("Path setting value %q is overriding the scm configuration (value %q)",
			gt.spec.Path,
			scm.GetDirectory())
	}

	if gt.spec.URL != "" && scm != nil {
		logrus.Warningf("URL setting value %q is overriding the scm configuration (value %q)",
			gt.spec.URL,
			scm.GetURL())
	}

	if gt.spec.URL != "" {
		gt.directory, err = gt.clone()
		if err != nil {
			return err
		}
	} else if gt.spec.Path != "" {
		gt.directory = gt.spec.Path
	} else if scm != nil {
		gt.directory = scm.GetDirectory()
	}

	err = gt.Validate()
	if err != nil {
		return err
	}

	if gt.directory == "" {
		return fmt.Errorf("Unkownn Git working directory. Did you specify one of `URL`, `scmID`, or `spec.path`?")
	}

	tagName := source
	if gt.spec.VersionFilter.Pattern != "" {
		tagName = gt.spec.VersionFilter.Pattern
	}

	if err := gt.target(tagName, dryRun, resultTarget); err != nil {
		return err
	}

	if dryRun || !resultTarget.Changed {
		return nil
	}

	if gt.spec.URL != "" || gt.spec.Path != "" {
		if err = gt.nativeGitHandler.Checkout(gt.spec.Username, gt.spec.Password, gt.spec.SourceBranch, gt.spec.SourceBranch, gt.directory, false); err != nil {
			logrus.Errorf("Git checkout branch error: %s", err)
			return err
		}

		if err := gt.nativeGitHandler.PushTag(tagName, gt.spec.Username, gt.spec.Password, gt.directory, false); err != nil {
			logrus.Errorf("Git push tag error: %s", err)
			return err
		}
	} else if scm != nil {

		sourceBranch, _, _ := scm.GetBranches()
		// Not specifying a username/password won't be an issue as it's only used to pull changes when the git branch already
		// ecists on the remote. In this case, we already know that it doesn't.
		// That being said, we may have a racing issue if the branch is created between the time Updatecli executed and the time
		// this code is executed so the current execution would fail but not then next one.
		if err = gt.nativeGitHandler.Checkout("", "", sourceBranch, sourceBranch, sourceBranch, false); err != nil {
			logrus.Errorf("Git checkout branch error: %s", err)
			return err
		}

		if err := scm.PushTag(tagName); err != nil {
			logrus.Errorf("Git push tag error: %s", err)
			return err
		}
	}

	resultTarget.Description = fmt.Sprintf(
		"git tag %q successfully created and pushed",
		tagName,
	)

	resultTarget.Description = fmt.Sprintf("git tag %q successfully created", source)
	if gt.spec.Message != "" {
		resultTarget.Description = gt.spec.Message
	}

	return nil
}

func (gt *GitTag) target(tagName string, dryRun bool, resultTarget *result.Target) error {

	if tagName == "" {
		return fmt.Errorf("no tag specify")
	}

	// Ensure that a git message is present to annotate the tag to create
	if len(gt.spec.Message) == 0 {
		// absence of a message is not blocking: warn the user and continue
		gt.spec.Message = "Generated by updatecli"
		logrus.Warningf("No specified message for gittag target. Using default value %q", gt.spec.Message)
	}

	// cfr https://github.com/updatecli/updatecli/issues/1126
	// to know why the following line is needed at the moment
	resultTarget.Files = []string{""}

	// Fail if the git tag resource cannot be validated
	err := gt.Validate()
	if err != nil {
		return err
	}

	// Check if the provided tag (from source input value) already exists
	gt.versionFilter.Pattern = tagName
	tags, err := gt.nativeGitHandler.Tags(gt.directory)
	if err != nil && !strings.Contains(err.Error(), "no tag found") {

		logrus.Errorf("Error while searching for tags: %s", err)
		return err
	}

	// Some tags have been found so we can search for the tag name if it already exists
	if len(tags) > 0 {
		gt.foundVersion, err = gt.versionFilter.Search(tags)
		notFoundError := &version.ErrNoVersionFoundForPattern{Pattern: tagName}
		if err != nil && err.Error() != notFoundError.Error() {
			return err
		}

		if gt.foundVersion.GetVersion() == tagName {
			resultTarget.Information = tagName
			resultTarget.NewInformation = tagName
			resultTarget.Result = result.SUCCESS
			resultTarget.Description = fmt.Sprintf("git tag %q already exists", tagName)
			return nil
		}

	}

	resultTarget.NewInformation = tagName
	resultTarget.Result = result.ATTENTION
	resultTarget.Changed = true

	if dryRun {
		// Dry run: no changes to apply.
		// Return early without creating tag but notify that a change should be made.
		resultTarget.Description = fmt.Sprintf("git tag %q should be created", tagName)
		return nil
	}

	_, err = gt.nativeGitHandler.NewTag(tagName, gt.spec.Message, gt.directory)
	if err != nil {
		return err
	}

	resultTarget.Description = fmt.Sprintf("git tag %q created", tagName)
	if gt.spec.Message != "" {
		resultTarget.Description = gt.spec.Message
	}

	return nil
}
