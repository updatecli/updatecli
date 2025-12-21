package gitlab

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitlab/client"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// GetBranches returns the source, working and target branches.
func (g *Gitlab) GetBranches() (sourceBranch, workingBranch, targetBranch string) {
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
func (g *Gitlab) CleanWorkingBranch() (bool, error) {
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

// GetURL returns a "GitLab" git URL
func (g *Gitlab) GetURL() string {
	url := client.EnsureValidURL(g.Spec.URL)

	URL := fmt.Sprintf("%s/%s/%s.git",
		url,
		g.Spec.Owner,
		g.Spec.Repository)

	return URL
}

// GetDirectory returns the local git repository path.
func (g *Gitlab) GetDirectory() (directory string) {
	return g.Spec.Directory
}

// Clean deletes github working directory.
func (g *Gitlab) Clean() error {
	err := os.RemoveAll(g.Spec.Directory)
	if err != nil {
		return err
	}
	return nil
}

// Clone run `git clone`.
func (g *Gitlab) Clone() (string, error) {
	g.setDirectory()

	err := g.nativeGitHandler.Clone(
		g.Spec.Username,
		g.Spec.Token,
		g.GetURL(),
		g.GetDirectory(),
		g.Spec.Submodules,
	)
	if err != nil {
		logrus.Errorf("failed cloning GitLab repository %q", g.GetURL())
		return "", err
	}

	return g.Spec.Directory, nil
}

// Commit run `git commit`.
func (g *Gitlab) Commit(message string) error {
	// Generate the conventional commit message
	commitMessage, err := g.Spec.CommitMessage.Generate(message)
	if err != nil {
		return err
	}

	err = g.nativeGitHandler.Commit(
		g.Spec.User,
		g.Spec.Email,
		commitMessage,
		g.GetDirectory(),
		g.Spec.GPG.SigningKey,
		g.Spec.GPG.Passphrase,
	)
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
func (g *Gitlab) Checkout() error {
	sourceBranch, workingBranch, _ := g.GetBranches()

	err := g.nativeGitHandler.Checkout(
		g.Spec.Username,
		g.Spec.Token,
		sourceBranch,
		workingBranch,
		g.Spec.Directory,
		g.force,
	)
	if err != nil {
		return err
	}
	return nil
}

// Add run `git add`.
func (g *Gitlab) Add(files []string) error {
	err := g.nativeGitHandler.Add(files, g.Spec.Directory)
	if err != nil {
		return err
	}
	return nil
}

// IsRemoteBranchUpToDate checks if the branch reference name is published on
// on the default remote
func (g *Gitlab) IsRemoteBranchUpToDate() (bool, error) {
	sourceBranch, workingBranch, _ := g.GetBranches()

	return g.nativeGitHandler.IsLocalBranchSyncedWithRemote(
		sourceBranch,
		workingBranch,
		g.Spec.Username,
		g.Spec.Token,
		g.GetDirectory())
}

// IsRemoteWorkingBranchExist checks if the branch reference name is published on
// on the default remote
func (g *Gitlab) IsRemoteWorkingBranchExist() (bool, error) {
	_, workingBranch, _ := g.GetBranches()

	return g.nativeGitHandler.IsRemoteBranchExist(
		workingBranch,
		g.Spec.Username,
		g.Spec.Token,
		g.GetDirectory())
}

// Push run `git push` to the corresponding GitLab remote branch if not already created.
func (g *Gitlab) Push() (bool, error) {
	return g.nativeGitHandler.Push(
		g.Spec.Username,
		g.Spec.Token,
		g.GetDirectory(),
		g.force,
	)
}

// PushTag push tags
func (g *Gitlab) PushTag(tag string) error {
	err := g.nativeGitHandler.PushTag(
		tag,
		g.Spec.Username,
		g.Spec.Token,
		g.GetDirectory(),
		g.force,
	)
	if err != nil {
		return err
	}

	return nil
}

// PushBranch push branch
func (g *Gitlab) PushBranch(branch string) error {
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

// GetChangedFiles returns a list of changed files
func (g *Gitlab) GetChangedFiles(workingDir string) ([]string, error) {
	return g.nativeGitHandler.GetChangedFiles(workingDir)
}
