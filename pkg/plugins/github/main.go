package github

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/shurcooL/githubv4"
	"github.com/updatecli/updatecli/pkg/core/tmp"
	"github.com/updatecli/updatecli/pkg/plugins/git/commit"
	"github.com/updatecli/updatecli/pkg/plugins/version"
	"golang.org/x/oauth2"
)

// Spec represents the configuration input
type Spec struct {
	Branch        string         // Branch specifies which github branch to work on
	Directory     string         // Directory specifies where the github repisotory is cloned on the local disk
	Email         string         // Email specifies which emails to use when creating commits
	Owner         string         // Owner specifies repository owner
	Repository    string         // Repository specifies the name of a repository for a specific owner
	Version       string         // **Deprecated** Version is deprecated in favor of `versionFilter.pattern`, this field will be removed in a future version
	VersionFilter version.Filter //VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	Token         string         // Token specifies the credential used to authenticate with
	URL           string         // URL specifies the default github url in case of GitHub enterprise
	Username      string         // Username specifies the username used to authenticate with Github API
	User          string         // User specific the user in git commit messages
	PullRequest   PullRequestSpec
}

// Github contains settings to interact with Github
type Github struct {
	spec          Spec          // Spec contains inputs coming from updatecli configuration
	remoteBranch  string        // remoteBranch is used when creating a temporary branch before opening a PR
	Force         bool          // Force is used during the git push phase to run `git push --force`.
	CommitMessage commit.Commit // CommitMessage represents conventional commit metadata as type or scope, used to generate the final commit message.
	pullRequest   struct {      // pullRequest contain the pull request information
		Title       string // Override default pull request title
		LabelIDs    []string
		Description string
		Report      string
	}
	remotePullRequest PullRequest
}

// New returns a new valid Github object.
func New(s Spec) (Github, error) {
	errs := s.Validate()

	if len(errs) > 0 {
		strErrs := []string{}
		for _, err := range errs {
			strErrs = append(strErrs, err.Error())
		}
		return Github{}, fmt.Errorf(strings.Join(strErrs, "\n"))
	}

	if s.Directory == "" {
		s.Directory = path.Join(tmp.Directory, s.Owner, s.Repository)
	}

	if s.URL == "" {
		s.URL = "github.com"
	}

	return Github{
		spec: s}, nil
}

// Validate verifies if mandatory Github parameters are provided and return false if not.
func (s *Spec) Validate() (errs []error) {
	required := []string{}

	if len(s.Token) == 0 {
		required = append(required, "token")
	}

	if len(s.Owner) == 0 {
		required = append(required, "owner")
	}

	if len(s.Repository) == 0 {
		required = append(required, "repository")
	}

	if len(s.VersionFilter.Pattern) == 0 {
		s.VersionFilter.Pattern = s.Version
	}

	if err := s.VersionFilter.Validate(); err != nil {
		errs = append(errs, err)
	}

	if len(s.Version) > 0 {
		logrus.Warningln("**Deprecated** Field `version` from resource githubRelease is deprecated in favor of `versionFilter.pattern`, this field will be removed in the next major version")
	}

	if len(required) > 0 {
		errs = append(errs, fmt.Errorf("github parameter(s) required: [%v]", strings.Join(required, ",")))
	}

	return errs
}

func (g *Github) setDirectory() {

	if _, err := os.Stat(g.spec.Directory); os.IsNotExist(err) {

		err := os.MkdirAll(g.spec.Directory, 0755)
		if err != nil {
			logrus.Errorf("err - %s", err)
		}
	}
}

//NewClient return a new client
func (g *Github) NewClient() *githubv4.Client {

	if err := g.spec.Validate(); err != nil {
		logrus.Errorln(err)
		return nil
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: g.spec.Token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	if g.spec.URL == "" || strings.HasSuffix(g.spec.URL, "github.com") {
		return githubv4.NewClient(httpClient)
	}

	return githubv4.NewEnterpriseClient(os.Getenv(g.spec.Token), httpClient)
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
		"owner": githubv4.String(g.spec.Owner),
		"name":  githubv4.String(g.spec.Repository),
	}

	err := client.Query(context.Background(), &query, variables)

	if err != nil {
		logrus.Errorf("err - %s", err)
		return "", err
	}

	return query.Repository.ID, nil

}
