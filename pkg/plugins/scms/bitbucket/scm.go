package bitbucket

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/bitbucket/client"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

func (b *Bitbucket) GetBranches() (sourceBranch, workingBranch, targetBranch string) {
	sourceBranch = b.Spec.Branch
	workingBranch = b.Spec.Branch
	targetBranch = b.Spec.Branch

	if len(b.pipelineID) > 0 && b.workingBranch {
		workingBranch = b.nativeGitHandler.SanitizeBranchName(
			strings.Join([]string{b.workingBranchPrefix, targetBranch, b.pipelineID}, b.workingBranchSeparator))
	}

	return sourceBranch, workingBranch, targetBranch
}

// CleanWorkingBranch checks if the working branch is diverged from the target branch
// and remove it if not.
func (b *Bitbucket) CleanWorkingBranch() (bool, error) {
	_, workingBranch, targetBranch := b.GetBranches()

	if workingBranch == targetBranch {
		logrus.Infof("Skipping cleaning working branch %q on %q (same as target branch)\n", workingBranch, b.GetURL())
		return false, nil
	}

	isSimilarBranch, err := b.nativeGitHandler.IsSimilarBranch(workingBranch, targetBranch, b.GetDirectory())
	if err != nil {
		return false, fmt.Errorf("failed to compare working branch %q with target branch %q: %w", workingBranch, targetBranch, err)
	}

	if isSimilarBranch {
		if err = b.nativeGitHandler.DeleteBranch(workingBranch, b.GetDirectory(), b.GetUsername(), b.GetPassword()); err != nil {
			return false, fmt.Errorf("failed to delete working branch %q from %q: %w", workingBranch, b.GetDirectory(), err)
		}
		return true, nil
	}

	return false, nil
}

// GetURL returns a "Stash" git URL
func (b *Bitbucket) GetURL() string {
	URL := fmt.Sprintf("%v/%v/%v",
		client.URL(),
		b.Spec.Owner,
		b.Spec.Repository)

	return URL
}

func (b *Bitbucket) GetUsername() string {
	if len(b.Spec.Token) > 0 {
		return "x-token-auth"
	} else {
		return b.Spec.Username
	}
}

func (b *Bitbucket) GetPassword() string {
	if len(b.Spec.Token) > 0 {
		return b.Spec.Token
	} else {
		return b.Spec.Password
	}
}

// GetDirectory returns the local git repository path.
func (b *Bitbucket) GetDirectory() (directory string) {
	return b.Spec.Directory
}

// Clean deletes github working directory.
func (b *Bitbucket) Clean() error {
	err := os.RemoveAll(b.Spec.Directory)
	if err != nil {
		return err
	}
	return nil
}

// Clone run `git clone`.
func (b *Bitbucket) Clone() (string, error) {
	b.setDirectory()

	err := b.nativeGitHandler.Clone(
		b.GetUsername(),
		b.GetPassword(),
		b.GetURL(),
		b.GetDirectory(),
		b.Spec.Submodules,
	)
	if err != nil {
		logrus.Errorf("failed cloning Bitbucket Cloud repository %q", b.GetURL())
		return "", err
	}

	return b.Spec.Directory, nil
}

// Commit run `git commit`.
func (b *Bitbucket) Commit(message string) error {
	// Generate the conventional commit message
	commitMessage, err := b.Spec.CommitMessage.Generate(message)
	if err != nil {
		return err
	}

	err = b.nativeGitHandler.Commit(b.Spec.User, b.Spec.Email, commitMessage, b.GetDirectory(), b.Spec.GPG.SigningKey, b.Spec.GPG.Passphrase)
	if err != nil {
		return err
	}

	if b.Spec.CommitMessage.IsSquash() {
		sourceBranch, workingBranch, _ := b.GetBranches()
		if err = b.nativeGitHandler.SquashCommit(b.GetDirectory(), sourceBranch, workingBranch, gitgeneric.SquashCommitOptions{
			IncludeCommitTitles: true,
			Message:             commitMessage,
			SigninKey:           b.Spec.GPG.SigningKey,
			SigninPassphrase:    b.Spec.GPG.Passphrase,
		}); err != nil {
			return err
		}
	}

	return nil
}

// Checkout create and then uses a temporary git branch.
func (b *Bitbucket) Checkout() error {
	sourceBranch, workingBranch, _ := b.GetBranches()

	err := b.nativeGitHandler.Checkout(
		b.GetUsername(),
		b.GetPassword(),
		sourceBranch,
		workingBranch,
		b.Spec.Directory,
		b.force)
	if err != nil {
		return err
	}
	return nil
}

// Add run `git add`.
func (b *Bitbucket) Add(files []string) error {
	err := b.nativeGitHandler.Add(files, b.Spec.Directory)
	if err != nil {
		return err
	}
	return nil
}

// IsRemoteBranchUpToDate checks if the local branch is up to date with the remote branch.
func (b *Bitbucket) IsRemoteBranchUpToDate() (bool, error) {
	sourceBranch, workingBranch, _ := b.GetBranches()

	return b.nativeGitHandler.IsLocalBranchSyncedWithRemote(
		sourceBranch,
		workingBranch,
		b.GetUsername(),
		b.GetPassword(),
		b.GetDirectory())
}

// IsRemoteWorkingBranchExist checks if the remote working branch exists.
func (b *Bitbucket) IsRemoteWorkingBranchExist() (bool, error) {
	_, workingBranch, _ := b.GetBranches()

	return b.nativeGitHandler.IsRemoteBranchExist(
		workingBranch,
		b.GetUsername(),
		b.GetPassword(),
		b.GetDirectory())
}

// Push run `git push` to the corresponding Bitbucket Server remote branch if not already created.
func (b *Bitbucket) Push() (bool, error) {
	return b.nativeGitHandler.Push(
		b.GetUsername(),
		b.GetPassword(),
		b.GetDirectory(),
		b.force)
}

// PushTag push tags
func (b *Bitbucket) PushTag(tag string) error {
	err := b.nativeGitHandler.PushTag(
		tag,
		b.GetUsername(),
		b.GetPassword(),
		b.GetDirectory(),
		b.force,
	)
	if err != nil {
		return err
	}

	return nil
}

// PushBranch push branch
func (b *Bitbucket) PushBranch(branch string) error {
	err := b.nativeGitHandler.PushTag(
		branch,
		b.GetUsername(),
		b.GetPassword(),
		b.GetDirectory(),
		b.force)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bitbucket) GetChangedFiles(workingDir string) ([]string, error) {
	return b.nativeGitHandler.GetChangedFiles(workingDir)
}
