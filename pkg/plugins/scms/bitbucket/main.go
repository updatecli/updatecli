package bitbucket

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/tmp"
	"github.com/updatecli/updatecli/pkg/plugins/resources/bitbucket/client"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/commit"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"

	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// Spec defines settings used to interact with Bitbucket Cloud release
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
	//  "directory" defines the local path where the git repository is cloned.
	//
	//  compatible:
	//    * scm
	//
	//  remark:
	//    Unless you know what you are doing, it is recommended to use the default value.
	//    The reason is that Updatecli may automatically clean up the directory after a pipeline execution.
	//
	//  default:
	//    The default value is based on your local temporary directory like: (on Linux)
	//    /tmp/updatecli/bitbucket/<owner>/<repository>
	Directory string `yaml:",omitempty"`
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
	//    false
	//
	//  remark:
	//    When force is set to true, Updatecli also recreate the working branches that
	//    diverged from their base branch.
	Force *bool `yaml:",omitempty"`
	//	"gpg" specifies the GPG key and passphrased used for commit signing
	//
	//	compatible:
	//		* scm
	GPG sign.GPGSpec `yaml:",omitempty"`
	//	"owner" defines the owner of a repository.
	//
	//	compatible:
	//		* scm
	Owner string `yaml:",omitempty" jsonschema:"required"`
	//	repository specifies the name of a repository for a specific owner.
	//
	//	compatible:
	//		* scm
	Repository string `yaml:",omitempty" jsonschema:"required"`
	//	"user" specifies the user associated with new git commit messages created by Updatecli
	//
	//	compatible:
	//		* scm
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
	//    depending on which resource references the Bitbucket Cloud scm, the behavior will be different.
	//
	//    If the scm is linked to a source or a condition (using scmid), the branch will be used to retrieve
	//    file(s) from that branch.
	//
	//    If the scm is linked to target then Updatecli creates a new "working branch" based on the branch value.
	//    The working branch created by Updatecli looks like "updatecli_<pipelineID>".
	//    The working branch can be disabled using the "workingBranch" parameter set to false.
	Branch string `yaml:",omitempty"`
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
	//	  * scm
	//
	//  default: true
	WorkingBranch *bool `yaml:",omitempty"`
}

// Bitbucket contains information to interact with Bitbucket Cloud API
type Bitbucket struct {
	// Spec contains inputs coming from updatecli configuration
	Spec Spec
	// client handle the api authentication
	client           *scm.Client
	pipelineID       string
	nativeGitHandler gitgeneric.GitHandler
	workingBranch    bool
	force            bool
}

// New returns a new valid Bitbucket Cloud object.
func New(spec interface{}, pipelineID string) (*Bitbucket, error) {
	var s Spec
	var clientSpec client.Spec

	// mapstructure.Decode cannot handle embedded fields
	// hence we decode it in two steps
	err := mapstructure.Decode(spec, &clientSpec)
	if err != nil {
		return &Bitbucket{}, err
	}

	err = mapstructure.Decode(spec, &s)
	if err != nil {
		return &Bitbucket{}, nil
	}

	s.Spec = clientSpec

	err = s.Validate()
	if err != nil {
		return &Bitbucket{}, err
	}

	if s.Directory == "" {
		s.Directory = path.Join(tmp.Directory, "bitbucket", s.Owner, s.Repository)
	}

	if len(s.Branch) == 0 {
		logrus.Warningf("no git branch specified, fallback to %q", "main")
		s.Branch = "main"
	}

	workingBranch := true
	if s.WorkingBranch != nil {
		workingBranch = *s.WorkingBranch
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
		return &Bitbucket{}, err
	}

	nativeGitHandler := gitgeneric.GoGit{}
	g := Bitbucket{
		Spec:             s,
		client:           c,
		pipelineID:       pipelineID,
		nativeGitHandler: &nativeGitHandler,
		workingBranch:    workingBranch,
		force:            force,
	}

	g.setDirectory()

	return &g, nil
}

// Retrieve git tags from a remote Bitbucket Server repository
func (b *Bitbucket) SearchTags() (tags []string, err error) {
	// Timeout api query after 30sec
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	references, resp, err := b.client.Git.ListTags(
		ctx,
		strings.Join([]string{b.Spec.Owner, b.Spec.Repository}, "/"),
		scm.ListOptions{
			URL:  client.URL(),
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

	if len(missingParameters) > 0 {
		logrus.Errorf("missing parameter(s) [%s]", strings.Join(missingParameters, ","))
	}

	if gotError {
		return fmt.Errorf("wrong Bitbucket Cloud configuration")
	}

	return nil
}
