package tag

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
"stash/tag" defines the specification for retrieving tags from a Bitbucket Server repository.
It can be used as a "source" or a "condition".
*/
type Spec struct {
	client.Spec `yaml:",inline,omitempty"`
	// "owner" defines the repository owner.
	//
	// compatible:
	//   * source
	//   * condition
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
	// "tag" defines the tag name to check.
	//
	// compatible:
	//   * condition
	//
	Tag string `yaml:",omitempty"`
}

// Stash contains information to interact with Stash api
type Stash struct {
	// spec contains inputs coming from updatecli configuration
	spec Spec
	// client handle the api authentication
	client        client.Client
	foundVersion  version.Version
	versionFilter version.Filter
}

// New returns a new valid Bitbucket Server object.
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
		return &Stash{}, err
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

// Retrieve git tags from a remote Bitbucket Server repository
func (g *Stash) SearchTags() (tags []string, err error) {
	// Timeout api query after 30sec
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	references, resp, err := g.client.Git.ListTags(
		ctx,
		strings.Join([]string{g.spec.Owner, g.spec.Repository}, "/"),
		scm.ListOptions{
			URL:  g.spec.URL,
			Page: 1,
			Size: 30,
		},
	)
	if err != nil {
		return nil, err
	}

	if resp.Status > 400 {
		logrus.Debugf("RC: %d\nBody:\n%s", resp.Status, resp.Body)
	}

	for _, ref := range references {
		tags = append(tags, ref.Name)
	}

	return tags, nil
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
		Tag:           s.spec.Tag,
	}
}
