package gitea

import (
	"fmt"
	"os"

	git "github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
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

	err := git.Clone(g.Spec.User, g.Spec.Token, URL, g.GetDirectory())

	if err != nil {
		return "", err
	}

	if len(g.HeadBranch) > 0 && len(g.GetDirectory()) > 0 {
		err = git.Checkout(g.Spec.Username, g.Spec.Token, g.Spec.Branch, g.HeadBranch, g.GetDirectory())
	}

	if err != nil {
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

	err = git.Commit(g.Spec.User, g.Spec.Email, commitMessage, g.GetDirectory(), g.Spec.GPG.SigningKey, g.Spec.GPG.Passphrase)
	if err != nil {
		return err
	}
	return nil
}

// Checkout create and then uses a temporary git branch.
func (g *Gitea) Checkout() error {
	err := git.Checkout(g.Spec.Username, g.Spec.Token, g.Spec.Branch, g.HeadBranch, g.Spec.Directory)
	if err != nil {
		return err
	}
	return nil
}

// Add run `git add`.
func (g *Gitea) Add(files []string) error {

	err := git.Add(files, g.Spec.Directory)
	if err != nil {
		return err
	}
	return nil
}

// Push run `git push` then open a pull request on Github if not already created.
func (g *Gitea) Push() error {

	err := git.Push(g.Spec.Username, g.Spec.Token, g.GetDirectory(), g.Spec.Force)
	if err != nil {
		return err
	}

	return nil
}

// PushTag push tags
func (g *Gitea) PushTag(tag string) error {

	err := git.PushTag(tag, g.Spec.Username, g.Spec.Token, g.GetDirectory(), g.Spec.Force)
	if err != nil {
		return err
	}

	return nil
}

func (g *Gitea) GetChangedFiles(workingDir string) ([]string, error) {
	return git.GetChangedFiles(workingDir)
}
