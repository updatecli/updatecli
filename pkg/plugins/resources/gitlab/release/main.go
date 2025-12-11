package release

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitlab/client"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// Spec defines settings used to interact with GitLab release
type Spec struct {
	client.Spec `yaml:",inline,omitempty"`
	// [S][C][T] Owner specifies repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// [S][C][T]Repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// [S] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// [T] Title defines the GitLab release title.
	Title string `yaml:",omitempty"`
	// [C][T] Tag defines the GitLab release tag.
	Tag string `yaml:",omitempty"`
	// [T] Commitish defines the commit-ish such as `main`
	Commitish string `yaml:",omitempty"`
	// [T] Description defines if the new release description
	Description string `yaml:",omitempty"`
	// [T] Draft defines if the release is a draft release
	Draft bool `yaml:",omitempty"`
	// [T] Prerelease defines if the release is a pre-release release
	Prerelease bool `yaml:",omitempty"`
}

const (
	// #nosec g101
	// updatecliCredits contains the message displayed at the end of a newly credit release
	updatecliCredits string = "Made with ❤️️ by updatecli"
)

// Gitlab contains information to interact with GitLab api
type Gitlab struct {
	// spec contains inputs coming from updatecli configuration
	spec Spec
	// client handle the api authentication
	client       client.Client
	foundVersion version.Version
	// Holds the "valid" version.filter, that might be different than the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	Owner         string `yaml:",omitempty" jsonschema:"required"`
	// Repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
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

// Retrieve git tags from a remote GitLab repository
func (g *Gitlab) SearchReleases() ([]string, error) {

	ctx := context.Background()
	// Timeout api query after 30sec
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	results := []string{}
	page := 0
	for {
		opt := &gitlab.ListReleasesOptions{ListOptions: gitlab.ListOptions{Page: int64(page), PerPage: 30}}

		releases, resp, err := g.client.Releases.ListReleases(
			g.getPID(),
			opt,
			gitlab.WithContext(ctx),
		)

		if err != nil {
			return nil, err
		}

		if resp.StatusCode > 400 {
			logrus.Debugf("GitLab Api Response:\n%+v", resp)
		}

		for i := len(releases) - 1; i >= 0; i-- {
			if !releases[i].UpcomingRelease {
				results = append(results, releases[i].TagName)
			}
		}

		// Means that we parsed all pages
		if int64(page) >= resp.NextPage {
			break
		}
		page++
	}

	return results, nil
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

	if len(missingParameters) > 0 {
		logrus.Errorf("missing parameter(s) [%s]", strings.Join(missingParameters, ","))
	}

	if gotError {
		return fmt.Errorf("wrong GitLab configuration")
	}

	return nil
}

// ReportConfig returns a new configuration with only the necessary fields
// to identify the resource without any sensitive information
// and context specific data.
func (g *Gitlab) ReportConfig() interface{} {
	return Spec{
		Owner: g.spec.Owner,
		Spec: client.Spec{
			URL: redact.URL(g.spec.URL),
		},
		Repository:    g.spec.Repository,
		VersionFilter: g.spec.VersionFilter,
		Tag:           g.spec.Tag,
	}
}

func (g *Gitlab) getPID() string {
	return strings.Join([]string{
		g.Owner,
		g.Repository}, "/")
}
