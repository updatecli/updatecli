package git

import (
	"os"

	"github.com/sirupsen/logrus"
)

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
	err := g.nativeGitHandler.Checkout(
		g.spec.Username,
		g.spec.Password,
		g.spec.Branch,
		g.remoteBranch,
		g.GetDirectory())
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
		g.spec.URL,
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

	return g.spec.Directory, nil
}

// Commit run `git commit`.
func (g *Git) Commit(message string) error {

	// Generate the conventional commit message
	commitMessage, err := g.spec.CommitMessage.Generate(message)
	if err != nil {
		return err
	}

	err = g.nativeGitHandler.Commit(
		g.spec.User,
		g.spec.Email,
		commitMessage,
		g.GetDirectory(),
		g.spec.GPG.SigningKey,
		g.spec.GPG.Passphrase,
	)
	if err != nil {
		return err
	}

	return nil
}

// Push run `git push`.
func (g *Git) Push() error {
	err := g.nativeGitHandler.Push(
		g.spec.Username,
		g.spec.Password,
		g.GetDirectory(),
		g.spec.Force)

	if err != nil {
		return err
	}

	return nil

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
