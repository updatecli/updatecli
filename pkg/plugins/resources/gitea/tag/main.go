package tag

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines settings used to interact with Gitea release
type Spec struct {
	client.Spec `yaml:",inline,omitempty"`
	// [S][C][T] Owner specifies repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// [S][C][T] Repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// [S][C][T] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// [S][C][T] Title defines the Gitea release title.
}

// Gittea contains information to interact with Gitea api
type Gitea struct {
	// Spec contains inputs coming from updatecli configuration
	Spec Spec
	// client handle the api authentication
	client        client.Client
	HeadBranch    string
	foundVersion  version.Version
	versionFilter version.Filter
}

// New returns a new valid Gitea object.
func New(s Spec) (*Gitea, error) {
	err := s.Validate()

	if err != nil {
		return &Gitea{}, err
	}

	c, err := client.New(client.Spec{
		URL:   s.URL,
		Token: s.Token,
	})

	if err != nil {
		return &Gitea{}, err
	}

	newFilter, err := s.VersionFilter.Init()
	if err != nil {
		return &Gitea{}, err
	}

	g := Gitea{
		Spec:          s,
		client:        c,
		versionFilter: newFilter,
	}

	return &g, nil

}

// Retrieve git tags from a remote gitea repository
func (g *Gitea) SearchTags() (tags []string, err error) {

	// Timeout api query after 30sec
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	references, resp, err := g.client.Git.ListTags(
		ctx,
		strings.Join([]string{g.Spec.Owner, g.Spec.Repository}, "/"),
		scm.ListOptions{
			URL:  g.Spec.URL,
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

	for _, ref := range references {
		tags = append(tags, ref.Name)
	}

	return tags, nil
}

func (s *Spec) Validate() error {
	gotError := false
	missingParameters := []string{}

	err := s.ValidateClient()

	if err != nil {
		gotError = true
	}

	if (s.VersionFilter == version.Filter{}) {
		newFilter, err := s.VersionFilter.Init()
		if err != nil {
			return err
		}
		s.VersionFilter = newFilter
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
		return fmt.Errorf("wrong gitea configuration")
	}

	return nil
}
