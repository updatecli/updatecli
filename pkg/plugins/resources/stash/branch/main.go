package branch

import (
	"context"
	"strings"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/stash/client"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Stash contains information to interact with Stash api
type Stash struct {
	// spec contains inputs coming from updatecli configuration
	spec Spec
	// client handle the api authentication
	client        client.Client
	HeadBranch    string
	foundVersion  version.Version
	versionFilter version.Filter
}

// New returns a new valid Bitbucket object.
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

// Retrieve bitbucket branches from a remote bitbucket repository
func (g *Stash) SearchBranches() (tags []string, err error) {

	// Timeout api query after 30sec
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	branches, resp, err := g.client.Git.ListBranches(
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
		logrus.Debugf("RC: %q\nBody:\n%s", resp.Status, resp.Body)
	}

	results := []string{}
	for _, branch := range branches {
		results = append(results, branch.Name)
	}

	return results, nil
}
