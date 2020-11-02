package git

import (
	"fmt"
	"os"

	git "github.com/olblak/updateCli/pkg/plugins/git/generic"
)

// Add run `git add`.
func (g *Git) Add(files []string) {

	err := git.Add(files, g.GetDirectory())

	if err != nil {
		fmt.Println(err)
	}
}

// Checkout create and then uses a temporary git branch.
func (g *Git) Checkout() {
	err := git.Checkout(g.Branch, g.remoteBranch, g.GetDirectory())
	if err != nil {
		fmt.Println(err)
	}
}

// GetDirectory returns the working git directory.
func (g *Git) GetDirectory() (directory string) {
	return g.Directory
}

// Clean removes the current git repository from local storage.
func (g *Git) Clean() {
	os.RemoveAll(g.Directory) // clean up
}

// Clone run `git clone`.
func (g *Git) Clone() string {

	err := git.Clone(
		g.Username,
		g.Password,
		g.URL,
		g.GetDirectory())

	g.setDirectory()

	if err != nil {
		fmt.Println(err)
	}

	g.Checkout()

	return g.Directory
}

// Commit run `git commit`.
func (g *Git) Commit(message string) {

	err := git.Commit(
		g.User,
		g.Email,
		message,
		g.GetDirectory())

	if err != nil {
		fmt.Println(err)
	}

}

// Init set Git parameters if needed.
func (g *Git) Init(source string, name string) error {
	g.Version = source
	g.setDirectory()
	g.remoteBranch = git.SanitizeBranchName(g.Branch)
	return nil
}

// Push run `git push`.
func (g *Git) Push() {

	err := git.Push(
		g.Username,
		g.Password,
		g.GetDirectory())

	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("\n")

}
