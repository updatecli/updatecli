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
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec represents the configuration input
type Spec struct {
	Branch        string          // Branch specifies which github branch to work on
	Directory     string          // Directory specifies where the github repisotory is cloned on the local disk
	Email         string          // Email specifies which emails to use when creating commits
	Owner         string          // Owner specifies repository owner
	Repository    string          // Repository specifies the name of a repository for a specific owner
	Version       string          // **Deprecated** Version is deprecated in favor of `versionFilter.pattern`, this field will be removed in a future version
	VersionFilter version.Filter  //VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	Token         string          // Token specifies the credential used to authenticate with
	URL           string          // URL specifies the default github url in case of GitHub enterprise
	Username      string          // Username specifies the username used to authenticate with Github API
	User          string          // User specific the user in git commit messages
	PullRequest   PullRequestSpec // Deprecated since https://github.com/updatecli/updatecli/issues/260, must be clean up
}

// Github contains settings to interact with Github
type Github struct {
	Spec          Spec          // Spec contains inputs coming from updatecli configuration
	HeadBranch    string        // remoteBranch is used when creating a temporary branch before opening a PR
	Force         bool          // Force is used during the git push phase to run `git push --force`.
	CommitMessage commit.Commit // CommitMessage represents conventional commit metadata as type or scope, used to generate the final commit message.
	client        GitHubClient
}

// New returns a new valid Github object.
func New(s Spec) (*Github, error) {
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

	if strings.HasSuffix(s.URL, "github.com") {
		return &Github{
			Spec:   s,
			client: githubv4.NewClient(httpClient),
		}, nil
	}

	return &Github{
		Spec:   s,
		client: githubv4.NewEnterpriseClient(os.Getenv(s.Token), httpClient),
	}, nil
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
