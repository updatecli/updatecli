package gitbranch

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Target creates and pushes a git tag based on the SCM configuration
func (gb *GitBranch) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) (err error) {

	if gb.spec.Path != "" && scm != nil {
		logrus.Warningf("Path setting value %q is overriding the scm configuration (value %q)",
			gb.spec.Path,
			scm.GetDirectory())
	}

	if gb.spec.URL != "" && scm != nil {
		logrus.Warningf("URL setting value %q is overriding the scm configuration (value %q)",
			gb.spec.URL,
			scm.GetURL())
	}

	if gb.spec.URL != "" {
		gb.directory, err = gb.clone()
		if err != nil {
			return err
		}
	} else if gb.spec.Path != "" {
		gb.directory = gb.spec.Path
	} else if scm != nil {
		gb.directory = scm.GetDirectory()
	}

	if gb.directory == "" {
		logrus.Errorf("unknown Git working directory. Did you specify one of `spec.URL`, `scmid` or a `spec.path`?")
		return fmt.Errorf("unknown Git working directory")
	}

	if gb.spec.SourceBranch == "" && scm == nil {
		return fmt.Errorf("source branch is required")
	}

	gb.branch = source
	if gb.spec.Branch != "" {
		gb.branch = gb.spec.Branch
	}

	resultTarget.NewInformation = gb.branch

	if gb.branch == "" {
		return fmt.Errorf("empty branch specified")
	}

	err = gb.target(dryRun, resultTarget)
	if err != nil {
		return err
	}

	if dryRun {
		// Dry run: no changes to apply.
		// Return early without creating branch but notify that a change should be made.
		return nil
	}

	if !resultTarget.Changed {
		resultTarget.Description = fmt.Sprintf(
			"the git branch %q already exist on the specified remote.",
			gb.branch,
		)
		return nil
	}

	logrus.Printf("git branch %q has been created.", gb.branch)

	switch scm {
	case nil:
		if err = gb.nativeGitHandler.Checkout(gb.spec.Username, gb.spec.Password, gb.spec.SourceBranch, gb.branch, gb.directory, false, gb.spec.Depth); err != nil {
			logrus.Errorf("Git checkout branch error: %s", err)
			return err
		}

		if err = gb.nativeGitHandler.PushBranch(gb.branch, gb.spec.Username, gb.spec.Password, gb.directory, false); err != nil {
			logrus.Errorf("Git push branch error: %s", err)
			return err
		}
	default:

		sourceBranch, _, _ := scm.GetBranches()
		// Not specifying a username/password won't be an issue as it's only used to pull changes when the git branch already
		// ecists on the remote. In this case, we already know that it doesn't.
		// That being said, we may have a racing issue if the branch is created between the time Updatecli executed and the time
		// this code is executed so the current execution would fail but not then next one.
		if err = gb.nativeGitHandler.Checkout("", "", sourceBranch, gb.branch, gb.directory, false, gb.spec.Depth); err != nil {
			logrus.Errorf("Git checkout branch error: %s", err)
			return err
		}

		if err = scm.PushBranch(gb.branch); err != nil {
			logrus.Errorf("Git push tag error: %s", err)
			return err
		}
	}

	resultTarget.Description = fmt.Sprintf("git branch %q created and pushed", gb.branch)

	return nil
}

func (gt *GitBranch) target(dryRun bool, resultTarget *result.Target) error {

	// cfr https://github.com/updatecli/updatecli/issues/1126
	// to know why the following line is needed at the moment
	resultTarget.Files = []string{""}

	// Check if the provided branch (from source input value) already exists
	branches, err := gt.nativeGitHandler.Branches(gt.directory)
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
		resultTarget.Information = gt.branch
		resultTarget.Description = fmt.Sprintf("git branch %q already exists", gt.branch)
		return nil
	}

	resultTarget.Result = result.ATTENTION
	resultTarget.Changed = true
	resultTarget.Information = ""

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
