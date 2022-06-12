package github

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/shurcooL/githubv4"
	"github.com/updatecli/updatecli/pkg/core/tmp"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/commit"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"

	git "github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// Spec represents the configuration input
type Spec struct {
	// Branch specifies which github branch to work on
	Branch string
	// Directory specifies where the github repository is cloned on the local disk
	Directory string
	// Email specifies which emails to use when creating commits
	Email string
	// Owner specifies repository owner
	Owner string `jsonschema:"required"`
	// Repository specifies the name of a repository for a specific owner
	Repository string `jsonschema:"required"`
	// Token specifies the credential used to authenticate with
	Token string `jsonschema:"required"`
	// URL specifies the default github url in case of GitHub enterprise
	URL string
	// Username specifies the username used to authenticate with Github API
	Username string `jsonschema:"required"`
	// User specifies the user of the git commit messages
	User string
	// Deprecated since https://github.com/updatecli/updatecli/issues/260, must be clean up
	PullRequest PullRequestSpec
	// GPG key and passphrased used for commit signing
	GPG sign.GPGSpec
	// Force is used during the git push phase to run `git push --force`.
	Force bool
	// CommitMessage represents conventional commit metadata as type or scope, used to generate the final commit message.
	CommitMessage commit.Commit
}

// Github contains settings to interact with Github
type Github struct {
	// Spec contains inputs coming from updatecli configuration
	Spec Spec
	// HeadBranch is used when creating a temporary branch before opening a PR
	HeadBranch string
	client     GitHubClient
}

// New returns a new valid Github object.
func New(s Spec, pipelineID string) (*Github, error) {
	errs := s.Validate()

	if len(errs) > 0 {
		strErrs := []string{}
		for _, err := range errs {
			strErrs = append(strErrs, err.Error())
		}
		return &Github{}, fmt.Errorf(strings.Join(strErrs, "\n"))
	}

	if s.Directory == "" {
		s.Directory = path.Join(tmp.Directory, s.Owner, s.Repository)
	}

	if s.URL == "" {
		s.URL = "github.com"
	}

	// Initialize github client
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: s.Token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	g := Github{
		Spec:       s,
		HeadBranch: git.SanitizeBranchName(fmt.Sprintf("updatecli_%v", pipelineID)),
	}

	if strings.HasSuffix(s.URL, "github.com") {
		g.client = githubv4.NewClient(httpClient)
	} else {
		g.client = githubv4.NewEnterpriseClient(os.Getenv(s.URL), httpClient)
	}

	g.setDirectory()

	return &g, nil
}

// Validate verifies if mandatory Github parameters are provided and return false if not.
func (s *Spec) Validate() (errs []error) {
	required := []string{}

	if err := s.PullRequest.Validate(); err != nil {
		errs = append(errs, err)
	}

	if len(s.Token) == 0 {
		required = append(required, "token")
	}

	if len(s.Owner) == 0 {
		required = append(required, "owner")
	}

	if len(s.Repository) == 0 {
		required = append(required, "repository")
	}

	if len(required) > 0 {
		errs = append(errs, fmt.Errorf("github parameter(s) required: [%v]", strings.Join(required, ",")))
	}

	return errs
}

func (g *Github) setDirectory() {

	if _, err := os.Stat(g.Spec.Directory); os.IsNotExist(err) {

		err := os.MkdirAll(g.Spec.Directory, 0755)
		if err != nil {
			logrus.Errorf("err - %s", err)
		}
	}
}

func (g *Github) queryRepositoryID() (string, error) {
	/*
		query($owner: String!, $name: String!) {
			repository(owner: $owner, name: $name){
				id
			}
		}
	*/

	var query struct {
		Repository struct {
			ID string
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner": githubv4.String(g.Spec.Owner),
		"name":  githubv4.String(g.Spec.Repository),
	}

	err := g.client.Query(context.Background(), &query, variables)

	if err != nil {
		logrus.Errorf("err - %s", err)
		return "", err
	}

	return query.Repository.ID, nil

}

// SpecToPullRequestSpec is a function that export the pullRequest spec from
// a GithubSpec to a PullRequest.Spec. It's temporary function until we totally remove
// the old scm configuration introduced by this https://github.com/updatecli/updatecli/pull/388
func (s *Spec) SpecToPullRequestSpec() interface{} {
	return s.PullRequest
}
