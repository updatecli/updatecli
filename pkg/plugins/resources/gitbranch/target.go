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

	changed, _, _, err = gt.target(dryRun)

	if !changed {
		logrus.Infof("%s The git branch %q already exist on the specified remote.", result.SUCCESS, gt.branch)
		return changed, err
	}
	logrus.Printf("%s The git branch %q has been created.", result.ATTENTION, gt.branch)

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

	changed, files, message, err = gt.target(dryRun)
	if err != nil {
		return changed, files, message, err
	}

	if dryRun {
		// Dry run: no changes to apply.
		// Return early without creating branch but notify that a change should be made.
		return changed, files, message, nil
	}

	if !changed {
		logrus.Infof("%s The git branch %q already exist on the specified remote.", result.SUCCESS, gt.branch)
		return changed, files, message, err
	}

	logrus.Printf("git branch %q has been created.", gt.branch)

	err = scm.PushBranch(gt.branch)
	if err != nil {
		logrus.Errorf("Git push tag error: %s", err)
		return changed, files, message, err
	}
	logrus.Infof("%s The git branch %q was pushed successfully to the specified remote.", result.ATTENTION, gt.branch)
	return changed, files, message, err
}

func (gt *GitBranch) target(dryRun bool) (bool, []string, string, error) {

	files := []string{}
	message := ""

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
			gt.branch)
		return false, files, message, nil
	}

	// Otherwise proceed to create this new branch
	logrus.Printf("%s The git branch %q does not exist: creating it.", result.ATTENTION, gt.branch)

	if dryRun {
		// Dry run: no changes to apply.
		// Return early without creating branch but notify that a change should be made.
		return true, files, message, nil
	}

	changed, err := gt.nativeGitHandler.NewBranch(gt.branch, gt.spec.Path)
	if err != nil {
		return changed, files, message, err
	}

	logrus.Printf("%s The git branch %q has been created.", result.ATTENTION, gt.branch)

	message = fmt.Sprintf("Git branch %q has been created.", gt.branch)

	return changed, files, message, nil
}
