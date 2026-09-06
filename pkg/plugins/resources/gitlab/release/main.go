package release

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
	// [S][C][T] Owner specifies repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// [S][C][T]Repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// [S] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// [S] Age defines the minimum or maximum age of a release to be considered valid.
	// It accepts a duration string (e.g., "24h", "7d", "3w", "1y").
	// The age of a release is the date at which it was released, or the date at which
	// it was created for a release published without one.
	Age age.Spec `yaml:",omitempty"`
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

// SearchReleases retrieves the release tags from a remote GitLab repository, keeping
// only the ones released inside the provided age window.
func (g *Gitlab) SearchReleases(releaseAge age.Spec) ([]string, error) {

	ctx := context.Background()
	// Timeout api query after 30sec
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	results := []string{}
	// Tracks whether the repository publishes releases at all, so that a running
	// cooldown isn't reported as a repository without any release.
	foundRelease := false
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
			if releases[i].UpcomingRelease {
				continue
			}
			foundRelease = true

			if !releaseAge.IsZero() {
				date, ok := releaseDate(releases[i])
				if !ok {
					logrus.Debugf("ignoring release %q, which carries no date, as the age filter cannot be applied to it", releases[i].TagName)
					continue
				}
				if !releaseAge.Matches(date) {
					logrus.Debugf("ignoring release %q, dated %s, as outside of the age window", releases[i].TagName, date)
					continue
				}
			}

			results = append(results, releases[i].TagName)
		}

		// Means that we parsed all pages
		if int64(page) >= resp.NextPage {
			break
		}
		page++
	}

	/*
		The repository does publish releases but the age filter discarded every one of
		them, which means the release we would have returned is still cooling down.
		That's not a lookup failure, so the sentinel lets the caller skip rather than
		fail.
	*/
	if !releaseAge.IsZero() && foundRelease && len(results) == 0 {
		return nil, fmt.Errorf("%w for the GitLab releases of %s", age.ErrNoVersionMatchingAge, g.getPID())
	}

	return results, nil
}

// releaseDate returns the date at which a release became public, and whether GitLab
// reported one. The release date is left empty for a release created without one, so
// the creation date is used as a fallback.
func releaseDate(r *gitlab.Release) (time.Time, bool) {
	if r.ReleasedAt != nil && !r.ReleasedAt.IsZero() {
		return *r.ReleasedAt, true
	}
	if r.CreatedAt != nil && !r.CreatedAt.IsZero() {
		return *r.CreatedAt, true
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
		Age:           g.spec.Age,
		Tag:           g.spec.Tag,
	}
}

func (g *Gitlab) getPID() string {
	return strings.Join([]string{
		g.spec.Owner,
		g.spec.Repository}, "/")
}
