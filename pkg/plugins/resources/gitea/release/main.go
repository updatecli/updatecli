package release

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines settings used to interact with Gitea release
type Spec struct {
	client.Spec `yaml:",inline,omitempty"`
	// [S][C][T] owner specifies the repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// [S][C][T] repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// [S] versionfilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// [S] age defines the minimum or maximum age of a release to be considered valid.
	// It accepts a duration string (e.g., "24h", "7d", "3w", "1y").
	Age age.Spec `yaml:",omitempty"`
	// [T] title defines the Gitea release title.
	Title string `yaml:",omitempty"`
	// [C][T] tag defines the Gitea release tag.
	Tag string `yaml:",omitempty"`
	// [T] commitish defines the commit-ish such as `main`
	Commitish string `yaml:",omitempty"`
	// [T] description defines if the new release description
	Description string `yaml:",omitempty"`
	// [T] draft defines if the release is a draft release
	Draft bool `yaml:",omitempty"`
	// [T] prerelease defines if the release is a pre-release release
	Prerelease bool `yaml:",omitempty"`
}

const (
	// #nosec g101
	// updatecliCredits contains the message displayed at the end of a newly credit release
	updatecliCredits string = "Made with ❤️️ by updatecli"
)

// Gitea contains information to interact with Gitea api
type Gitea struct {
	// spec contains inputs coming from updatecli configuration
	spec Spec
	// client handle the api authentication
	client       client.Client
	foundVersion version.Version
	// Holds the "valid" version.filter, that might be different than the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New returns a new valid GitHub object.
func New(spec interface{}) (*Gitea, error) {
	var s Spec
	var clientSpec client.Spec

	// mapstructure.Decode cannot handle embedded fields
	// hence we decode it in two steps
	err := mapstructure.Decode(spec, &clientSpec)
	if err != nil {
		return &Gitea{}, err
	}

	err = mapstructure.Decode(spec, &s)
	if err != nil {
		return &Gitea{}, nil
	}

	err = clientSpec.Validate()
	if err != nil {
		return &Gitea{}, err
	}

	err = clientSpec.Sanitize()
	if err != nil {
		return &Gitea{}, err
	}

	s.Spec = clientSpec

	err = s.Validate()

	if err != nil {
		return &Gitea{}, err
	}

	c, err := client.New(clientSpec)

	if err != nil {
		return &Gitea{}, err
	}

	newFilter, err := s.VersionFilter.Init()
	if err != nil {
		return &Gitea{}, err
	}
	s.VersionFilter = newFilter

	g := Gitea{
		spec:          s,
		client:        c,
		versionFilter: newFilter,
	}

	return &g, nil
}

// SearchReleases retrieves the release tags from a remote gitea repository,
// keeping only the ones published inside the provided age window.
func (g *Gitea) SearchReleases(ctx context.Context, releaseAge age.Spec) ([]string, error) {

	results := []string{}
	// Tracks whether the repository publishes releases at all, so that a running
	// cooldown isn't reported as a repository without any release.
	foundRelease := false
	page := 0
	for {
		// Timeout api query after 30sec
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		releases, resp, err := g.client.Releases.List(
			ctx,
			strings.Join([]string{g.spec.Owner, g.spec.Repository}, "/"),
			scm.ReleaseListOptions{
				Page:   page,
				Size:   30,
				Open:   true,
				Closed: true,
			},
		)

		if err != nil {
			return nil, err
		}

		if resp.Status > 400 {
			logrus.Debugf("Gitea Api Response:\n%+v", resp)
		}

		for i := len(releases) - 1; i >= 0; i-- {
			if releases[i].Draft {
				continue
			}
			foundRelease = true
			date := releaseDate(releases[i])
			if !releaseAge.Matches(date) {
				logrus.Debugf("ignoring release %q, dated %s, as outside of the age window", releases[i].Tag, date)
				continue
			}
			results = append(results, releases[i].Tag)
		}

		if page >= resp.Page.Last {
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
	if foundRelease && len(results) == 0 {
		return nil, fmt.Errorf("%w for the Gitea releases of %s/%s", age.ErrNoVersionMatchingAge, g.spec.Owner, g.spec.Repository)
	}

	return results, nil
}

// releaseDate returns the date at which a release became public.
// Gitea leaves the publication date empty for releases created without one, so the
// creation date is used as a fallback.
func releaseDate(r *scm.Release) time.Time {
	if !r.Published.IsZero() {
		return r.Published
	}
	return r.Created
}

func (s Spec) Validate() error {
	gotError := false
	missingParameters := []string{}

	err := s.Spec.Validate()

	if err != nil {
		logrus.Errorln(err)
		gotError = true
	}

	if len(s.Owner) == 0 {
		gotError = true
		missingParameters = append(missingParameters, "owner")
	}

	if len(s.Repository) == 0 {
		gotError = true
		missingParameters = append(missingParameters, "repository")
	}

	if err := s.Age.Validate(); err != nil {
		logrus.Errorln(err)
		gotError = true
	}

	if len(missingParameters) > 0 {
		logrus.Errorf("missing parameter(s) [%s]", strings.Join(missingParameters, ","))
	}

	if gotError {
		return fmt.Errorf("wrong gitea configuration")
	}

	return nil
}

// ReportConfig returns a new configuration with only the necessary configuration fields
// to identify the resource without any sensitive information
func (g *Gitea) ReportConfig() interface{} {
	return Spec{
		Owner:      g.spec.Owner,
		Repository: g.spec.Repository,
		Tag:        g.spec.Tag,
		Spec: client.Spec{
			URL: redact.URL(g.spec.URL),
		},
		VersionFilter: g.spec.VersionFilter,
		Age:           g.spec.Age,
	}
}
