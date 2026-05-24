package tangled

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// GetBranches returns the source, working and target branches.
func (t *Tangled) GetBranches() (sourceBranch, workingBranch, targetBranch string) {
	sourceBranch = t.Spec.Branch
	workingBranch = t.Spec.Branch
	targetBranch = t.Spec.Branch

	if len(t.pipelineID) > 0 && t.workingBranch {
		workingBranch = t.nativeGitHandler.SanitizeBranchName(
			strings.Join([]string{t.workingBranchPrefix, targetBranch, t.pipelineID}, t.workingBranchSeparator))
	}

	return sourceBranch, workingBranch, targetBranch
}

// CleanWorkingBranch checks if the working branch is diverged from the target branch
// and remove it if not.
func (t *Tangled) CleanWorkingBranch() (bool, error) {
	_, workingBranch, targetBranch := t.GetBranches()

	if workingBranch == targetBranch {
		logrus.Infof("Skipping cleaning working branch %q on %q (same as target branch)\n", workingBranch, t.GetURL())
		return false, nil
	}

	isSimilar, err := t.nativeGitHandler.IsSimilarBranch(workingBranch, targetBranch, t.GetDirectory())
	if err != nil {
		return false, fmt.Errorf("failed to compare working branch %q with target branch %q: %w", workingBranch, targetBranch, err)
	}

	if isSimilar {
		if err = t.nativeGitHandler.DeleteBranch(workingBranch, t.GetDirectory(), "", ""); err != nil {
			return false, fmt.Errorf("failed to delete working branch %q from %q: %w", workingBranch, t.GetDirectory(), err)
		}
		return true, nil
	}

	return false, nil
}

// GetURL returns a "Tangled" git URL.
//
// Tangled knots reject HTTPS pushes, so SSH is used for both clone and push.
// The default-knot hostname (knot1.tangled.sh) is rewritten to tangled.org
// because that is where the SSH gateway lives.
func (t *Tangled) GetURL() string {
	if t.Spec.CloneURL != "" {
		return t.Spec.CloneURL
	}
	sshHost := t.Spec.Knot
	if sshHost == "knot1.tangled.sh" || sshHost == "" {
		sshHost = "tangled.org"
	}
	return fmt.Sprintf("git@%s:%s/%s", sshHost, t.Spec.Owner, t.Spec.Repository)
}

// GetDirectory returns the local git repository path.
func (t *Tangled) GetDirectory() string {
	return t.Spec.Directory
}

// Clean deletes the Tangled working directory.
func (t *Tangled) Clean() error {
	return os.RemoveAll(t.Spec.Directory)
}

// Clone runs `git clone`.
func (t *Tangled) Clone() (string, error) {
	t.setDirectory()

	if err := t.nativeGitHandler.Clone(
		"",
		"",
		t.GetURL(),
		t.GetDirectory(),
		t.Spec.Submodules,
		t.Spec.Depth,
	); err != nil {
		logrus.Errorf("failed cloning Tangled repository %q", t.GetURL())
		return "", err
	}

	return t.Spec.Directory, nil
}

// Commit runs `git commit`.
func (t *Tangled) Commit(_ context.Context, message string) error {
	commitMessage, err := t.Spec.CommitMessage.Generate(message)
	if err != nil {
		return err
	}

	if err := t.nativeGitHandler.Commit(t.Spec.User, t.Spec.Email, commitMessage, t.GetDirectory(), t.Spec.GPG.SigningKey, t.Spec.GPG.Passphrase); err != nil {
		return err
	}

	if t.Spec.CommitMessage.IsSquash() {
		sourceBranch, workingBranch, _ := t.GetBranches()
		if err = t.nativeGitHandler.SquashCommit(t.GetDirectory(), sourceBranch, workingBranch, gitgeneric.SquashCommitOptions{
			IncludeCommitTitles: true,
			Message:             commitMessage,
			SigninKey:           t.Spec.GPG.SigningKey,
			SigninPassphrase:    t.Spec.GPG.Passphrase,
		}); err != nil {
			return err
		}
	}

	return nil
}

// Checkout create and then uses a temporary git branch.
func (t *Tangled) Checkout() error {
	sourceBranch, workingBranch, _ := t.GetBranches()
	return t.nativeGitHandler.Checkout(
		"",
		"",
		sourceBranch,
		workingBranch,
		t.Spec.Directory,
		t.force,
		t.Spec.Depth,
	)
}

// Add runs `git add`.
func (t *Tangled) Add(files []string) error {
	return t.nativeGitHandler.Add(files, t.Spec.Directory)
}

// IsRemoteBranchUpToDate checks if the local working branch is up to date with the remote branch.
func (t *Tangled) IsRemoteBranchUpToDate() (bool, error) {
	sourceBranch, workingBranch, _ := t.GetBranches()
	return t.nativeGitHandler.IsLocalBranchSyncedWithRemote(
		sourceBranch,
		workingBranch,
		"",
		"",
		t.GetDirectory(),
	)
}

// IsRemoteWorkingBranchExist checks if the remote working branch exists.
func (t *Tangled) IsRemoteWorkingBranchExist() (bool, error) {
	_, workingBranch, _ := t.GetBranches()
	return t.nativeGitHandler.IsRemoteBranchExist(
		workingBranch,
		"",
		"",
		t.GetDirectory(),
	)
}

// Push runs `git push` to the corresponding Tangled remote branch if not already created.
func (t *Tangled) Push() (bool, error) {
	return t.nativeGitHandler.Push(
		"",
		"",
		t.GetDirectory(),
		t.force,
	)
}

// PushTag push tags.
func (t *Tangled) PushTag(tag string) error {
	return t.nativeGitHandler.PushTag(tag, "", "", t.GetDirectory(), t.force)
}

// PushBranch push branch.
func (t *Tangled) PushBranch(branch string) error {
	return t.nativeGitHandler.PushBranch(branch, "", "", t.GetDirectory(), t.force)
}

// GetChangedFiles returns the list of files changed in the working directory.
func (t *Tangled) GetChangedFiles(workingDir string) ([]string, error) {
	return t.nativeGitHandler.GetChangedFiles(workingDir)
}
