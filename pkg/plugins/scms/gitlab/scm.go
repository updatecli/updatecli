package gitlab

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitlab/client"
)

func (g *Gitlab) GetBranches() (sourceBranch, workingBranch, targetBranch string) {
	sourceBranch = g.Spec.Branch
	workingBranch = g.Spec.Branch
	targetBranch = g.Spec.Branch

	if len(g.pipelineID) > 0 {
		workingBranch = g.nativeGitHandler.SanitizeBranchName(fmt.Sprintf("updatecli_%v", g.pipelineID))
	}

	return sourceBranch, workingBranch, targetBranch
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
		g.GetDirectory())

	if err != nil {
		logrus.Errorf("failed cloning GitLab repository %q", g.GetURL())
		return "", err
	}

	sourceBranch, workingBranch, _ := g.GetBranches()

	if len(workingBranch) > 0 && len(g.GetDirectory()) > 0 {
		err = g.nativeGitHandler.Checkout(
			g.Spec.Username,
			g.Spec.Token,
			sourceBranch,
			workingBranch,
			g.GetDirectory(),
			true)
	}

	if err != nil {
		logrus.Errorf("initial GitLab checkout failed for repository %q", g.GetURL())
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

	err = g.nativeGitHandler.Commit(g.Spec.User, g.Spec.Email, commitMessage, g.GetDirectory(), g.Spec.GPG.SigningKey, g.Spec.GPG.Passphrase)
	if err != nil {
		return err
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
		false)
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

	return g.nativeGitHandler.IsLocalBranchPublished(
		sourceBranch,
		workingBranch,
		g.Spec.Username,
		g.Spec.Token,
		g.GetDirectory())
}

// Push run `git push` to the corresponding GitLab remote branch if not already created.
func (g *Gitlab) Push() error {

	err := g.nativeGitHandler.Push(g.Spec.Username, g.Spec.Token, g.GetDirectory(), g.Spec.Force)
	if err != nil {
		return err
	}

	return nil
}

// PushTag push tags
func (g *Gitlab) PushTag(tag string) error {

	err := g.nativeGitHandler.PushTag(tag, g.Spec.Username, g.Spec.Token, g.GetDirectory(), g.Spec.Force)
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
		g.Spec.Force)
	if err != nil {
		return err
	}

	return nil
}

func (g *Gitlab) GetChangedFiles(workingDir string) ([]string, error) {
	return g.nativeGitHandler.GetChangedFiles(workingDir)
}
