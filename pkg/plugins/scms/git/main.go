package git

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/tmp"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/commit"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// Spec contains settings to manipulate a git repository.
type Spec struct {
	//	"url" specifies the git url to work on.
	//
	//	compatible:
	//	  * scm
	//
	//	example:
	//	  * git@github.com:updatecli/updatecli.git
	//	  * https://github.com/updatecli/updatecli.git
	//
	//	remarks:
	//		when using the ssh protocol, the user must have the right to clone the repository
	//		based on its local ssh configuration
	URL string `yaml:",omitempty" jsonschema:"required"`
	//	"username" specifies the username when using the HTTP protocol
	//
	//	compatible
	//	  * scm
	Username string `yaml:",omitempty"`
	//	"password" specifies the password when using the HTTP protocol
	//
	//	compatible:
	//	  * scm
	Password string `yaml:",omitempty"`
	// 	"branch" defines the git branch to work on.
	//
	// 	compatible:
	// 	  * scm
	//
	// 	default:
	// 		main
	//
	// 	remark:
	// 		depending on which resource references the GitHub scm, the behavior will be different.
	//
	// 		If the scm is linked to a source or a condition (using scmid), the branch will be used to retrieve
	// 		file(s) from that branch.
	//
	// 		If the scm is linked to target then Updatecli will push any changes to that branch
	//
	// 		For more information, please refer to the following issue:
	// 		https://github.com/updatecli/updatecli/issues/1139
	Branch string `yaml:",omitempty"`
	//	"user" specifies the user associated with new git commit messages created by Updatecli
	//
	//	compatible:
	//	  * scm
	User string `yaml:",omitempty"`
	//	"email" defines the email used to commit changes.
	//
	//	compatible:
	//	  * scm
	//
	//	default:
	//		default set to your global git configuration
	Email string `yaml:",omitempty"`
	//	"directory" defines the local path where the git repository is cloned.
	//
	//	compatible:
	//	  * scm
	//
	//	remark:
	//	  Unless you know what you are doing, it is highly recommended to use the default value.
	//	  The reason is that Updatecli may automatically clean up the directory after a pipeline execution.
	//
	//	default:
	// 	  The default value is based on your local temporary directory like /tmp/updatecli/<url> on Linux
	Directory string `yaml:",omitempty"`
	//	"force" is used during the git push phase to run `git push --force`.
	//
	//	compatible:
	//	  * scm
	//
	//  default:
	//	  false
	//
	//  remark:
	//    When force is set to true, Updatecli also recreate the working branches that
	//    diverged from their base branch.
	Force bool `yaml:",omitempty"`
	//	"commitMessage" is used to generate the final commit message.
	//
	//	compatible:
	//	  * scm
	//
	//	remark:
	//	  it's worth mentioning that the commit message is applied to all targets linked to the same scm.
	//
	//	default:
	//	  false
	CommitMessage commit.Commit `yaml:",omitempty"`
	//	"gpg" specifies the GPG key and passphrased used for commit signing
	//
	//	compatible:
	//	  * scm
	GPG sign.GPGSpec `yaml:",omitempty"`
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
	//  default: false
	WorkingBranch *bool `yaml:",omitempty"`
}

// Git contains the git scm handler
type Git struct {
	// spec contains the git scm specification
	spec Spec
	// nativeGitHandler is the native git handler
	nativeGitHandler gitgeneric.GitHandler
	// workingBranch is used to create a temporary branch to work on.
	workingBranch bool
	pipelineID    string
}

// New returns a new git object
func New(s Spec, pipelineID string) (*Git, error) {
	var err error
	if len(s.Directory) == 0 {
		s.Directory, err = newDirectory(s.URL)
		if err != nil {
			return nil, err
		}
	}

	if len(s.Branch) == 0 {
		s.Branch = "main"
	}

	var workingBranch bool
	switch s.WorkingBranch {
	case nil:
		workingBranch = false

		if s.Force {
			errorMsg := fmt.Sprintf(`
Better safe than sorry.

The scm force option set to true means that Updatecli is going to run "git push --force"
Some target plugin, like the shell one, run "git commit -A" to catch all changes done by that target.
Because the Git scm plugin has by default the workingBranch option set to false,
Updatecli may be pushing unwanted changes to the branch %q.

If you know what you are doing, please set the workingBranch option to false in your configuration file to ignore this error message.
`, s.Branch)

			logrus.Errorln(errorMsg)
			return nil, errors.New("unclear configuration, better safe than sorry")
		}
	default:
		workingBranch = *s.WorkingBranch
	}

	nativeGitHandler := gitgeneric.GoGit{}

	return &Git{
		spec:             s,
		nativeGitHandler: &nativeGitHandler,
		workingBranch:    workingBranch,
		pipelineID:       pipelineID,
	}, nil
}

