package gitea

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

func (g *Gitea) GetBranches() (sourceBranch, workingBranch, targetBranch string) {
	sourceBranch = g.Spec.Branch
	workingBranch = g.Spec.Branch
	targetBranch = g.Spec.Branch

	if len(g.pipelineID) > 0 && g.workingBranch {
		workingBranch = g.nativeGitHandler.SanitizeBranchName(
			strings.Join([]string{g.workingBranchPrefix, targetBranch, g.pipelineID}, g.workingBranchSeparator))
	}

	return sourceBranch, workingBranch, targetBranch
}

// CleanWorkingBranch checks if the working branch is diverged from the target branch
// and remove it if not.
func (g *Gitea) CleanWorkingBranch() (bool, error) {
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
		if err = g.nativeGitHandler.DeleteBranch(workingBranch, g.GetDirectory(), g.Spec.Username, g.Spec.Token); err != nil {
			return false, fmt.Errorf("failed to delete working branch %q from %q: %w", workingBranch, g.GetDirectory(), err)
		}
		return true, nil
	}

	return false, nil
}

// GetURL returns a "Gitea" git URL
func (g *Gitea) GetURL() string {
	URL := fmt.Sprintf("%v/%v/%v.git",
		g.Spec.URL,
		g.Spec.Owner,
		g.Spec.Repository)

	return URL
}

// GetDirectory returns the local git repository path.
func (g *Gitea) GetDirectory() (directory string) {
	return g.Spec.Directory
}

// Clean deletes github working directory.
func (g *Gitea) Clean() error {
	err := os.RemoveAll(g.Spec.Directory)
	if err != nil {
		return err
	}
	return nil
}

// Clone run `git clone`.
func (g *Gitea) Clone() (string, error) {
	g.setDirectory()

	err := g.nativeGitHandler.Clone(
		g.Spec.Username,
		g.Spec.Token,
		g.GetURL(),
		g.GetDirectory(),
		g.Spec.Submodules,
	)
	if err != nil {
		logrus.Errorf("failed cloning Gitea repository %q", g.GetURL())
		return "", err
	}

	return g.Spec.Directory, nil
}

// Commit run `git commit`.
func (g *Gitea) Commit(message string) error {
	// Generate the conventional commit message
	commitMessage, err := g.Spec.CommitMessage.Generate(message)
	if err != nil {
		return err
	}

	err = g.nativeGitHandler.Commit(g.Spec.User, g.Spec.Email, commitMessage, g.GetDirectory(), g.Spec.GPG.SigningKey, g.Spec.GPG.Passphrase)
	if err != nil {
		return err
	}

	if g.Spec.CommitMessage.IsSquash() {
		sourceBranch, workingBranch, _ := g.GetBranches()
		if err = g.nativeGitHandler.SquashCommit(g.GetDirectory(), sourceBranch, workingBranch, gitgeneric.SquashCommitOptions{
			IncludeCommitTitles: true,
			Message:             commitMessage,
			SigninKey:           g.Spec.GPG.SigningKey,
			SigninPassphrase:    g.Spec.GPG.Passphrase,
		}); err != nil {
			return err
		}
	}

	return nil
}

// Checkout create and then uses a temporary git branch.
func (g *Gitea) Checkout() error {
	sourceBranch, workingBranch, _ := g.GetBranches()

	err := g.nativeGitHandler.Checkout(
		g.Spec.Username,
		g.Spec.Token,
		sourceBranch,
		workingBranch,
		g.Spec.Directory,
		g.force)
	if err != nil {
		return err
	}
	return nil
}

// Add run `git add`.
func (g *Gitea) Add(files []string) error {
	err := g.nativeGitHandler.Add(files, g.Spec.Directory)
	if err != nil {
		return err
	}
	return nil
}

// IsRemoteBranchUpToDate checks if the local working branch is up to date with the remote branch.
func (g *Gitea) IsRemoteBranchUpToDate() (bool, error) {
	sourceBranch, workingBranch, _ := g.GetBranches()

	return g.nativeGitHandler.IsLocalBranchSyncedWithRemote(
		sourceBranch,
		workingBranch,
		g.Spec.Username,
		g.Spec.Token,
		g.GetDirectory())
}

// IsRemoteWorkingBranchExist checks if the remote working branch exists.
func (g *Gitea) IsRemoteWorkingBranchExist() (bool, error) {
	_, workingBranch, _ := g.GetBranches()

	return g.nativeGitHandler.IsRemoteBranchExist(
		workingBranch,
		g.Spec.Username,
		g.Spec.Token,
		g.GetDirectory())
}

// Push run `git push` to the corresponding Gitea remote branch if not already created.
func (g *Gitea) Push() (bool, error) {
	return g.nativeGitHandler.Push(
		g.Spec.Username,
		g.Spec.Token,
		g.GetDirectory(),
		g.force,
	)
}

// PushTag push tags
func (g *Gitea) PushTag(tag string) error {
	err := g.nativeGitHandler.PushTag(
		tag,
		g.Spec.Username,
		g.Spec.Token,
		g.GetDirectory(),
		g.force)
	if err != nil {
		return err
	}

	return nil
}

// PushBranch push branch
func (g *Gitea) PushBranch(branch string) error {
	err := g.nativeGitHandler.PushTag(
		branch,
		g.Spec.Username,
		g.Spec.Token,
		g.GetDirectory(),
		g.force)
	if err != nil {
		return err
	}

	return nil
}

func (g *Gitea) GetChangedFiles(workingDir string) ([]string, error) {
	return g.nativeGitHandler.GetChangedFiles(workingDir)
}
