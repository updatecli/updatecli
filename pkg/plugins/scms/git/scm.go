package git

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

func (g *Git) GetBranches() (sourceBranch, workingBranch, targetBranch string) {
	sourceBranch = g.spec.Branch
	workingBranch = g.spec.Branch
	targetBranch = g.spec.Branch

	if g.workingBranch && len(g.pipelineID) > 0 {
		workingBranch = g.nativeGitHandler.SanitizeBranchName(
			strings.Join([]string{g.workingBranchPrefix, targetBranch, g.pipelineID}, g.workingBranchSeparator))
	}

	return sourceBranch, workingBranch, targetBranch
}

// CleanWorkingBranch checks if the working branch is diverged from the target branch
// and remove it if not.
func (g *Git) CleanWorkingBranch() (bool, error) {
	_, workingBranch, targetBranch := g.GetBranches()

	if workingBranch == targetBranch {
		logrus.Infof("Skipping cleaning working branch %q on %q (same as target branch)\n", workingBranch, g.GetURL())
		return false, nil
	}

	isSimilarBranch, err := g.nativeGitHandler.IsSimilarBranch(workingBranch, targetBranch, g.GetDirectory())
	if err != nil {
		return false, fmt.Errorf("failed to compare working branch %q with target branch %q: %w", workingBranch, targetBranch, err)
	}

	if isSimilarBranch {
		if err = g.nativeGitHandler.DeleteBranch(workingBranch, g.GetDirectory(), g.spec.Username, g.spec.Password); err != nil {
			return false, fmt.Errorf("failed to delete working branch %q from %q: %w", workingBranch, g.GetDirectory(), err)
		}
		return true, nil
	}

	return false, nil
}

// GetURL returns a git URL
func (g *Git) GetURL() string {
	return g.spec.URL
}

// Add run `git add`.
func (g *Git) Add(files []string) error {
	err := g.nativeGitHandler.Add(files, g.GetDirectory())
	if err != nil {
		return err
	}
	return nil
}

// Checkout create and then uses a temporary git branch.
func (g *Git) Checkout() error {
	sourceBranch, workingBranch, _ := g.GetBranches()

	err := g.nativeGitHandler.Checkout(
		g.spec.Username,
		g.spec.Password,
		sourceBranch,
		workingBranch,
		g.GetDirectory(),
		g.spec.Force)
	if err != nil {
		return err
	}
	return nil
}

// GetDirectory returns the working git directory.
func (g *Git) GetDirectory() (directory string) {
	return g.spec.Directory
}

// Clean removes the current git repository from local storage.
func (g *Git) Clean() error {
	err := os.RemoveAll(g.spec.Directory) // clean up
	if err != nil {
		return err
	}
	return nil
}

// Clone run `git clone`.
func (g *Git) Clone() (string, error) {
	err := g.nativeGitHandler.Clone(
		g.spec.Username,
		g.spec.Password,
		g.GetURL(),
		g.GetDirectory(),
		g.spec.Submodules,
	)
	if err != nil {
		logrus.Errorf("failed cloning git repository %q - %s", g.GetURL(), err)
		return "", err
	}

	return g.spec.Directory, nil
}

// Commit run `git commit`.
func (g *Git) Commit(message string) error {
	// Generate the conventional commit message
	commitMessage, err := g.spec.CommitMessage.Generate(message)
	if err != nil {
		return err
	}

	if err = g.nativeGitHandler.Commit(
		g.spec.User,
		g.spec.Email,
		commitMessage,
		g.GetDirectory(),
		g.spec.GPG.SigningKey,
		g.spec.GPG.Passphrase,
	); err != nil {
		return err
	}

	if g.spec.CommitMessage.IsSquash() {
		sourceBranch, workingBranch, _ := g.GetBranches()
		if err = g.nativeGitHandler.SquashCommit(g.GetDirectory(), sourceBranch, workingBranch, gitgeneric.SquashCommitOptions{
			IncludeCommitTitles: true,
			Message:             commitMessage,
			SigninKey:           g.spec.GPG.SigningKey,
			SigninPassphrase:    g.spec.GPG.Passphrase,
		}); err != nil {
			return err
		}
	}

	return nil
}

// Push run `git push`.
func (g *Git) Push() (bool, error) {
	return g.nativeGitHandler.Push(
		g.spec.Username,
		g.spec.Password,
		g.GetDirectory(),
		g.spec.Force)
}

// PushBranch push tags
func (g *Git) PushBranch(branch string) error {
	err := g.nativeGitHandler.PushBranch(
		branch,
		g.spec.Username,
		g.spec.Password,
		g.GetDirectory(),
		g.spec.Force)
	if err != nil {
		return err
	}

	return nil
}

// IsRemoteBranchUpToDate checks if the local working branch is up to date with the remote branch.
func (g *Git) IsRemoteBranchUpToDate() (bool, error) {
	sourceBranch, workingBranch, _ := g.GetBranches()

	return g.nativeGitHandler.IsLocalBranchSyncedWithRemote(
		sourceBranch,
		workingBranch,
		g.spec.Username,
		g.spec.Password,
		g.GetDirectory())
}

// IsRemoteWorkingBranchExist checks if the remote working branch exists.
func (g *Git) IsRemoteWorkingBranchExist() (bool, error) {
	_, workingBranch, _ := g.GetBranches()

	return g.nativeGitHandler.IsRemoteBranchExist(
		workingBranch,
		g.spec.Username,
		g.spec.Password,
		g.GetDirectory())
}

// PushTag push tags
func (g *Git) PushTag(tag string) error {
	err := g.nativeGitHandler.PushTag(
		tag,
		g.spec.Username,
		g.spec.Password,
		g.GetDirectory(),
		g.spec.Force)
	if err != nil {
		return err
	}

	return nil
}

func (g *Git) GetChangedFiles(workingDir string) ([]string, error) {
	return g.nativeGitHandler.GetChangedFiles(workingDir)
}
