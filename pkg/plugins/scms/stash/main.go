package stash

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/tmp"
	"github.com/updatecli/updatecli/pkg/plugins/resources/stash/client"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/commit"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"

	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

/*
"stash" defines the specification for a Bitbucket Server (stash) repository used as an scm.
Updatecli clones the repository, reads files from it, and commits and pushes the changes made by targets.
*/
type Spec struct {
	client.Spec `yaml:",inline,omitempty"`
	// "commitmessage" defines the settings used to generate commit messages.
	//
	// remark:
	//   * the settings apply to every target using this scm.
	//
	CommitMessage commit.Commit `yaml:",omitempty"`
	// "directory" defines the local path where the git repository is cloned.
	//
	// default:
	//   a directory under the Updatecli temporary directory, such as
	//   "/tmp/updatecli/stash/<owner>/<repository>" on Linux.
	//
	// remark:
	//   * keep the default value unless you have a good reason to change it,
	//     as Updatecli may delete the directory after a pipeline run.
	//
	Directory string `yaml:",omitempty"`
	// "depth" defines the depth used when cloning the git repository.
	//
	// default:
	//   empty, which means a full clone.
	//
	// remark:
	//   * a value greater than 0 creates a shallow clone, so Updatecli cannot see the full git history.
	//     Pushing changes may then fail, in which case setting "force" to true may be needed.
	//   * a negative value is rejected.
	//
	// example:
	//   * depth: 1
	//
	Depth *int `yaml:",omitempty"`
	// "singlebranch" defines whether Updatecli clones and fetches only the configured branch,
	// instead of every branch, tag and other reference of the remote.
	//
	// default:
	//   false
	//
	// remark:
	//   * enabling it can make operations much faster on repositories with many branches, tags
	//     or other references, because Updatecli skips the fetch that mirrors every remote reference.
	//   * in some edge cases, Updatecli may then miss a working branch that was already pushed,
	//     and open a duplicate pull request.
	//
	SingleBranch *bool `yaml:",omitempty"`
	// "email" defines the email address used to author commits.
	//
	// default:
	//   updatecli-bot@updatecli.io
	//
	Email string `yaml:",omitempty"`
	// "force" defines whether Updatecli runs `git push --force` when pushing changes.
	//
	// default:
	//   true
	//
	// remark:
	//   * when true, Updatecli also recreates the working branches that diverged from their base branch.
	//   * when "workingbranch" is false and "force" is not set, the Bitbucket Server scm returns an error,
	//     to avoid force pushing to "branch" by mistake. Set "force" explicitly to confirm the behaviour.
	//
	Force *bool `yaml:",omitempty"`
	// "gpg" defines the GPG key and passphrase used to sign commits.
	//
	GPG sign.GPGSpec `yaml:",omitempty"`
	// "owner" defines the owner of the repository.
	//
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// "repository" defines the name of the repository.
	//
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// "user" defines the name used to author commits.
	//
	// default:
	//   updatecli-bot
	//
	User string `yaml:",omitempty"`
	// "branch" defines the git branch to work on.
	//
	// default:
	//   main
	//
	// remark:
	//   * when the Bitbucket Server scm is used by a source or a condition, files are read from this branch.
	//   * when the Bitbucket Server scm is used by a target, Updatecli pushes changes to a working branch
	//     based on this branch, named "updatecli_<branch>_<pipelineid>" by default.
	//   * set "workingbranch" to false to push changes directly to this branch.
	//
	// example:
	//   * branch: main
	//
	Branch string `yaml:",omitempty"`
	// "workingbranchprefix" defines the prefix of the working branch name.
	//
	// default:
	//   updatecli
	//
	// remark:
	//   * the working branch name joins the prefix, the target branch and the pipeline ID,
	//     separated by "workingbranchseparator".
	//   * when set to an empty string, the name starts with the separator, for example "_main_<pipelineid>".
	//
	WorkingBranchPrefix *string `yaml:",omitempty"`
	// "workingbranchseparator" defines the separator between the parts of the working branch name.
	//
	// default:
	//   _
	//
	WorkingBranchSeparator *string `yaml:",omitempty"`
	// "submodules" defines whether Updatecli clones the git submodules of the repository.
	//
	// default:
	//   true
	//
	Submodules *bool `yaml:",omitempty"`
	// "workingbranch" defines whether Updatecli pushes changes to a temporary working branch
	// based on "branch", instead of pushing to "branch" directly.
	//
	// default:
	//   true
	//
	WorkingBranch *bool `yaml:",omitempty"`
}

