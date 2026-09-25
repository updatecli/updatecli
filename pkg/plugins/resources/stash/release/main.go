package release

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/stash/client"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

/*
"stash/release" defines the specification for managing releases of a Bitbucket Server repository.
It can be used as a "source", a "condition", or a "target".
*/
type Spec struct {
	client.Spec `yaml:",inline,omitempty"`
	// "owner" defines the repository owner.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "owner" is required.
	//
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// "repository" defines the repository name.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "repository" is required.
	//
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// "versionfilter" defines the version pattern and its kind, such as "regex", "semver" or "latest".
	//
	// compatible:
	//   * source
	//
	// default:
	//   latest
	//
	VersionFilter version.Filter `yaml:",omitempty"`
	// "title" defines the release title.
	//
	// compatible:
	//   * target
	//
	// default:
	//   the value of "tag".
	//
	Title string `yaml:",omitempty"`
	// "tag" defines the release tag.
	//
	// compatible:
	//   * condition
	//   * target
	//
	// default:
	//   in a target, the output of the associated source.
	//
	Tag string `yaml:",omitempty"`
	// "commitish" defines the commit-ish the release is created from.
	//
	// compatible:
	//   * target
	//
	// default:
	//   main
	//
	// example:
	//   * commitish: main
	//
	Commitish string `yaml:",omitempty"`
	// "description" defines the release description.
	//
	// compatible:
	//   * target
	//
	Description string `yaml:",omitempty"`
	// "draft" defines whether the release is a draft.
	//
	// compatible:
	//   * target
	//
	// default:
	//   false
	//
	Draft bool `yaml:",omitempty"`
	// "prerelease" defines whether the release is a prerelease.
	//
	// compatible:
	//   * target
	//
	// default:
	//   false
	//
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

// New returns a new valid Stash object.
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

// ReportConfig returns a new configuration object with only the necessary fields
// to identify the resource without any sensitive information or context specific data.
func (s *Stash) ReportConfig() interface{} {
	return Spec{
		Owner:         s.spec.Owner,
		Repository:    s.spec.Repository,
		VersionFilter: s.spec.VersionFilter,
		Title:         s.spec.Title,
		Tag:           s.spec.Tag,
		Draft:         s.spec.Draft,
		Prerelease:    s.spec.Prerelease,
		Description:   s.spec.Description,
	}
}
