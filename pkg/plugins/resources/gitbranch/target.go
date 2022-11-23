package gitbranch

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Target creates a tag if needed from a local git repository, without pushing the tag
func (gt *GitBranch) Target(source string, dryRun bool) (changed bool, err error) {

	gt.branch = source
	if gt.spec.Branch != "" {
		gt.branch = gt.spec.Branch
	}

	if gt.branch == "" {
		return false, fmt.Errorf("empty branch specified")
	}

	changed, _, _, err = gt.target(gt.branch, dryRun)

	return changed, err
}

// TargetFromSCM creates and pushes a git tag based on the SCM configuration
func (gt *GitBranch) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (changed bool, files []string, message string, err error) {
	if len(gt.spec.Path) > 0 {
		logrus.Warningf("Path setting value %q overridden by the scm configuration (value %q)",
			gt.spec.Path,
			scm.GetDirectory())
	}
	gt.spec.Path = scm.GetDirectory()

	gt.branch = source
	if gt.spec.Branch != "" {
		gt.branch = gt.spec.Branch
	}

	if gt.branch == "" {
		return changed, files, message, fmt.Errorf("empty branch specified")
	}

	changed, files, message, err = gt.target(gt.branch, dryRun)
	if err != nil {
		return changed, files, message, err
	}

	err = scm.PushBranch(gt.branch)
	if err != nil {
		logrus.Errorf("Git push tag error: %s", err)
		return changed, files, message, err
	}
	logrus.Infof("%s The git branch %q was pushed successfully to the specified remote.", result.ATTENTION, gt.branch)
	return changed, files, message, err
}

func (gt *GitBranch) target(source string, dryRun bool) (bool, []string, string, error) {

	files := []string{}
	message := ""

	// Fail if a pattern is specified
	if gt.versionFilter.Pattern != "" {
		return false, files, message, fmt.Errorf("target validation error: spec.versionfilter.pattern is not allowed for targets of type gittag")
	}

	// Fail if the git tag resource cannot be validated
	err := gt.Validate()
	if err != nil {
		logrus.Errorln(err)
		return false, files, message, err
	}

	// Check if the provided branch (from source input value) already exists
	branches, err := gt.nativeGitHandler.Branches(gt.spec.Path)
	if err != nil {
		return false, files, message, err
	}

	found := false
	for _, b := range branches {
		if b == gt.branch {
			found = true
		}
	}

	if found {
		// No error, but no change
		logrus.Printf("%s The Git branch %q already exists, nothing else to do.",
			result.SUCCESS,
			source)
		return false, files, message, nil
	}

	// Otherwise proceed to create this new branch
	logrus.Printf("%s The git branch %q does not exist: creating it.", result.ATTENTION, gt.branch)

	if dryRun {
		// Dry run: no changes to apply.
		// Return early without creating tag but notify that a change should be made.
		return true, files, message, nil
	}

	changed, err := gt.nativeGitHandler.NewBranch(gt.branch, gt.spec.Path)
	if err != nil {
		return changed, files, message, err
	}
	logrus.Printf("%s The git branch %q has been created.", result.ATTENTION, gt.branch)

	return changed, files, message, nil
}