// Merge returns nil if it successfully merges the child Spec into target receiver.
// Please note that child attributes always overrides receiver's
func (gs *Spec) Merge(child interface{}) error {
	childGHSpec, ok := child.(Spec)
	if !ok {
		return fmt.Errorf("unable to merge GitHub spec with unknown object type")
	}

	if childGHSpec.Branch != "" {
		gs.Branch = childGHSpec.Branch
	}
	if childGHSpec.CommitMessage != (commit.Commit{}) {
		gs.CommitMessage = childGHSpec.CommitMessage
	}
	if childGHSpec.Directory != "" {
		gs.Directory = childGHSpec.Directory
	}
	if childGHSpec.Email != "" {
		gs.Email = childGHSpec.Email
	}
	if childGHSpec.Force {
		gs.Force = childGHSpec.Force
	}
	if childGHSpec.GPG != (sign.GPGSpec{}) {
		gs.GPG = childGHSpec.GPG
	}
	if childGHSpec.URL != "" {
		gs.URL = childGHSpec.URL
	}
	if childGHSpec.User != "" {
		gs.User = childGHSpec.User
	}
	if childGHSpec.Username != "" {
		gs.Username = childGHSpec.Username
	}
	if childGHSpec.Submodules != nil {
		gs.Submodules = childGHSpec.Submodules
	}

	return nil
}

// MergeFromEnv updates the target receiver with the "non zero-ed" environment variables
func (gs *Spec) MergeFromEnv(envPrefix string) {
	prefix := fmt.Sprintf("%s_", envPrefix)
	if os.Getenv(fmt.Sprintf("%s%s", prefix, "BRANCH")) != "" {
		gs.Branch = os.Getenv(fmt.Sprintf("%s%s", prefix, "BRANCH"))
	}
	if os.Getenv(fmt.Sprintf("%s%s", prefix, "DIRECTORY")) != "" {
		gs.Directory = os.Getenv(fmt.Sprintf("%s%s", prefix, "DIRECTORY"))
	}
	if os.Getenv(fmt.Sprintf("%s%s", prefix, "EMAIL")) != "" {
		gs.Email = os.Getenv(fmt.Sprintf("%s%s", prefix, "EMAIL"))
	}
	if os.Getenv(fmt.Sprintf("%s%s", prefix, "URL")) != "" {
		gs.URL = os.Getenv(fmt.Sprintf("%s%s", prefix, "URL"))
	}
	if os.Getenv(fmt.Sprintf("%s%s", prefix, "USERNAME")) != "" {
		gs.Username = os.Getenv(fmt.Sprintf("%s%s", prefix, "USERNAME"))
	}
	if os.Getenv(fmt.Sprintf("%s%s", prefix, "USER")) != "" {
		gs.User = os.Getenv(fmt.Sprintf("%s%s", prefix, "USER"))
	}
}

func newDirectory(URL string) (string, error) {

	directory := path.Join(
		tmp.Directory,
		sanitizeDirectoryName(URL))

	if _, err := os.Stat(directory); os.IsNotExist(err) {

		err = os.MkdirAll(directory, 0755)
		if err != nil {
			logrus.Errorf("err - %s", err)
			return "", err
		}
	}

	return directory, nil

}

// sanitizeDirectoryName ensures that we don't have unwanted information
// in the git directory name like username/password or special characters
func sanitizeDirectoryName(URL string) string {

	gitProtocols := []string{"https://", "ssh://", "http://", "file://"}

	forbiddenCharacters := []string{
		"@", "~", "%", "$", "*", " ",
		"+", "?", "\"", "<", ">", "|",
	}

	// Trim git protocols
	for _, str := range gitProtocols {
		URL = strings.TrimPrefix(URL, str)
	}

	for _, str := range forbiddenCharacters {
		if strings.Contains(URL, str) {
			URL = strings.ReplaceAll(URL, str, "")
		}
	}

	for _, str := range []string{"/", "\\", ".", ":"} {
		if strings.Contains(URL, str) {
			URL = strings.ReplaceAll(URL, str, "_")
		}
	}
	return URL
}
