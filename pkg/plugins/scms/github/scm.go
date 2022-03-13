package github

import (
	"fmt"
	"os"
)

// GetDirectory returns the local git repository path.
func (g *Github) GetDirectory() (directory string) {
	return g.Spec.Directory
}

// Clean deletes github working directory.
func (g *Github) Clean() error {
	err := os.RemoveAll(g.Spec.Directory)
	if err != nil {
		return err
	}
	return nil
}

// Clone run `git clone`.
func (g *Github) Clone() (string, error) {

	URL := fmt.Sprintf("https://github.com/%v/%v.git",
		g.Spec.Owner,
		g.Spec.Repository)

	g.setDirectory()

	err := g.NativeGitHandler.Clone(g.Spec.Username, g.Spec.Token, URL, g.GetDirectory())

	if err != nil {
		return "", err
	}

	if len(g.HeadBranch) > 0 && len(g.GetDirectory()) > 0 {
		err = g.NativeGitHandler.Checkout(g.Spec.Username, g.Spec.Token, g.Spec.Branch, g.HeadBranch, g.GetDirectory())
	}

	if err != nil {
		return "", err
	}

	return g.Spec.Directory, nil
}

// Commit run `git commit`.
func (g *Github) Commit(message string) error {

	// Generate the conventional commit message
	commitMessage, err := g.Spec.CommitMessage.Generate(message)
	if err != nil {
		return err
	}

	err = g.NativeGitHandler.Commit(g.Spec.User, g.Spec.Email, commitMessage, g.GetDirectory(), g.Spec.GPG.SigningKey, g.Spec.GPG.Passphrase)
	if err != nil {
		return err
	}
	return nil
}

// Checkout create and then uses a temporary git branch.
func (g *Github) Checkout() error {
	err := g.NativeGitHandler.Checkout(g.Spec.Username, g.Spec.Token, g.Spec.Branch, g.HeadBranch, g.Spec.Directory)
	if err != nil {
		return err
	}
	return nil
}

// Add run `git add`.
func (g *Github) Add(files []string) error {

	err := g.NativeGitHandler.Add(files, g.Spec.Directory)
	if err != nil {
		return err
	}
	return nil
}

// Push run `git push` then open a pull request on Github if not already created.
func (g *Github) Push() error {

	err := g.NativeGitHandler.Push(g.Spec.Username, g.Spec.Token, g.GetDirectory(), g.Spec.Force)
	if err != nil {
		return err
	}

	return nil
}

// PushTag push tags
func (g *Github) PushTag(tag string) error {

	err := g.NativeGitHandler.PushTag(tag, g.Spec.Username, g.Spec.Token, g.GetDirectory(), g.Spec.Force)
	if err != nil {
		return err
	}

	return nil
}
