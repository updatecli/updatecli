package gitbranch

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Target creates and pushes a git tag based on the SCM configuration
func (gt *GitBranch) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) (err error) {

	if scm != nil {
		if len(gt.spec.Path) > 0 {
			logrus.Warningf("Path setting value %q overridden by the scm configuration (value %q)",
				gt.spec.Path,
				scm.GetDirectory())
		}

		gt.spec.Path = scm.GetDirectory()
	}

	gt.branch = source
	if gt.spec.Branch != "" {
		gt.branch = gt.spec.Branch
	}

	resultTarget.NewInformation = gt.branch

	if gt.branch == "" {
		return fmt.Errorf("empty branch specified")
	}

	err = gt.target(dryRun, resultTarget)
	if err != nil {
		return err
	}

	if dryRun {
		// Dry run: no changes to apply.
		// Return early without creating branch but notify that a change should be made.
		return nil
	}

	if !resultTarget.Changed {
		resultTarget.Description = fmt.Sprintf("the git branch %q already exist on the specified remote.", gt.branch)
		return nil
	}

	logrus.Printf("git branch %q has been created.", gt.branch)

	if scm == nil {
		resultTarget.Description = fmt.Sprintf("The git branch %q created but missing scm configuration to push it", gt.branch)
		return nil
	}

	err = scm.PushBranch(gt.branch)
	if err != nil {
		logrus.Errorf("Git push tag error: %s", err)
		return err
	}

	resultTarget.Description = fmt.Sprintf("git branch %q created and pushed", gt.branch)
	return nil
}

func (gt *GitBranch) target(dryRun bool, resultTarget *result.Target) error {

	// cfr https://github.com/updatecli/updatecli/issues/1126
	// to know why the following line is needed at the moment
	resultTarget.Files = []string{""}

	// Fail if the git tag resource cannot be validated
	err := gt.Validate()
	if err != nil {
		logrus.Errorln(err)
		return err
	}

	// Check if the provided branch (from source input value) already exists
	branches, err := gt.nativeGitHandler.Branches(gt.spec.Path)
	if err != nil {
		return err
	}

	found := false
	for _, b := range branches {
		if b == gt.branch {
			found = true
			break
		}
	}

	resultTarget.NewInformation = gt.branch

	if found {
		resultTarget.Result = result.SUCCESS
		resultTarget.OldInformation = gt.branch
		resultTarget.Description = fmt.Sprintf("git branch %q already exists", gt.branch)
		return nil
	}

	resultTarget.Result = result.ATTENTION
	resultTarget.Changed = true
	resultTarget.OldInformation = ""

	logrus.Debugf("git branch %q does not exist: creating it.", gt.branch)

	if dryRun {
		// Dry run: no changes to apply.
		// Return early without creating branch but notify that a change should be made.
		resultTarget.Description = fmt.Sprintf("git branch %q should be created", gt.branch)
		return nil
	}

	resultTarget.Changed, err = gt.nativeGitHandler.NewBranch(gt.branch, gt.spec.Path)
	if err != nil {
		return err
	}

	resultTarget.Description = fmt.Sprintf("git branch %q created.", gt.branch)

	return nil
}
