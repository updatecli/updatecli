package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	transportHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Github contains settings to interact with Github
type Github struct {
	Owner        string
	Repository   string
	Username     string
	Token        string
	URL          string
	Version      string
	Name         string
	directory    string
	Branch       string
	remoteBranch string
	User         string
	Email        string
}

// GetDirectory returns the local git repository path.
func (g *Github) GetDirectory() (directory string) {
	return g.directory
}

// Source retrieves a specific version tag from Github Releases.
func (g *Github) Source() (string, error) {

	_, err := g.Check()
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/%s",
		g.Owner,
		g.Repository,
		g.Version)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "token "+g.Token)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	v := map[string]string{}
	json.Unmarshal(body, &v)

	if val, ok := v["tag_name"]; ok {
		fmt.Printf("\u2714 '%s' github release version founded: %s\n", g.Version, val)
		return val, nil
	}
	fmt.Printf("\u2717 No '%s' github release version founded from %s\n", g.Version, url)
	return "", nil

}

// Check verifies if mandatory Github parameters are provided and return false if not.
func (g *Github) Check() (bool, error) {
	ok := true
	required := []string{}

	if g.Token == "" {
		required = append(required, "token")
	}

	if g.Username == "" {
		required = append(required, "username")
	}

	if g.Owner == "" {
		required = append(required, "owner")
	}

	if g.Repository == "" {
		required = append(required, "repository")
	}

	if len(required) > 0 {
		err := fmt.Errorf("\u2717 Github parameter(s) required: [%v]", strings.Join(required, ","))
		return false, err
	}

	return ok, nil

}

// Init set default Github parameters if not set.
func (g *Github) Init(source string, name string) error {
	g.Version = source
	g.Name = name
	g.setDirectory(source)
	g.remoteBranch = strings.ReplaceAll(fmt.Sprintf("updatecli/%v/%v", g.Name, g.Version), " ", "_")

	if ok, err := g.Check(); !ok {
		return err
	}
	return nil
}

func (g *Github) setDirectory(version string) {

	directory := fmt.Sprintf("%v/%v/%v/%v", os.TempDir(), g.Owner, g.Repository, g.Version)

	if _, err := os.Stat(directory); os.IsNotExist(err) {

		err := os.MkdirAll(directory, 0755)
		if err != nil {
			fmt.Println(err)
		}
	}

	g.directory = directory
}

// Clean deletes github working directory.
func (g *Github) Clean() {
	os.RemoveAll(g.directory)
}

// Clone run `git clone`.
func (g *Github) Clone() string {
	URL := fmt.Sprintf("https://github.com/%v/%v.git",
		g.Owner,
		g.Repository)

	fmt.Printf("Cloning git repository: %s\n", URL)
	fmt.Printf("Downloaded in %s\n\n", g.directory)

	_, err := git.PlainClone(g.directory, false, &git.CloneOptions{
		URL: URL,
		Auth: &transportHttp.BasicAuth{
			Username: g.Username, // anything except an empty string
			Password: g.Token,
		},
		RemoteName: g.Branch,
		Progress:   os.Stdout,
	})
	fmt.Printf("\n")

	g.Checkout()

	if err != nil {
		fmt.Println(err)
	}
	return g.directory
}

// Commit run `git commit`.
func (g *Github) Commit(file, message string) {

	fmt.Printf("Commit changes \n\n")

	r, err := git.PlainOpen(g.directory)
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

// Checkout create and then uses a temporary git branch.
func (g *Github) Checkout() {

	r, err := git.PlainOpen(g.directory)
	if err != nil {
		fmt.Println(err)
	}

	w, err := r.Worktree()
	if err != nil {
		fmt.Println(err)
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(g.remoteBranch),
		Create: true,
		Force:  false,
		Keep:   true,
	})

	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("\n")
}

// Add run `git add`.
func (g *Github) Add(file string) {

	fmt.Printf("Adding file: %s\n", file)

	r, err := git.PlainOpen(g.directory)
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

// Push run `git push` then open a pull request on Github if not already created.
func (g *Github) Push() {

	fmt.Printf("Push changes\n\n")

	r, err := git.PlainOpen(g.directory)
	if err != nil {
		fmt.Println(err)
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

	if err != nil {
		fmt.Println(err)
	}

	err = r.Push(&git.PushOptions{
		RemoteName: g.remoteBranch,
		Progress:   os.Stdout,
	})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("\n")

	g.OpenPR()
}

// OpenPR creates a new pull request.
func (g *Github) OpenPR() {
	fmt.Println("Opening Github pull request")
	title := fmt.Sprintf("[updatecli] Update %v version to %v", g.Name, g.Version)

	if g.isPRExist(title) {
		fmt.Printf("Pull Request titled '%v' already exist\n", title)
		return
	}

	bodyPR := "**! Experimental project**\\nThis pull request was automatically created using [olblak/updatecli](https://github.com/olblak/updatecli)\\nBased on a source rule, it checks if yaml value can be update to the latest version\\nPlease carefully review yaml modification as it also reformat it.\\nPlease report any issues with this tool [here](https://github.com/olblak/updatecli/issues/new)"

	URL := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls",
		g.Owner,
		g.Repository)

	jsonData := fmt.Sprintf("{ \"title\": \"%v\", \"body\": \"%v\", \"head\": \"%v\", \"base\": \"%v\"}", title, bodyPR, g.remoteBranch, g.Branch)

	var jsonStr = []byte(jsonData)

	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(jsonStr))

	req.Header.Add("Authorization", "token "+g.Token)
	req.Header.Add("Content-Type", "application/json")

	if err != nil {
		fmt.Println(err)
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println(err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		fmt.Println(err)
	}

	v := make(map[string]interface{})
	err = json.Unmarshal(body, &v)

	if err != nil {
		fmt.Println(err)
	}

	if res.StatusCode != 201 {
		fmt.Printf("Json Request: %v\n", jsonData)
		fmt.Printf("Response: %v\n", v)
	}

}

// isPRExist checks if an open pull request already exist based on a title.
func (g *Github) isPRExist(title string) bool {

	URL := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls",
		g.Owner,
		g.Repository)

	req, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		fmt.Println(err)
	}

	req.Header.Add("Authorization", "token "+g.Token)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println(err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		fmt.Println(err)
	}

	v := []map[string]string{}

	err = json.Unmarshal(body, &v)

	if err != nil {
		fmt.Println(err)
	}

	for _, v := range v {
		if v["title"] == title {
			return true
		}
	}

	return false
}
