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

// IsRemoteBranchUpToDate checks if the branch reference name is published on
// on the default remote
func (g *Gitea) IsRemoteBranchUpToDate() (bool, error) {
	sourceBranch, workingBranch, _ := g.GetBranches()

	return g.nativeGitHandler.IsLocalBranchPublished(
		sourceBranch,
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