// Stash contains information to interact with Stash api
type Stash struct {
	// Spec contains inputs coming from updatecli configuration
	Spec Spec
	// client handle the api authentication
	client                 client.Client
	pipelineID             string
	nativeGitHandler       gitgeneric.GitHandler
	workingBranch          bool
	workingBranchPrefix    string
	workingBranchSeparator string
	force                  bool
}

// New returns a new valid Bitbucket Server object.
func New(spec interface{}, pipelineID string) (*Stash, error) {
	var s Spec
	var clientSpec client.Spec

	// mapstructure.Decode cannot handle embedded fields
	// hence we decode it in two steps
	err := mapstructure.Decode(spec, &clientSpec)
	if err != nil {
		return &Stash{}, err
	}

	err = clientSpec.Sanitize()
	if err != nil {
		return &Stash{}, err
	}

	err = clientSpec.Validate()
	if err != nil {
		return &Stash{}, err
	}

	err = mapstructure.Decode(spec, &s)
	if err != nil {
		return &Stash{}, nil
	}

	s.Spec = clientSpec

	err = s.Validate()
	if err != nil {
		return &Stash{}, err
	}

	if s.Directory == "" {
		s.Directory = path.Join(tmp.Directory, "stash", s.Owner, s.Repository)
	}

	if len(s.Branch) == 0 {
		logrus.Warningf("no git branch specified, fallback to %q", "main")
		s.Branch = "main"
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

	if force {
		if !workingBranch && s.Force == nil {
			errorMsg := fmt.Sprintf(`
Better safe than sorry.

Updatecli may be pushing unwanted changes to the branch %q.

The Stash scm plugin has by default the force option set to true,
The scm force option set to true means that Updatecli is going to run "git push --force"
Some target plugin, like the shell one, run "git commit -A" to catch all changes done by that target.

If you know what you are doing, please set the force option to true in your configuration file to ignore this error message.
`, s.Branch)

			logrus.Errorln(errorMsg)
			return nil, errors.New("unclear configuration, better safe than sorry")

		}
	}

	c, err := client.New(clientSpec)
	if err != nil {
		return &Stash{}, err
	}

	if s.Email == "" {
		s.Email = gitgeneric.DefaultGitCommitEmailAddress
	}

	if s.User == "" {
		s.User = gitgeneric.DefaultGitCommitUserName
	}

	nativeGitHandler := gitgeneric.GoGit{}
	g := Stash{
		Spec:                   s,
		client:                 c,
		pipelineID:             pipelineID,
		nativeGitHandler:       &nativeGitHandler,
		workingBranch:          workingBranch,
		workingBranchPrefix:    workingBranchPrefix,
		workingBranchSeparator: workingBranchSeparator,
		force:                  force,
	}

	g.setDirectory()

	return &g, nil
}

// Retrieve git tags from a remote Bitbucket Server repository
func (s *Stash) SearchTags() (tags []string, err error) {
	// Timeout api query after 30sec
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	references, resp, err := s.client.Git.ListTags(
		ctx,
		strings.Join([]string{s.Spec.Owner, s.Spec.Repository}, "/"),
		scm.ListOptions{
			URL:  s.Spec.URL,
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

func (s *Spec) Validate() error {
	gotError := false
	missingParameters := []string{}

	if len(missingParameters) > 0 {
		logrus.Errorf("missing parameter(s) [%s]", strings.Join(missingParameters, ","))
	}

	if gotError {
		return fmt.Errorf("wrong bitbucket configuration")
	}

	return nil
}
