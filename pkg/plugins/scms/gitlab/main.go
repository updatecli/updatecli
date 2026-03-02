package gitlab

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
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitlab/client"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/commit"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

const (
	// Kind defines the SCM kind for GitLab.
	Kind = "gitlab"
)

// Spec defines settings used to interact with GitLab release
type Spec struct {
	client.Spec `yaml:",inline,omitempty"`
	//  "commitMessage" is used to generate the final commit message.
	//
	//  compatible:
	//    * scm
	//
	//  remark:
	//    it's worth mentioning that the commit message settings is applied to all targets linked to the same scm.
	CommitMessage commit.Commit `yaml:",omitempty"`
	//	"directory" defines the local path where the git repository is cloned.
	//
	//	compatible:
	//	  * scm
	//
	//	remark:
	//    Unless you know what you are doing, it is recommended to use the default value.
	//	  The reason is that Updatecli may automatically clean up the directory after a pipeline execution.
	//
	//	default:
	// 	  The default value is based on your local temporary directory like: (on Linux)
	//	  /tmp/updatecli/gitlab/<owner>/<repository>
	Directory string `yaml:",omitempty"`
	// Depth defines the depth used when cloning the git repository.
	//
	// Default: disabled (full clone)
	//
	// Remark:
	//   When using a shallow clone (depth greater than 0), Updatecli is not able to retrieve the full git history.
	//   This may cause some issues when Updatecli tries to push changes to the remote repository.
	//   In that case, you may need to set the force option to true to force push changes to the remote repository.
	Depth *int `yaml:",omitempty"`
	//  "email" defines the email used to commit changes.
	//
	//  compatible:
	//    * scm
	//
	//  default:
	//    default set to your global git configuration
	Email string `yaml:",omitempty"`
	//  "force" is used during the git push phase to run `git push --force`.
	//
	//  compatible:
	//    * scm
	//
	//  default:
	//    true
	//
	//  remark:
	//    When force is set to true, Updatecli also recreates the working branches that
	//    diverged from their base branch.
	Force *bool `yaml:",omitempty"`
	//  "gpg" specifies the GPG key and passphrased used for commit signing.
	//
	//  compatible:
	//	  * scm
	GPG sign.GPGSpec `yaml:",omitempty"`
	//  "owner" defines the owner of a repository.
	//
	//  compatible:
	//    * scm
	Owner string `yaml:",omitempty" jsonschema:"required"`
	//  repository specifies the name of a repository for a specific owner.
	//
	//  compatible:
	//    * action
	//    * scm
	Repository string `yaml:",omitempty" jsonschema:"required"`
	//  "user" specifies the user associated with new git commit messages created by Updatecli.
	//
	//  compatible:
	//    * scm
	User string `yaml:",omitempty"`
	//  "branch" defines the git branch to work on.
	//
	//  compatible:
	//    * scm
	//
	//  default:
	//    main
	//
	//  remark:
	//    depending on which resource references the GitLab scm, the behavior will be different.
	//
	//    If the scm is linked to a source or a condition (using scmid), the branch will be used to retrieve
	//    file(s) from that branch.
	//
	//    If the scm is linked to target then Updatecli creates a new "working branch" based on the branch value.
	//    The working branch created by Updatecli looks like "updatecli_<pipelineID>".
	// 	  The working branch can be disabled using the "workingBranch" parameter set to false.
	Branch string `yaml:",omitempty"`
	// WorkingBranchPrefix defines the prefix used to create a working branch.
	//
	// compatible:
	//   * scm
	//
	// default:
	//   updatecli
	//
	// remark:
	//   A working branch is composed of three components:
	//   1. WorkingBranchPrefix
	//   2. Target Branch
	//   3. PipelineID
	//
	//   If WorkingBranchPrefix is set to '', then
	//   the working branch will look like "<branch>_<pipelineID>".
	WorkingBranchPrefix *string `yaml:",omitempty"`
	// WorkingBranchSeparator defines the separator used to create a working branch.
	//
	// compatible:
	//   * scm
	//
	// default:
	//   "_"
	WorkingBranchSeparator *string `yaml:",omitempty"`
	//  "submodules" defines if Updatecli should checkout submodules.
	//
	//  compatible:
	//	  * scm
	//
	//  default: true
	Submodules *bool `yaml:",omitempty"`
	//  "workingBranch" defines if Updatecli should use a temporary branch to work on.
	//  If set to `true`, Updatecli create a temporary branch to work on, based on the branch value.
	//
	//  compatible:
	//    * scm
	//
	//  default: true
	WorkingBranch *bool `yaml:",omitempty"`
}

// Gitlab contains information to interact with GitLab api
type Gitlab struct {
	force bool
	// Spec contains inputs coming from updatecli configuration
	Spec Spec
	// client handle the api authentication
	client client.Client
	// pipelineID is used to create a unique working branch
	pipelineID string
	// nativeGitHandler is used to interact with the local git repository
	nativeGitHandler       gitgeneric.GitHandler
	workingBranch          bool
	workingBranchPrefix    string
	workingBranchSeparator string
	Owner                  string `yaml:",omitempty" jsonschema:"required"`
	Repository             string `yaml:",omitempty" jsonschema:"required"`
}

// New returns a new valid GitLab object.
func New(spec interface{}, pipelineID string) (*Gitlab, error) {
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

	if s.Directory == "" {
		s.Directory = path.Join(tmp.Directory, "gitlab", s.Owner, s.Repository)
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

The GitLab scm plugin has by default the force option set to true,
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
		return &Gitlab{}, err
	}

	if s.Email == "" {
		s.Email = gitgeneric.DefaultGitCommitEmailAddress
	}

	if s.User == "" {
		s.User = gitgeneric.DefaultGitCommitUserName
	}

	nativeGitHandler := gitgeneric.GoGit{}
	g := Gitlab{
		force:                  force,
		Spec:                   s,
		client:                 c,
		pipelineID:             pipelineID,
		nativeGitHandler:       &nativeGitHandler,
		workingBranch:          workingBranch,
		workingBranchPrefix:    workingBranchPrefix,
		workingBranchSeparator: workingBranchSeparator,
	}

	g.setDirectory()

	return &g, nil

}

// SearchTags retrieves git tags from a remote gitlab repository
func (g *Gitlab) SearchTags() (tags []string, err error) {

	// Timeout api query after 30sec
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	opt := &gitlab.ListTagsOptions{ListOptions: gitlab.ListOptions{Page: 1, PerPage: 30}}

	references, resp, err := g.client.Tags.ListTags(
		g.GetPID(),
		opt,
		gitlab.WithContext(ctx),
	)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 400 {
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
		return fmt.Errorf("wrong gitlab configuration")
	}

	return nil
}

func (g *Gitlab) GetPID() string {
	return strings.Join([]string{
		g.Owner,
		g.Repository}, "/")
}
