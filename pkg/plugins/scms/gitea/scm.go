package gitea

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

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

	URL := fmt.Sprintf("%v/%v/%v.git",
		g.Spec.URL,
		g.Spec.Owner,
		g.Spec.Repository)

	g.setDirectory()

	err := g.nativeGitHandler.Clone(g.Spec.User, g.Spec.Token, URL, g.GetDirectory())

	if err != nil {
		logrus.Errorf("failed cloning Gitea repository %q", URL)
		return "", err
	}

	if len(g.HeadBranch) > 0 && len(g.GetDirectory()) > 0 {
		err = g.nativeGitHandler.Checkout(
			g.Spec.Username,
			g.Spec.Token,
			g.Spec.Branch,
			g.HeadBranch,
			g.GetDirectory(),
			true)
	}

	if err != nil {
		logrus.Errorf("initial Gitea checkout failed for repository %q", URL)
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
	return nil
}

// Checkout create and then uses a temporary git branch.
func (g *Gitea) Checkout() error {
	err := g.nativeGitHandler.Checkout(
		g.Spec.Username,
		g.Spec.Token,
		g.Spec.Branch,
		g.HeadBranch,
		g.Spec.Directory,
		false)
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
	return g.nativeGitHandler.IsLocalBranchPublished(
		g.Spec.Branch,
		g.HeadBranch,
		g.Spec.Username,
		g.Spec.Token,
		g.GetDirectory())
}

// Push run `git push` to the corresponding Gitea remote branch if not already created.
func (g *Gitea) Push() error {

	err := g.nativeGitHandler.Push(g.Spec.Username, g.Spec.Token, g.GetDirectory(), g.Spec.Force)
	if err != nil {
		return err
	}

	return nil
}

// PushTag push tags
func (g *Gitea) PushTag(tag string) error {

	err := g.nativeGitHandler.PushTag(tag, g.Spec.Username, g.Spec.Token, g.GetDirectory(), g.Spec.Force)
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
		g.Spec.Force)
	if err != nil {
		return err
	}

	return nil
}

func (g *Gitea) GetChangedFiles(workingDir string) ([]string, error) {
	return g.nativeGitHandler.GetChangedFiles(workingDir)
}
