package release

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/stash/client"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines settings used to interact with Bitbucket release
type Spec struct {
	client.Spec `yaml:",inline,omitempty"`
	// [S][C][T] owner specifies repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// [S][C][T] repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// [S] versionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// [T] title defines the Bitbucket release title.
	Title string `yaml:",omitempty"`
	// [C][T] tag defines the Bitbucket release tag.
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

// Stash contains information to interact with Stash api
type Stash struct {
	// spec contains inputs coming from updatecli configuration
	spec Spec
	// client handle the api authentication
	client       client.Client
	foundVersion version.Version
	// Holds the "valid" version.filter, that might be different than the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New returns a new valid GitHub object.
func New(spec interface{}) (*Stash, error) {
	var s Spec
	var clientSpec client.Spec

	// mapstructure.Decode cannot handle embedded fields
	// hence we decode it in two steps
	err := mapstructure.Decode(spec, &clientSpec)
	if err != nil {
		return &Stash{}, err
	}

	err = mapstructure.Decode(spec, &s)
	if err != nil {
		return &Stash{}, nil
	}

	err = clientSpec.Validate()
	if err != nil {
		return &Stash{}, err
	}

	err = clientSpec.Sanitize()
	if err != nil {
		return &Stash{}, err
	}

	s.Spec = clientSpec

	err = s.Validate()

	if err != nil {
		return &Stash{}, err
	}

	c, err := client.New(clientSpec)

	if err != nil {
		return &Stash{}, err
	}

	newFilter, err := s.VersionFilter.Init()
	if err != nil {
		return &Stash{}, err
	}
	s.VersionFilter = newFilter

	g := Stash{
		spec:          s,
		client:        c,
		versionFilter: newFilter,
	}

	return &g, nil
}

// Retrieve git tags from a remote bitbucket repository
func (g *Stash) SearchReleases() ([]string, error) {

	ctx := context.Background()
	// Timeout api query after 30sec
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	releases, resp, err := g.client.Releases.List(
		ctx,
		strings.Join([]string{g.spec.Owner, g.spec.Repository}, "/"),
		scm.ReleaseListOptions{
			Page:   1,
			Size:   30,
			Open:   true,
			Closed: true,
		},
	)

	if err != nil {
		return nil, err
	}

	if resp.Status > 400 {
		logrus.Debugf("Bitbucket Api Response:\n%+v", resp)
	}

	results := []string{}
	for i := len(releases) - 1; i >= 0; i-- {
		if !releases[i].Draft {
			results = append(results, releases[i].Tag)
		}
	}

	return results, nil
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

	if len(missingParameters) > 0 {
		logrus.Errorf("missing parameter(s) [%s]", strings.Join(missingParameters, ","))
	}

	if gotError {
		return fmt.Errorf("wrong bitbucket configuration")
	}

	return nil
}
