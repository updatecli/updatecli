package git

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// Git contains settings to manipulate a git repository.
type Git struct {
	URL       string
	Branch    string
	User      string
	Email     string
	Directory string
	Version   string
}

// GetDirectory returns the working git directory.
func (g *Git) GetDirectory() (directory string) {
	return g.Directory
}

// Init set Git parameters if needed.
func (g *Git) Init(source string) error {
	g.Version = source
	g.setDirectory(source)
	return nil
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

// Clean removes the current git repository from local storage.
func (g *Git) Clean() {
	os.RemoveAll(g.Directory) // clean up
}

// Clone run `git clone`.
func (g *Git) Clone() string {

	_, err := git.PlainClone(g.Directory, false, &git.CloneOptions{
		URL:        g.URL,
		RemoteName: g.Branch,
		Progress:   os.Stdout,
	})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%s downloaded in %s \n", g.URL, g.Directory)
	return g.Directory
}

// Commit run `git commit`.
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

// Add run `git add`.
func (g *Git) Add(file string) {

	fmt.Printf("Adding file: %s\n", file)

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

// TmpBranch creates an ephemeral branch.
func (g *Git) TmpBranch(name string) {

	r, err := git.PlainOpen(g.Directory)
	if err != nil {
		fmt.Println(err)
	}

	headRef, err := r.Head()
	if err != nil {
		fmt.Println(err)
	}

	reference := plumbing.NewBranchReferenceName(fmt.Sprintf("refs/head/%s", name))
	ref := plumbing.NewHashReference(reference, headRef.Hash())

	err = r.Storer.SetReference(ref)

	if err != nil {
		fmt.Println(err)
	}
}

// Push run `git push`.
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
