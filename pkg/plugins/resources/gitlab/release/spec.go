package release

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitlab/client"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
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

func (s Spec) Atomic() Spec {
	return Spec{
		Owner:      s.Owner,
		Repository: s.Repository,
		Tag:        s.Tag,
	}

}
