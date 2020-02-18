package github

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// Github describe github settings
type Github struct {
	Owner      string
	Repository string
	Username   string
	Token      string
	URL        string
	Version    string
	Directory  string
	Branch     string
	User       string
	Email      string
}

// GetDirectory returns the git working directory
func (g *Github) GetDirectory() (directory string) {
	return g.Directory
}

// GetVersion retrieves the version tag from releases
func (g *Github) GetVersion() string {

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/%s",
		g.Owner,
		g.Repository,
		g.Version)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.Println(err)
	}

	req.Header.Add("Authorization", "token "+g.Token)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Println(err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Println(err)
	}

	v := map[string]string{}
	json.Unmarshal(body, &v)

	if val, ok := v["name"]; ok {
		return val
	}
	log.Printf("\u2717 No tag founded from %s\n", url)
	return ""

}

// Init Directory before running git clone
func (g *Github) Init() {
	if g.Directory == "" {
		name, err := ioutil.TempDir("", "updateCli")
		g.Directory = name
		if err != nil {
			log.Fatal(err)
		}
	}
}

// Clean remote working directory
func (g *Github) Clean() {
	os.RemoveAll(g.Directory)
}

// Clone run git clone
func (g *Github) Clone() string {
	URL := fmt.Sprintf("https://%v:%v@github.com/%v/%v.git",
		g.Username,
		g.Token,
		g.Owner,
		g.Repository)
	_, err := git.PlainClone(g.Directory, false, &git.CloneOptions{
		URL:        URL,
		RemoteName: g.Branch,
		Progress:   os.Stdout,
	})

	if err != nil {
		log.Println(err)
	}
	log.Printf("\t\t%s downloaded in %s", URL, g.Directory)
	return g.Directory
}

// Commit run git commit
func (g *Github) Commit(file, message string) {
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
func (g *Github) Add(file string) {

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
func (g *Github) TmpBranch(name string) {

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
func (g *Github) Push() {

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
