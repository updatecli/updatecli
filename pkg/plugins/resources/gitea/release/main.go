package release

import (
	"context"
	"strings"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

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

// Retrieve git tags from a remote gitea repository
func (g *Gitea) SearchReleases() ([]string, error) {

	ctx := context.Background()

	results := []string{}
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
			if !releases[i].Draft {
				results = append(results, releases[i].Tag)
			}
		}

		if page >= resp.Page.Last {
			break
		}
		page++

	}

	return results, nil
}
