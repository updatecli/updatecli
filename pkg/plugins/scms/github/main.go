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

	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// Spec represents the configuration input
type Spec struct {
	// Branch specifies which github branch to work on
	Branch string `yaml:",omitempty"`
	// Directory specifies where the github repository is cloned on the local disk
	Directory string `yaml:",omitempty"`
	// Email specifies which emails to use when creating commits
	Email string `yaml:",omitempty"`
	// Owner specifies repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// Repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// Token specifies the credential used to authenticate with
	Token string `yaml:",omitempty" jsonschema:"required"`
	// URL specifies the default github url in case of GitHub enterprise
	URL string `yaml:",omitempty"`
	// Username specifies the username used to authenticate with Github API
	Username string `yaml:",omitempty" jsonschema:"required"`
	// User specifies the user of the git commit messages
	User string `yaml:",omitempty"`
	// Deprecated since https://github.com/updatecli/updatecli/issues/260, must be clean up
	PullRequest PullRequestSpec `yaml:",omitempty"`
	// GPG key and passphrased used for commit signing
	GPG sign.GPGSpec `yaml:",omitempty"`
	// Force is used during the git push phase to run `git push --force`.
	Force bool `yaml:",omitempty"`
	// CommitMessage represents conventional commit metadata as type or scope, used to generate the final commit message.
	CommitMessage commit.Commit `yaml:",omitempty"`
}

// Github contains settings to interact with Github
type Github struct {
	// Spec contains inputs coming from updatecli configuration
	Spec Spec
	// HeadBranch is used when creating a temporary branch before opening a PR
	HeadBranch       string
	client           GitHubClient
	nativeGitHandler gitgeneric.GitHandler
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
	nativeGitHandler := gitgeneric.GoGit{}

	g := Github{
		Spec:             s,
		HeadBranch:       nativeGitHandler.SanitizeBranchName(fmt.Sprintf("updatecli_%v", pipelineID)),
		nativeGitHandler: nativeGitHandler,
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

// Merge returns nil if it successfully merges the child Spec into target receiver.
// Please note that child attributes always overrides receiver's
func (gs *Spec) Merge(child interface{}) error {
	childGHSpec, ok := child.(Spec)
	if !ok {
		return fmt.Errorf("unable to merge GitHub spec with unknown object type.")
	}

	if childGHSpec.Branch != "" {
		gs.Branch = childGHSpec.Branch
	}
	if childGHSpec.CommitMessage != (commit.Commit{}) {
		gs.CommitMessage = childGHSpec.CommitMessage
	}
	if childGHSpec.Directory != "" {
		gs.Directory = childGHSpec.Directory
	}
	if childGHSpec.Email != "" {
		gs.Email = childGHSpec.Email
	}
	if childGHSpec.Force {
		gs.Force = childGHSpec.Force
	}
	if childGHSpec.GPG != (sign.GPGSpec{}) {
		gs.GPG = childGHSpec.GPG
	}
	if childGHSpec.Owner != "" {
		gs.Owner = childGHSpec.Owner
	}
	// PullRequest is deprecated so not merging it
	if childGHSpec.Repository != "" {
		gs.Repository = childGHSpec.Repository
	}
	if childGHSpec.Token != "" {
		gs.Token = childGHSpec.Token
	}
	if childGHSpec.URL != "" {
		gs.URL = childGHSpec.URL
	}
	if childGHSpec.User != "" {
		gs.User = childGHSpec.User
	}
	if childGHSpec.Username != "" {
		gs.Username = childGHSpec.Username
	}

	return nil
}

// MergeFromEnv updates the target receiver with the "non zero-ed" environment variables
func (gs *Spec) MergeFromEnv(envPrefix string) {
	prefix := fmt.Sprintf("%s_", envPrefix)
	if os.Getenv(fmt.Sprintf("%s%s", prefix, "BRANCH")) != "" {
		gs.Branch = os.Getenv(fmt.Sprintf("%s%s", prefix, "BRANCH"))
	}
	if os.Getenv(fmt.Sprintf("%s%s", prefix, "DIRECTORY")) != "" {
		gs.Directory = os.Getenv(fmt.Sprintf("%s%s", prefix, "DIRECTORY"))
	}
	if os.Getenv(fmt.Sprintf("%s%s", prefix, "EMAIL")) != "" {
		gs.Email = os.Getenv(fmt.Sprintf("%s%s", prefix, "EMAIL"))
	}
	if os.Getenv(fmt.Sprintf("%s%s", prefix, "OWNER")) != "" {
		gs.Owner = os.Getenv(fmt.Sprintf("%s%s", prefix, "OWNER"))
	}
	if os.Getenv(fmt.Sprintf("%s%s", prefix, "REPOSITORY")) != "" {
		gs.Repository = os.Getenv(fmt.Sprintf("%s%s", prefix, "REPOSITORY"))
	}
	if os.Getenv(fmt.Sprintf("%s%s", prefix, "TOKEN")) != "" {
		gs.Token = os.Getenv(fmt.Sprintf("%s%s", prefix, "TOKEN"))
	}
	if os.Getenv(fmt.Sprintf("%s%s", prefix, "URL")) != "" {
		gs.URL = os.Getenv(fmt.Sprintf("%s%s", prefix, "URL"))
	}
	if os.Getenv(fmt.Sprintf("%s%s", prefix, "USERNAME")) != "" {
		gs.Username = os.Getenv(fmt.Sprintf("%s%s", prefix, "USERNAME"))
	}
	if os.Getenv(fmt.Sprintf("%s%s", prefix, "USER")) != "" {
		gs.User = os.Getenv(fmt.Sprintf("%s%s", prefix, "USER"))
	}
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
