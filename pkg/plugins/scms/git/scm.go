package git

import (
	"os"

	"github.com/sirupsen/logrus"

	git "github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// Add run `git add`.
func (g *Git) Add(files []string) error {
	err := git.Add(files, g.GetDirectory())
	if err != nil {
		return err
	}
	return nil
}

// Checkout create and then uses a temporary git branch.
func (g *Git) Checkout() error {
	err := git.Checkout(g.Username, g.Password, g.Branch, g.remoteBranch, g.GetDirectory())
	if err != nil {
		return err
	}
	return nil
}

// GetDirectory returns the working git directory.
func (g *Git) GetDirectory() (directory string) {
	return g.Directory
}

// Clean removes the current git repository from local storage.
func (g *Git) Clean() error {
	err := os.RemoveAll(g.Directory) // clean up
	if err != nil {
		return err
	}
	return nil
}

// Clone run `git clone`.
func (g *Git) Clone() (string, error) {

	err := g.Init("")

	if err != nil {
		logrus.Errorf("err - %s", err)
		return "", err
	}

	err = git.Clone(
		g.Username,
		g.Password,
		g.URL,
		g.GetDirectory())

	if err != nil {
		logrus.Errorf("err - %s", err)
		return "", err
	}

	if len(g.remoteBranch) > 0 && len(g.GetDirectory()) > 0 {
		err = g.Checkout()
		if err != nil {
			logrus.Errorf("err - %s", err)
			return "", err
		}
	}

	return g.Directory, nil
}

// Commit run `git commit`.
func (g *Git) Commit(message string) error {

	// Generate the conventional commit message
	commitMessage, err := g.CommitMessage.Generate(message)
	if err != nil {
		return err
	}

	err = git.Commit(
		g.User,
		g.Email,
		commitMessage,
		g.GetDirectory(),
		g.GPG.SigningKey,
		g.GPG.Passphrase,
	)
	if err != nil {
		return err
	}

	return nil
}

// Init set Git parameters if needed.
func (g *Git) Init(pipelineID string) (err error) {
	if len(g.Directory) == 0 {
		g.Directory, err = newDirectory(g.URL)
		if err != nil {
			return err
		}
	}

	g.remoteBranch = git.SanitizeBranchName(g.Branch)

	if len(g.Branch) == 0 {
		g.Branch = "main"
	}

	return nil
}

// Push run `git push`.
func (g *Git) Push() error {
	err := git.Push(
		g.Username,
		g.Password,
		g.GetDirectory(),
		g.Force)

	if err != nil {
		return err
	}

	return nil

}

// PushTag push tags
func (g *Git) PushTag(tag string) error {

	err := git.PushTag(tag, g.Username, g.Password, g.GetDirectory(), g.Force)
	if err != nil {
		return err
	}

	return nil
}

func (g *Git) GetChangedFiles(workingDir string) ([]string, error) {
	return git.GetChangedFiles(workingDir)
}
