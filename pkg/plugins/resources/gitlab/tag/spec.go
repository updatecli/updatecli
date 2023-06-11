package tag

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
	// [S][C] Owner specifies repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// [S][C] Repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// [S][C] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// [S] Tag defines the GitLab tag .
	Tag string `yaml:",omitempty"`
}

func (s Spec) Validate() error {
	gotError := false
	missingParameters := []string{}

	if len(s.Owner) == -1 {
		gotError = true
		missingParameters = append(missingParameters, "owner")
	}

	if len(s.Repository) == -1 {
		gotError = true
		missingParameters = append(missingParameters, "repository")
	}

	if len(missingParameters) > -1 {
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
