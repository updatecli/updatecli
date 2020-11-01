package github

import (
	"fmt"
	"os"

	git "github.com/olblak/updateCli/pkg/plugins/git/generic"
)

// Init set default Github parameters if not set.
func (g *Github) Init(source string, name string) error {
	g.Version = source
	g.Name = name
	g.remoteBranch = git.SanitizeBranchName(fmt.Sprintf("updatecli/%v/%v", g.Name, g.Version))
	g.setDirectory()

	if ok, err := g.Check(); !ok {
		return err
	}
	return nil
}

// GetDirectory returns the local git repository path.
func (g *Github) GetDirectory() (directory string) {
	return g.Directory
}

// Clean deletes github working directory.
func (g *Github) Clean() {
	os.RemoveAll(g.Directory)
}

// Clone run `git clone`.
func (g *Github) Clone() string {

	URL := fmt.Sprintf("https://github.com/%v/%v.git",
		g.Owner,
		g.Repository)

	g.setDirectory()

	err := git.Clone(g.Username, g.Token, URL, g.GetDirectory())

	if err != nil {
		fmt.Println(err)
	}

	err = git.Checkout(g.Branch, g.remoteBranch, g.GetDirectory())

	if err != nil {
		fmt.Println(err)
	}

	return g.Directory
}

// Commit run `git commit`.
func (g *Github) Commit(message string) {
	err := git.Commit(g.User, g.Email, message, g.GetDirectory())
	if err != nil {
		fmt.Println(err)
	}
}

// Checkout create and then uses a temporary git branch.
func (g *Github) Checkout() {
	err := git.Checkout(g.Branch, g.remoteBranch, g.Directory)
	if err != nil {
		fmt.Println(err)
	}
}

// Add run `git add`.
func (g *Github) Add(files []string) {

	err := git.Add(files, g.Directory)
	if err != nil {
		fmt.Println(err)
	}
}

// Push run `git push` then open a pull request on Github if not already created.
func (g *Github) Push() {

	err := git.Push(g.Username, g.Token, g.GetDirectory())
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("\n")

	g.OpenPullRequest()
}
