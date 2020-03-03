package git

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// Git stores git configuration for the repository that need to be updated
type Git struct {
	URL       string
	Branch    string
	User      string
	Email     string
	Directory string
	Version   string
}

// GetDirectory returns the git working directory
func (g *Git) GetDirectory() (directory string) {
	return g.Directory
}

// Init Directory
func (g *Git) Init(version string) {
	g.Version = version
	g.setDirectory(version)
}

func (g *Git) setDirectory(version string) {

	directory := fmt.Sprintf("%v/%v", os.TempDir(), g.URL)

	if _, err := os.Stat(directory); os.IsNotExist(err) {

		err := os.MkdirAll(directory, 0755)
		if err != nil {
			fmt.Println(err)
		}
	}

	g.Directory = directory

	fmt.Printf("Directory: %v\n", g.Directory)
}

// Clean remove unneeded git repository
func (g *Git) Clean() {
	os.RemoveAll(g.Directory) // clean up
}

// Clone run git clone on a repository containing the yaml that need to be updated
func (g *Git) Clone() string {

	// Disable for now

	_, err := git.PlainClone(g.Directory, false, &git.CloneOptions{
		URL:        g.URL,
		RemoteName: g.Branch,
		Progress:   os.Stdout,
	})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("\t\t%s downloaded in %s", g.URL, g.Directory)
	return g.Directory
}

// Commit run git commit
func (g *Git) Commit(file, message string) {
	// Opens an already existing repository.
	r, err := git.PlainOpen(g.Directory)
	if err != nil {
		fmt.Println(err)
	}

	w, err := r.Worktree()
	if err != nil {
		fmt.Println(err)
	}

	status, err := w.Status()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(status)

	commit, err := w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  g.User,
			Email: g.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		fmt.Println(err)
	}
	obj, err := r.CommitObject(commit)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(obj)

}

// Add execute `git add`
func (g *Git) Add(file string) {

	fmt.Printf("\t\tAdding file: %s", file)

	r, err := git.PlainOpen(g.Directory)
	if err != nil {
		fmt.Println(err)
	}

	w, err := r.Worktree()
	if err != nil {
		fmt.Println(err)
	}
	_, err = w.Add(file)
	if err != nil {
		fmt.Println(err)
	}
}

// TmpBranch will create a new ephemeral branch
func (g *Git) TmpBranch(name string) {

	r, err := git.PlainOpen(g.Directory)
	if err != nil {
		fmt.Println(err)
	}

	// Create a new branch to the current HEAD

	headRef, err := r.Head()
	if err != nil {
		fmt.Println(err)
	}

	// Create a new plumbing.HashReference object with the name of the branch
	// and the hash from the HEAD. The reference name should be a full reference
	// name and not an abbreviated one, as is used on the git cli.
	//
	// For tags we should use `refs/tags/%s` instead of `refs/heads/%s` used
	// for branches.

	reference := plumbing.NewBranchReferenceName(fmt.Sprintf("refs/head/%s", name))
	ref := plumbing.NewHashReference(reference, headRef.Hash())

	// The created reference is saved in the storage.
	// https://github.com/src-d/go-git/blob/master/_examples/checkout/main.go
	err = r.Storer.SetReference(ref)

	if err != nil {
		fmt.Println(err)
	}
}

// Push execute git push
func (g *Git) Push() {

	r, err := git.PlainOpen(g.Directory)
	if err != nil {
		fmt.Println(err)
	}

	err = r.Push(&git.PushOptions{
		RemoteName: g.Branch,
	})
	if err != nil {
		fmt.Println(err)
	}

}
