package git

import (
	"fmt"
	"io/ioutil"
	"log"
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
}

// GetDirectory returns the git working directory
func (g *Git) GetDirectory() (directory string) {
	return g.Directory
}

// Init Directory
func (g *Git) Init() {
	if g.Directory == "" {
		// Create temporary working directory
		name, err := ioutil.TempDir("", "updateCli")
		g.Directory = name
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("\tInit Directory %s", g.Directory)
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
		log.Println(err)
	}

	log.Printf("\t\t%s downloaded in %s", g.URL, g.Directory)
	return g.Directory
}

// Commit run git commit
func (g *Git) Commit(file, message string) {
	// Opens an already existing repository.
	r, err := git.PlainOpen(g.Directory)
	if err != nil {
		log.Println(err)
	}

	w, err := r.Worktree()
	if err != nil {
		log.Println(err)
	}

	status, err := w.Status()
	if err != nil {
		log.Println(err)
	}
	log.Println(status)

	commit, err := w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  g.User,
			Email: g.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		log.Println(err)
	}
	obj, err := r.CommitObject(commit)
	if err != nil {
		log.Println(err)
	}
	log.Println(obj)

}

// Add execute `git add`
func (g *Git) Add(file string) {

	log.Printf("\t\tAdding file: %s", file)

	r, err := git.PlainOpen(g.Directory)
	if err != nil {
		log.Println(err)
	}

	w, err := r.Worktree()
	if err != nil {
		log.Println(err)
	}
	_, err = w.Add(file)
	if err != nil {
		log.Println(err)
	}
}

// TmpBranch will create a new ephemeral branch
func (g *Git) TmpBranch(name string) {

	r, err := git.PlainOpen(g.Directory)
	if err != nil {
		log.Println(err)
	}

	// Create a new branch to the current HEAD

	headRef, err := r.Head()
	if err != nil {
		log.Println(err)
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
		log.Println(err)
	}
}

// Push execute git push
func (g *Git) Push() {

	r, err := git.PlainOpen(g.Directory)
	if err != nil {
		log.Println(err)
	}

	err = r.Push(&git.PushOptions{
		RemoteName: g.Branch,
	})
	if err != nil {
		log.Println(err)
	}

}
