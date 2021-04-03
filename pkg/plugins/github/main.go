package github

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/olblak/updateCli/pkg/core/tmp"
	"github.com/olblak/updateCli/pkg/plugins/version"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// Github contains settings to interact with Github
type Github struct {
	Owner                  string
	Description            string
	PullRequestDescription Changelog `yaml:"-"`
	Repository             string
	Username               string
	Token                  string
	URL                    string
	Version                string         // **Deprecated** Version is deprecated in favor of `versionFilter.pattern`, this field will be removed in a futur version
	VersionFilter          version.Filter //VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	Directory              string
	Branch                 string
	remoteBranch           string
	User                   string
	Email                  string
}

// Check verifies if mandatory Github parameters are provided and return false if not.
func (g *Github) Check() (errs []error) {
	required := []string{}

	if g.Token == "" {
		required = append(required, "token")
	}

	if g.Owner == "" {
		required = append(required, "owner")
	}

	if g.Repository == "" {
		required = append(required, "repository")
	}

	if len(g.VersionFilter.Pattern) == 0 {
		g.VersionFilter.Pattern = g.Version
	}

	if err := g.VersionFilter.Validate(); err != nil {
		errs = append(errs, err)
	}

	if len(g.Version) > 0 {
		logrus.Warningln("**Deprecated** Field `version` from resource githubRelease is deprecated in favor of `versionFilter.pattern`, this field will be removed in the next major version")
	}

	if len(required) > 0 {
		errs = append(errs, fmt.Errorf("github parameter(s) required: [%v]", strings.Join(required, ",")))
	}

	return errs
}

func (g *Github) setDirectory() {
	if g.Directory == "" {
		g.Directory = path.Join(tmp.Directory, g.Owner, g.Repository)
	}

	if _, err := os.Stat(g.Directory); os.IsNotExist(err) {

		err := os.MkdirAll(g.Directory, 0755)
		if err != nil {
			logrus.Errorf("err - %s", err)
		}
	}
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
		logrus.Errorf("err - %s", err)
		return "", err
	}

	return query.Repository.ID, nil

}
