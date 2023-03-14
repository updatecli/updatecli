package azure

import (
	"fmt"
	"path"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/tmp"
	"github.com/updatecli/updatecli/pkg/plugins/resources/azure/devops/client"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/commit"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"

	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// Spec defines settings used to interact with Azure release
type Spec struct {
	client.Spec `yaml:",inline,omitempty"`
	// CommitMessage represents conventional commit metadata as type or scope, used to generate the final commit message.
	CommitMessage commit.Commit `yaml:",omitempty"`
	// Directory specifies where the github repository is cloned on the local disk
	Directory string `yaml:",omitempty"`
	// Email specifies which emails to use when creating commits
	Email string `yaml:",omitempty"`
	// Force is used during the git push phase to run `git push --force`.
	Force bool `yaml:",omitempty"`
	// GPG key and passphrased used for commit signing
	GPG sign.GPGSpec `yaml:",omitempty"`
	// User specifies the user of the git commit messages
	User string `yaml:",omitempty"`
	// Branch specifies which Azure repository branch to work on
	Branch string `yaml:",omitempty"`
}

// Azure contains information to interact with Azure api
type Azure struct {
	// Spec contains inputs coming from updatecli configuration
	Spec Spec
	// client handle the api authentication
	client           client.Client
	HeadBranch       string
	nativeGitHandler gitgeneric.GitHandler
}

// New returns a new valid Azure object.
func New(spec interface{}, pipelineID string) (*Azure, error) {
	var s Spec
	var clientSpec client.Spec

	// mapstructure.Decode cannot handle embedded fields
	// hence we decode it in two steps
	err := mapstructure.Decode(spec, &clientSpec)
	if err != nil {
		return &Azure{}, err
	}

	err = mapstructure.Decode(spec, &s)
	if err != nil {
		return &Azure{}, nil
	}

	s.Spec = clientSpec

	err = s.Validate()

	if err != nil {
		return &Azure{}, err
	}

	if s.Directory == "" {
		s.Directory = path.Join(tmp.Directory, "azure", s.Owner, s.Project, s.RepoID)
	}

	if len(s.Branch) == 0 {
		logrus.Warningf("no git branch specified, fallback to %q", "main")
		s.Branch = "main"
	}

	c, err := client.New(clientSpec)

	if err != nil {
		return &Azure{}, err
	}

	nativeGitHandler := gitgeneric.GoGit{}
	g := Azure{
		Spec:             s,
		client:           c,
		HeadBranch:       nativeGitHandler.SanitizeBranchName(fmt.Sprintf("updatecli_%v", pipelineID)),
		nativeGitHandler: nativeGitHandler,
	}

	g.setDirectory()

	return &g, nil

}
