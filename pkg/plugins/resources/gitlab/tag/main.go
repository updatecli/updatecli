package tag

import (
	"context"
	"strings"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitlab/client"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Gitlab contains information to interact with GitLab api
type Gitlab struct {
	// spec contains inputs coming from updatecli configuration
	spec Spec
	// client handle the api authentication
	client        client.Client
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

// Retrieve git tags from a remote GitLab repository
func (g *Gitlab) SearchTags() (tags []string, err error) {

	// Timeout api query after 30sec
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	page := 0

	// Query gitlab api until we visit all pages
	for {
		references, resp, err := g.client.Git.ListTags(
			ctx,
			strings.Join([]string{g.spec.Owner, g.spec.Repository}, "/"),
			scm.ListOptions{
				URL:  g.client.BaseURL.Host,
				Page: page,
				Size: 100,
			},
		)

		if err != nil {
			return nil, err
		}

		if resp.Status > 400 {
			logrus.Debugf("RC: %q\nBody:\n%s", resp.Status, resp.Body)
		}

		for _, ref := range references {
			tags = append(tags, ref.Name)
		}

		if page >= resp.Page.Last {
			break
		}
		page++
	}

	return tags, nil
}
