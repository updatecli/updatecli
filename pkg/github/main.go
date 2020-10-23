package github

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	git "github.com/olblak/updateCli/pkg/git/generic"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// Changelog contains various information used to describe target changes
type Changelog struct {
	Description string
	Report      string
}

// Github contains settings to interact with Github
type Github struct {
	Owner                  string
	Description            string
	PullRequestDescription Changelog `yaml:"-"`
	Repository             string
	Username               string
	Token                  string
	URL                    string
	Version                string
	Name                   string
	directory              string
	Branch                 string
	remoteBranch           string
	User                   string
	Email                  string
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

	/*
			https://developer.github.com/v4/explorer/
		# Query
		query getLatestRelease($owner: String!, $repository: String!){
			repository(owner: $owner, name: $repository){
				releases(first:10, orderBy:$orderBy){
					nodes{
						name
						tagName
						isDraft
						isPrerelease
					}
				}
			}
		}
		# Variables
		{
			"owner": "olblak",
			"repository": "charts",
		}
	*/

	client := g.NewClient()

	type release struct {
		Name         string
		TagName      string
		IsDraft      bool
		IsPrerelease bool
	}

	var query struct {
		Repository struct {
			Releases struct {
				Nodes []release
			} `graphql:"releases(first: 5, orderBy: $orderBy)"`
		} `graphql:"repository(owner: $owner, name: $repository)"`
	}

	variables := map[string]interface{}{
		"owner":      githubv4.String(g.Owner),
		"repository": githubv4.String(g.Repository),
		"orderBy": githubv4.ReleaseOrder{
			Field:     "CREATED_AT",
			Direction: "DESC",
		},
	}

	err = client.Query(context.Background(), &query, variables)

	if err != nil {
		fmt.Printf("\u2717 Couldn't find a valid github release version\n")
		fmt.Printf("\t %s\n", err)
		return "", err
	}

	value := ""
	for _, release := range query.Repository.Releases.Nodes {
		if !release.IsDraft && !release.IsPrerelease {
			value = release.TagName
			break
		}
	}

	fmt.Printf("\u2714 '%s' github release version founded: %s\n", g.Version, value)
	return value, nil
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
	g.remoteBranch = git.SanitizeBranchName(fmt.Sprintf("updatecli/%v/%v", g.Name, g.Version))

	if ok, err := g.Check(); !ok {
		return err
	}
	return nil
}

func (g *Github) setDirectory(version string) {

	directory := os.path.Join(os.TempDir(), "updatecli", g.Owner, g.Repository)

	if _, err := os.Stat(directory); os.IsNotExist(err) {

		err := os.MkdirAll(directory, 0755)
		if err != nil {
			fmt.Println(err)
		}
	}

	g.directory = directory
}

// Changelog returns a changelog description based on a release name
func (g *Github) Changelog(name string) (string, error) {
	_, err := g.Check()
	if err != nil {
		return "", err
	}

	/*
			https://developer.github.com/v4/explorer/
		# Query
		query getLatestRelease($owner: String!, $repository: String!){
			repository(owner: $owner, name: $repository){
				release(tagName: 1){
					description
					publishedAt
					url
				}
			}
		}
		# Variables
		{
			"owner": "olblak",
			"repository": "charts",
		}
	*/

	client := g.NewClient()

	type release struct {
		Name    string
		TagName string
		Url     string
		Body    string
	}

	var query struct {
		Repository struct {
			Release struct {
				Description string
				Url         string
				PublishedAt time.Time
			} `graphql:"release(tagName: $tagName)"`
		} `graphql:"repository(owner: $owner, name: $repository)"`
	}

	variables := map[string]interface{}{
		"owner":      githubv4.String(g.Owner),
		"repository": githubv4.String(g.Repository),
		"tagName":    githubv4.String(name),
	}

	err = client.Query(context.Background(), &query, variables)

	if err != nil {
		fmt.Printf("\t %s\n", err)
		return "", err
	}

	changelog := fmt.Sprintf("\nRelease published on the %v at the url %v\n%v\n",
		query.Repository.Release.PublishedAt.String(),
		query.Repository.Release.Url,
		query.Repository.Release.Description)

	return changelog, nil
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

	err := git.Clone(g.Username, g.Token, URL, g.GetDirectory())

	if err != nil {
		fmt.Println(err)
	}

	err = git.Checkout(g.Branch, g.remoteBranch, g.GetDirectory())

	if err != nil {
		fmt.Println(err)
	}

	return g.directory
}

