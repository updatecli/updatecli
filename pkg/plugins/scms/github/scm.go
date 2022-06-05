package github

import (
	"fmt"
	"os"

	git "github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// Init set default Github parameters if not set.
func (g *Github) Init(pipelineID string) error {
	g.HeadBranch = git.SanitizeBranchName(fmt.Sprintf("updatecli_%v", pipelineID))
	g.setDirectory()

	return nil
}

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

	err := git.Clone(g.Spec.Username, g.Spec.Token, URL, g.GetDirectory())

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
func (g *Github) Commit(message string) error {

	// Generate the conventional commit message
	commitMessage, err := g.CommitMessage.Generate(message)
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
func (g *Github) Checkout() error {
	err := git.Checkout(g.Spec.Username, g.Spec.Token, g.Spec.Branch, g.HeadBranch, g.Spec.Directory)
	if err != nil {
		return err
	}
	return nil
}

// Add run `git add`.
func (g *Github) Add(files []string) error {

	err := git.Add(files, g.Spec.Directory)
	if err != nil {
		return err
	}
	return nil
}

// Push run `git push` then open a pull request on Github if not already created.
func (g *Github) Push() error {

	err := git.Push(g.Spec.Username, g.Spec.Token, g.GetDirectory(), g.Force)
	if err != nil {
		return err
	}

	return nil
}

// PushTag push tags
func (g *Github) PushTag(tag string) error {

	err := git.PushTag(tag, g.Spec.Username, g.Spec.Token, g.GetDirectory(), g.Force)
	if err != nil {
		return err
	}

	return nil
}

func (g *Github) GetChangedFiles(workingDir string) ([]string, error) {
	return git.GetChangedFiles(workingDir)
}

func (g *Github) ToString() string {
	redacted := Spec{
		Branch:      g.Spec.Branch,
		Directory:   g.Spec.Directory,
		Email:       g.Spec.Email,
		Owner:       g.Spec.Owner,
		Repository:  g.Spec.Repository,
		Token:       "******",
		URL:         g.Spec.URL,
		Username:    g.Spec.Username,
		User:        g.Spec.User,
		PullRequest: g.Spec.PullRequest,
		GPG:         g.Spec.GPG,
	}

	return fmt.Sprintf("%+v", redacted)
}
