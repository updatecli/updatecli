package azure

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/azure/devops/client"
)

// GetDirectory returns the local git repository path.
func (g *Azure) GetDirectory() (directory string) {
	return g.Spec.Directory
}

// Clean deletes github working directory.
func (g *Azure) Clean() error {
	err := os.RemoveAll(g.Spec.Directory)
	if err != nil {
		return err
	}
	return nil
}

// Clone run `git clone`.
func (g *Azure) Clone() (string, error) {

	// Azure DevOps requires capabilities multi_ack / multi_ack_detailed,
	// which are not fully implemented and by default are included in
	// transport.UnsupportedCapabilities.
	//
	// The initial clone operations require a full download of the repository,
	// and therefore those unsupported capabilities are not as crucial, so
	// by removing them from that list allows for the first clone to work
	// successfully.
	//
	// Additional fetches will yield issues, therefore work always from a clean
	// clone until those capabilities are fully supported.
	//
	// New commits and pushes against a remote worked without any issues.
	transport.UnsupportedCapabilities = []capability.Capability{
		capability.ThinPack,
	}

	url := g.Spec.URL

	if url == "" {
		url = client.AZUREDOMAIN
	}

	URL := fmt.Sprintf("https://%s/%s/%s/_git/%s",
		url,
		g.Spec.Owner,
		g.Spec.Project,
		g.Spec.RepoID)

	g.setDirectory()

	err := g.nativeGitHandler.Clone(g.Spec.User, g.Spec.Token, URL, g.GetDirectory())

	if err != nil {
		logrus.Errorf("failed cloning Azure DevOps repository %q", URL)
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
		logrus.Errorf("initial Azure DevOps checkout failed for repository %q", URL)
		return "", err
	}

	return g.Spec.Directory, nil
}

// Commit run `git commit`.
func (g *Azure) Commit(message string) error {

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
		g.Spec.GPG.Passphrase)

	if err != nil {
		return err
	}
	return nil
}

// Checkout create and then uses a temporary git branch.
func (g *Azure) Checkout() error {
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
func (g *Azure) Add(files []string) error {

	err := g.nativeGitHandler.Add(files, g.Spec.Directory)
	if err != nil {
		return err
	}
	return nil
}

// IsRemoteBranchUpToDate checks if the branche reference name is published on
// on the default remote
func (g *Azure) IsRemoteBranchUpToDate() (bool, error) {
	return g.nativeGitHandler.IsLocalBranchPublished(
		g.Spec.Branch,
		g.HeadBranch,
		g.Spec.Username,
		g.Spec.Token,
		g.GetDirectory())
}

// Push run `git push` to the corresponding Azure DevOps remote branch if not already created.
func (g *Azure) Push() error {

	err := g.nativeGitHandler.Push(g.Spec.Username, g.Spec.Token, g.GetDirectory(), g.Spec.Force)
	if err != nil {
		return err
	}

	return nil
}

// PushTag push tags
func (g *Azure) PushTag(tag string) error {

	err := g.nativeGitHandler.PushTag(tag, g.Spec.Username, g.Spec.Token, g.GetDirectory(), g.Spec.Force)
	if err != nil {
		return err
	}

	return nil
}

// PushBranch push branch
func (g *Azure) PushBranch(branch string) error {

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

func (g *Azure) GetChangedFiles(workingDir string) ([]string, error) {
	return g.nativeGitHandler.GetChangedFiles(workingDir)
}
