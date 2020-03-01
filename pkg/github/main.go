package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// Github holds github specific configuration
type Github struct {
	Owner        string
	Repository   string
	Username     string
	Token        string
	URL          string
	Version      string
	directory    string
	Branch       string
	remoteBranch string
	User         string
	Email        string
}

// GetDirectory returns the git working directory
func (g *Github) GetDirectory() (directory string) {
	return g.directory
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
func (g *Github) Init(version string) {
	g.Version = version
	g.setDirectory(version)
	g.remoteBranch = fmt.Sprintf("updatecli/%v", version)

}

func (g *Github) setDirectory(version string) {

	directory := fmt.Sprintf("%v/%v/%v/%v", os.TempDir(), g.Owner, g.Repository, g.Version)

	if _, err := os.Stat(directory); os.IsNotExist(err) {

		err := os.MkdirAll(directory, 0755)
		if err != nil {
			log.Println(err)
		}
	}

	g.directory = directory

	fmt.Printf("Directory: %v\n", g.directory)
}

// Clean remote working directory
func (g *Github) Clean() {
	os.RemoveAll(g.directory)
}

// Clone run git clone
func (g *Github) Clone() string {
	URL := fmt.Sprintf("https://%v:%v@github.com/%v/%v.git",
		g.Username,
		g.Token,
		g.Owner,
		g.Repository)
	_, err := git.PlainClone(g.directory, false, &git.CloneOptions{
		URL:        URL,
		RemoteName: g.Branch,
		Progress:   os.Stdout,
	})

	g.Checkout()

	if err != nil {
		log.Println(err)
	}
	log.Printf("\t\t%s downloaded in %s", URL, g.directory)
	return g.directory
}

// Commit run git commit
func (g *Github) Commit(file, message string) {
	// Opens an existing repository.
	r, err := git.PlainOpen(g.directory)
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

// Checkout will create and use a temporary branch
func (g *Github) Checkout() {
	r, err := git.PlainOpen(g.directory)
	if err != nil {
		log.Println(err)
	}

	w, err := r.Worktree()
	if err != nil {
		log.Println(err)
	}

	branch := "updatecli/" + g.Version
	log.Printf("Creating Branch: %v\n", branch)

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
		Create: true,
		Force:  false,
		Keep:   true,
	})

	if err != nil {
		log.Println(err)
	}
}

// Add execute `git add`
func (g *Github) Add(file string) {

	log.Printf("\t\tAdding file: %s", file)

	r, err := git.PlainOpen(g.directory)
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

// Push execute git push
func (g *Github) Push() {

	r, err := git.PlainOpen(g.directory)
	if err != nil {
		log.Println(err)
	}

	URL := fmt.Sprintf("https://%v:%v@github.com/%v/%v.git",
		g.Username,
		g.Token,
		g.Owner,
		g.Repository)

	_, err = r.CreateRemote(&config.RemoteConfig{
		Name: g.remoteBranch,
		URLs: []string{URL},
	})

	err = r.Push(&git.PushOptions{
		RemoteName: g.remoteBranch,
		Progress:   os.Stdout,
	})
	if err != nil {
		log.Println(err)
	}

	g.OpenPR()
}

// OpenPR creates a new Github Pull Request
func (g *Github) OpenPR() {
	title := fmt.Sprintf("[Updatecli] Bump to version %v", g.Version)

	if g.isPRExist(title) {
		log.Println("PR already exist")
		return
	}

	bodyPR := "Please pull these awesome changes in!"

	URL := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls",
		g.Owner,
		g.Repository)

	jsonData := fmt.Sprintf("{ \"title\": \"%v\", \"body\": \"%v\", \"head\": \"%v\", \"base\": \"%v\"}", title, bodyPR, g.remoteBranch, g.Branch)

	var jsonStr = []byte(jsonData)

	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(jsonStr))

	req.Header.Add("Authorization", "token "+g.Token)
	req.Header.Add("Content-Type", "application/json")

	if err != nil {
		log.Println(err)
	}

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
	err = json.Unmarshal(body, &v)

	if err != nil {
		log.Println(err)
	}

	log.Println(res.Status)

}

// OpenPR creates a new Github Pull Request
func (g *Github) isPRExist(title string) bool {

	URL := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls",
		g.Owner,
		g.Repository)

	req, err := http.NewRequest("GET", URL, nil)

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

	v := []map[string]string{}

	err = json.Unmarshal(body, &v)

	if err != nil {
		log.Println(err)
	}

	for _, v := range v {
		if v["title"] == title {
			return true
		}
	}

	return false
}
