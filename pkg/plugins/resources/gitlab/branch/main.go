package branch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitlab/client"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// Spec defines settings used to interact with GitLab release
type Spec struct {
	client.Spec `yaml:",inline,omitempty"`
	// [S][C] Owner specifies repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// [S][C] Repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// [S] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// [S] Age defines the minimum or maximum age of a branch to be considered valid.
	// It accepts a duration string (e.g., "24h", "7d", "3w", "1y").
	// The age of a branch is the committer date of its latest commit.
	Age age.Spec `yaml:",omitempty"`
	// [C] Branch specifies the branch name
	Branch string `yaml:",omitempty"`
}

// Gitlab contains information to interact with GitLab api
type Gitlab struct {
	// spec contains inputs coming from updatecli configuration
	spec Spec
	// client handle the api authentication
	client        client.Client
	HeadBranch    string
	foundVersion  version.Version
	versionFilter version.Filter
}

// New returns a new valid GitLab object.
func New(spec interface{}) (*Gitlab, error) {

	var s Spec
	var clientSpec client.Spec

	// mapstructure.Decode cannot handle embedded fields
	// hence we decode it in two steps
	err := mapstructure.Decode(spec, &clientSpec)
	if err != nil {
		return &Gitlab{}, err
	}

	err = mapstructure.Decode(spec, &s)
	if err != nil {
		return &Gitlab{}, nil
	}

	s.Spec = clientSpec

	err = s.Validate()

	if err != nil {
		return &Gitlab{}, err
	}

	c, err := client.New(clientSpec)

	if err != nil {
		return &Gitlab{}, err
	}

	newFilter, err := s.VersionFilter.Init()
	if err != nil {
		return &Gitlab{}, err
	}
	s.VersionFilter = newFilter

	g := Gitlab{
		spec:          s,
		client:        c,
		versionFilter: newFilter,
	}

	return &g, nil

}

// SearchBranches retrieves the branches of a remote GitLab repository, keeping only
// the ones whose latest commit falls inside the provided age window.
func (g *Gitlab) SearchBranches(branchAge age.Spec) (tags []string, err error) {

	// Timeout api query after 30sec
	ctx := context.Background()
	results := []string{}
	// Tracks whether the repository holds branches at all, so that a running cooldown
	// isn't reported as a repository without any branch.
	foundBranch := false
	page := 0
	for {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		opt := &gitlab.ListBranchesOptions{ListOptions: gitlab.ListOptions{Page: int64(page), PerPage: 30}}

		branches, resp, err := g.client.Branches.ListBranches(
			g.getPID(),
			opt,
			gitlab.WithContext(ctx),
		)

		if err != nil {
			return nil, err
		}

		if resp.StatusCode > 400 {
			logrus.Debugf("RC: %q\nBody:\n%s", resp.Status, resp.Body)
		}

		for _, branch := range branches {
			foundBranch = true

			if !branchAge.IsZero() {
				date, ok := branchDate(branch)
				if !ok {
					logrus.Debugf("ignoring branch %q, which carries no date, as the age filter cannot be applied to it", branch.Name)
					continue
				}
				if !branchAge.Matches(date) {
					logrus.Debugf("ignoring branch %q, dated %s, as outside of the age window", branch.Name, date)
					continue
				}
			}

			results = append(results, branch.Name)
		}
		// if the next page is 0 then it means we visited all pages
		if int64(page) >= resp.NextPage {
			break
		}
		page++
	}

	/*
		The repository does hold branches but the age filter discarded every one of them,
		which means the branch we would have returned is still cooling down. That's not a
		lookup failure, so the sentinel lets the caller skip rather than fail.
	*/
	if !branchAge.IsZero() && foundBranch && len(results) == 0 {
		return nil, fmt.Errorf("%w for the GitLab branches of %s", age.ErrNoVersionMatchingAge, g.getPID())
	}

	return results, nil
}

// branchDate returns the committer date of the latest commit of a branch, and whether
// GitLab reported one.
func branchDate(b *gitlab.Branch) (time.Time, bool) {
	if b.Commit == nil {
		return time.Time{}, false
	}
	if b.Commit.CommittedDate != nil && !b.Commit.CommittedDate.IsZero() {
		return *b.Commit.CommittedDate, true
	}
	if b.Commit.AuthoredDate != nil && !b.Commit.AuthoredDate.IsZero() {
		return *b.Commit.AuthoredDate, true
	}
	return time.Time{}, false
}

func (s Spec) Validate() error {
	gotError := false
	missingParameters := []string{}

	if len(s.Owner) == 0 {
		gotError = true
		missingParameters = append(missingParameters, "owner")
	}

	if len(s.Repository) == 0 {
		gotError = true
		missingParameters = append(missingParameters, "repository")
	}

	if err := s.Age.Validate(); err != nil {
		gotError = true
		logrus.Errorln(err)
	}

	if len(missingParameters) > 0 {
		logrus.Errorf("missing parameter(s) [%s]", strings.Join(missingParameters, ","))
	}

	if gotError {
		return fmt.Errorf("wrong GitLab configuration")
	}

	return nil
}

// ReportConfiguration returns a new configuration with only the necessary fields
// to identify the resource without any sensitive information
// and context specific data.
func (g *Gitlab) ReportConfig() interface{} {
	return Spec{
		Owner:         g.spec.Owner,
		Repository:    g.spec.Repository,
		VersionFilter: g.spec.VersionFilter,
		Age:           g.spec.Age,
		Branch:        g.spec.Branch,
		Spec: client.Spec{
			URL: redact.URL(g.spec.URL),
		},
	}
}

func (g *Gitlab) getPID() string {
	return strings.Join([]string{
		g.spec.Owner,
		g.spec.Repository}, "/")
}
