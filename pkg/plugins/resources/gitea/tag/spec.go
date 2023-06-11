package tag

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines settings used to interact with Gitea release
type Spec struct {
	client.Spec `yaml:",inline,omitempty"`
	// [S][C] Owner specifies repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// [S][C] Repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// [S][C] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// [S] Tag defines the Gitea tag .
	Tag string `yaml:",omitempty"`
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
		return fmt.Errorf("wrong gitea configuration")
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
