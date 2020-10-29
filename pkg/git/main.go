package git

import (
	"fmt"
	"os"
	"path"

	git "github.com/olblak/updateCli/pkg/git/generic"
	"github.com/olblak/updateCli/pkg/tmp"
)

// Git contains settings to manipulate a git repository.
type Git struct {
	URL          string
	Username     string
	Password     string
	Branch       string
	remoteBranch string
	User         string
	Email        string
	Directory    string
	Version      string
}

// GetDirectory returns the working git directory.
func (g *Git) GetDirectory() (directory string) {
	return g.Directory
}

// Init set Git parameters if needed.
func (g *Git) Init(source string, name string) error {
	g.Version = source
	g.setDirectory(source)
	g.remoteBranch = git.SanitizeBranchName(g.Branch)
	return nil
}

func (g *Git) setDirectory(version string) {

	directory := path.Join(tmp.Directory, g.URL)

	if _, err := os.Stat(directory); os.IsNotExist(err) {

		err := os.MkdirAll(directory, 0755)
		if err != nil {
			fmt.Println(err)
		}
	}

	g.Directory = directory

	fmt.Printf("Directory: %v\n", g.Directory)
}

// Checkout create and then uses a temporary git branch.
func (g *Git) Checkout() {
	err := git.Checkout(g.Branch, g.remoteBranch, g.GetDirectory())
	if err != nil {
		fmt.Println(err)
	}
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

	if err != nil {
		fmt.Println(err)
	}

	g.Checkout()

	return g.Directory
}

// Commit run `git commit`.
func (g *Git) Commit(file, message string) {

	err := git.Commit(file,
		g.User,
		g.Email,
		message,
		g.GetDirectory())

	if err != nil {
		fmt.Println(err)
	}

}

// Add run `git add`.
func (g *Git) Add(file string) {

	err := git.Add([]string{file}, g.GetDirectory())

	if err != nil {
		fmt.Println(err)
	}
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
