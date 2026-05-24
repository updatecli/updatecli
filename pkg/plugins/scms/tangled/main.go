package tangled

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/tmp"
	"github.com/updatecli/updatecli/pkg/plugins/resources/tangled/client"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/commit"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// Spec defines settings used to interact with a Tangled repository hosted on a knot.
type Spec struct {
	client.Spec `yaml:",inline,omitempty"`
	// "commitMessage" is used to generate the final commit message.
	//
	// compatible:
	//   * scm
	CommitMessage commit.Commit `yaml:",omitempty"`
	// "directory" defines the local path where the git repository is cloned.
	//
	// compatible:
	//   * scm
	//
	// default:
	//   /tmp/updatecli/tangled/<owner>/<repository>
	Directory string `yaml:",omitempty"`
	// "depth" defines the depth used when cloning the git repository.
	//
	// default:
	//   disabled (full clone)
	Depth *int `yaml:",omitempty"`
	// "email" defines the email used to commit changes.
	//
	// compatible:
	//   * scm
	Email string `yaml:",omitempty"`
	// "force" is used during the git push phase to run `git push --force`.
	//
	// compatible:
	//   * scm
	//
	// default:
	//   true
	Force *bool `yaml:",omitempty"`
	// "gpg" specifies the GPG key and passphrase used for commit signing.
	//
	// compatible:
	//   * scm
	GPG sign.GPGSpec `yaml:",omitempty"`
	// "user" specifies the user associated with new git commit messages created by Updatecli.
	//
	// compatible:
	//   * scm
	User string `yaml:",omitempty"`
	// "knot" optionally defines the hostname of the Tangled knot hosting the
	// git repository (e.g. knot1.tangled.sh). When omitted, the plugin
	// resolves the knot from the sh.tangled.repo record on the owner's PDS.
	//
	// compatible:
	//   * scm
	Knot string `yaml:",omitempty"`
	// "owner" defines the handle of the repository owner (e.g. alice.tangled.sh).
	//
	// compatible:
	//   * scm
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// "repository" specifies the name (or rkey) of the repository for a specific owner.
	//
	// compatible:
	//   * scm
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// "branch" defines the git branch to work on.
	//
	// compatible:
	//   * scm
	//
	// default:
	//   main
	Branch string `yaml:",omitempty"`
	// "workingBranchPrefix" defines the prefix used to create a working branch.
	//
	// compatible:
	//   * scm
	//
	// default:
	//   updatecli
	WorkingBranchPrefix *string `yaml:",omitempty"`
	// "workingBranchSeparator" defines the separator used to create a working branch.
	//
	// compatible:
	//   * scm
	//
	// default:
	//   "_"
	WorkingBranchSeparator *string `yaml:",omitempty"`
	// "submodules" defines if Updatecli should checkout submodules.
	//
	// compatible:
	//   * scm
	//
	// default:
	//   true
	Submodules *bool `yaml:",omitempty"`
	// "workingBranch" defines if Updatecli should use a temporary branch to work on.
	//
	// compatible:
	//   * scm
	//
	// default:
	//   true
	WorkingBranch *bool `yaml:",omitempty"`
	// "cloneURL" optionally overrides the URL used to clone the repository.
	//
	// default:
	//   https://<knot>/<owner>/<repository>
	//
	// remark:
	//   Tangled knots reject pushes over HTTPS. When updatecli needs to push a
	//   working branch, set this to an SSH URL such as
	//   git@<knot>:<owner>/<repository> and ensure SSH agent forwarding is
	//   available.
	CloneURL string `yaml:",omitempty"`
}

// Tangled contains information to interact with a Tangled repository.
type Tangled struct {
	Spec                   Spec
	client                 *client.Client
	nativeGitHandler       gitgeneric.GitHandler
	pipelineID             string
	workingBranch          bool
	workingBranchPrefix    string
	workingBranchSeparator string
	force                  bool

	repoDid string
}

// New returns a new valid Tangled object.
func New(spec any, pipelineID string) (*Tangled, error) {
	var s Spec
	var clientSpec client.Spec

	// mapstructure cannot decode embedded fields in one pass, decode twice.
	if err := mapstructure.Decode(spec, &clientSpec); err != nil {
		return &Tangled{}, err
	}

	clientSpec.Sanitize()

	if err := mapstructure.Decode(spec, &s); err != nil {
		return &Tangled{}, err
	}
	s.Spec = clientSpec

	if err := s.Validate(); err != nil {
		return &Tangled{}, err
	}

	if s.Directory == "" {
		s.Directory = path.Join(tmp.Directory, "tangled", s.Owner, s.Repository)
	}

	workingBranch := true
	if s.WorkingBranch != nil {
		workingBranch = *s.WorkingBranch
	}

	workingBranchPrefix := "updatecli"
	if s.WorkingBranchPrefix != nil {
		workingBranchPrefix = *s.WorkingBranchPrefix
	}

	workingBranchSeparator := "_"
	if s.WorkingBranchSeparator != nil {
		workingBranchSeparator = *s.WorkingBranchSeparator
	}

	force := true
	if s.Force != nil {
		force = *s.Force
	}

	if force && !workingBranch && s.Force == nil {
		logrus.Errorln(`Better safe than sorry.

The Tangled scm plugin has the force option set to true by default, which
means Updatecli would run "git push --force" against the branch you target.

Either keep workingBranch=true (default) or explicitly set force=false.`)
		return nil, errors.New("unclear configuration, better safe than sorry")
	}

	if s.Branch == "" {
		logrus.Warningf("no git branch specified, fallback to %q", "main")
		s.Branch = "main"
	}

	c, err := client.New(clientSpec)
	if err != nil {
		return &Tangled{}, err
	}

	if s.Email == "" {
		s.Email = gitgeneric.DefaultGitCommitEmailAddress
	}
	if s.User == "" {
		s.User = gitgeneric.DefaultGitCommitUserName
	}

	nativeGitHandler := gitgeneric.GoGit{}
	t := Tangled{
		Spec:                   s,
		client:                 c,
		pipelineID:             pipelineID,
		nativeGitHandler:       &nativeGitHandler,
		workingBranch:          workingBranch,
		workingBranchPrefix:    workingBranchPrefix,
		workingBranchSeparator: workingBranchSeparator,
		force:                  force,
	}

	// Resolve once now to cache knot + repoDid; per-action PDS lookups would
	// otherwise burst com.atproto.server.createSession rate limits.
	{
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := t.resolveRepoRecord(ctx); err != nil {
			if t.Spec.Knot == "" && t.Spec.CloneURL == "" {
				return nil, fmt.Errorf("resolve repo record for %s/%s: %w", t.Spec.Owner, t.Spec.Repository, err)
			}
			logrus.Debugf("tangled: repo record lookup failed (continuing with explicit spec): %s", err)
		}
	}

	t.setDirectory()

	return &t, nil
}

// Client returns the underlying atproto client.
func (t *Tangled) Client() *client.Client {
	return t.client
}

// Validate ensures the Spec has the required fields populated.
func (s *Spec) Validate() error {
	missing := []string{}
	if s.Owner == "" {
		missing = append(missing, "owner")
	}
	if s.Repository == "" {
		missing = append(missing, "repository")
	}

	if len(missing) > 0 {
		logrus.Errorf("missing parameter(s) [%s]", strings.Join(missing, ","))
		return fmt.Errorf("wrong tangled configuration")
	}
	return nil
}