// Commit run `git commit`.
func (g *Github) Commit(file, message string) {
	err := git.Commit(file, g.User, g.Email, message, g.GetDirectory())
	if err != nil {
		fmt.Println(err)
	}
}

// Checkout create and then uses a temporary git branch.
func (g *Github) Checkout() {
	err := git.Checkout(g.Branch, g.remoteBranch, g.directory)
	if err != nil {
		fmt.Println(err)
	}
}

// Add run `git add`.
func (g *Github) Add(file string) {

	err := git.Add([]string{file}, g.directory)
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

// OpenPullRequest creates a new pull request.
func (g *Github) OpenPullRequest() {

	/*
		mutation($input: CreatePullRequestInput!){
			createPullRequest(input:$input){
				pullRequest{
					url
				}
			}
		}
		{

		"input":{
			"baseRefName": "x" ,
			"repositoryId":"y",
			"headRefName": "z",
			"title",
			"body",
		}


	*/
	client := g.NewClient()

	var mutation struct {
		createPullRequest struct {
			pullRequest PullRequest
		} `graphql:"createPullRequest(input: $input)"`
	}

	fmt.Println("Opening Github pull request")

	title := fmt.Sprintf("[updatecli] Update %v version to %v", g.Name, g.Version)
	repositoryID, err := g.queryRepositoryID()
	maintainerCanModify := true
	draft := false

	bodyPR, err := SetBody(g.PullRequestDescription)

	if err != nil {
		fmt.Println(err)
		return
	}

	if ok, url, err := g.isPRExist(); ok && err == nil {
		fmt.Printf("Pull Request titled '%v' already exist at\n\t%s\n", title, url)
		return
	}

	if err != nil {
		fmt.Println(err)
	}

	input := githubv4.CreatePullRequestInput{
		BaseRefName:         githubv4.String(g.Branch),
		RepositoryID:        githubv4.String(repositoryID),
		HeadRefName:         githubv4.String(g.remoteBranch),
		Title:               githubv4.String(title),
		Body:                githubv4.NewString(githubv4.String(bodyPR)),
		MaintainerCanModify: githubv4.NewBoolean(githubv4.Boolean(maintainerCanModify)),
		Draft:               githubv4.NewBoolean(githubv4.Boolean(draft)),
	}

	err = client.Mutate(context.Background(), &mutation, input, nil)

	return

}

//// isPRExist checks if an open pull request already exist and is in an open state.
func (g *Github) isPRExist() (bool, string, error) {
	/*
			https://developer.github.com/v4/explorer/
		# Query
		query getPullRequests(
			$owner: String!,
			$name:String!,
			$baseRefName:String!,
			$headRefName:String!){
				repository(owner: $owner, name: $name) {
					pullRequests(baseRefName: $baseRefName, headRefName: $headRefName, last: 1) {
						nodes {
							state
							id
						}
					}
				}
			}
		}
		# Variables
		{
			"owner": "olblak",
			"name": "charts",
			"baseRefName": "master",
			"headRefName": "updatecli/HelmChart/2.4.0"
		}
	*/

	client := g.NewClient()

	var query struct {
		Repository struct {
			PullRequests struct {
				Nodes []PullRequest
			} `graphql:"pullRequests(baseRefName: $baseRefName, headRefName: $headRefName, last: 1, states: OPEN)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner":       githubv4.String(g.Owner),
		"name":        githubv4.String(g.Repository),
		"baseRefName": githubv4.String(g.Branch),
		"headRefName": githubv4.String(g.remoteBranch),
	}

	err := client.Query(context.Background(), &query, variables)

	if err != nil {
		fmt.Println(err)
		return false, "", err
	}

	if len(query.Repository.PullRequests.Nodes) > 0 {
		return true, query.Repository.PullRequests.Nodes[0].Url, nil
	}

	return false, "", nil
}

//NewClient return a new client
func (g *Github) NewClient() *githubv4.Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: g.Token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	if g.URL == "" || strings.HasSuffix(g.URL, "github.com") {
		return githubv4.NewClient(httpClient)

	}

	return githubv4.NewEnterpriseClient(os.Getenv(g.Token), httpClient)
}

func (g *Github) queryRepositoryID() (string, error) {
	/*
		query($owner: String!, $name: String!) {
			repository(owner: $owner, name: $name){
				id
			}
		}
	*/

	client := g.NewClient()

	var query struct {
		Repository struct {
			ID string
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner": githubv4.String(g.Owner),
		"name":  githubv4.String(g.Repository),
	}

	err := client.Query(context.Background(), &query, variables)

	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return query.Repository.ID, nil

}
