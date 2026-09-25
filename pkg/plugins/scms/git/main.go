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

/*
"git" defines the specification for a git repository used as an scm.
Updatecli clones the repository, reads files from it, and commits and pushes the changes made by targets.
*/
type Spec struct {
	// "url" defines the URL of the git repository.
	//
	// remark:
	//   * with the ssh protocol, the local ssh configuration must allow cloning the repository.
	//
	// example:
	//   * url: git@github.com:updatecli/updatecli.git
	//   * url: https://github.com/updatecli/updatecli.git
	//
	URL string `yaml:",omitempty" jsonschema:"required"`
	// "username" defines the username used to authenticate over HTTP.
	//
	Username string `yaml:",omitempty"`
	// "password" defines the password used to authenticate over HTTP.
	//
	// remark:
	//   * a password is sensitive, so avoid writing it in the manifest.
	//     Read it from an environment variable, for example with `{{ requiredEnv "GIT_PASSWORD" }}`.
	//
	Password string `yaml:",omitempty"`
	// "branch" defines the git branch to work on.
	//
	// default:
	//   main
	//
	// remark:
	//   * when the git scm is used by a source or a condition, files are read from this branch.
	//   * when the git scm is used by a target, Updatecli pushes changes to this branch,
	//     or to a working branch based on it when "workingbranch" is true.
	//   * see https://github.com/updatecli/updatecli/issues/1139
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
	// "user" defines the name used to author commits.
	//
	// default:
	//   updatecli-bot
	//
	User string `yaml:",omitempty"`
	// "email" defines the email address used to author commits.
	//
	// default:
	//   updatecli-bot@updatecli.io
	//
	Email string `yaml:",omitempty"`
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
	// "directory" defines the local path where the git repository is cloned.
	//
	// default:
	//   a directory under the Updatecli temporary directory, such as
	//   "/tmp/updatecli/<sanitized url>" on Linux.
	//
	// remark:
	//   * keep the default value unless you have a good reason to change it,
	//     as Updatecli may delete the directory after a pipeline run.
	//
	Directory string `yaml:",omitempty"`
	// "force" defines whether Updatecli runs `git push --force` when pushing changes.
	//
	// default:
	//   false
	//
	// remark:
	//   * when true, Updatecli also recreates the working branches that diverged from their base branch.
	//   * when "force" is true and "workingbranch" is not set, the git scm returns an error,
	//     to avoid force pushing to "branch" by mistake. Set "workingbranch" explicitly to confirm the behavior.
	//
	Force bool `yaml:",omitempty"`
	// "commitmessage" defines the settings used to generate commit messages.
	//
	// remark:
	//   * the settings apply to every target using this scm.
	//
	CommitMessage commit.Commit `yaml:",omitempty"`
	// "gpg" defines the GPG key and passphrase used to sign commits.
	//
	GPG sign.GPGSpec `yaml:",omitempty"`
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
	//   false
	//
	WorkingBranch *bool `yaml:",omitempty"`
}

// Git contains the git scm handler
type Git struct {
	// spec contains the git scm specification
	spec Spec
	// nativeGitHandler is the native git handler
	nativeGitHandler gitgeneric.GitHandler
	// workingBranch is used to create a temporary branch to work on.
	workingBranch          bool
	workingBranchPrefix    string
	workingBranchSeparator string
	pipelineID             string
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

	workingBranchPrefix := "updatecli"
	if s.WorkingBranchPrefix != nil {
		workingBranchPrefix = *s.WorkingBranchPrefix
	}

	workingBranchSeparator := "_"
	if s.WorkingBranchSeparator != nil {
		workingBranchSeparator = *s.WorkingBranchSeparator
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

	if s.Email == "" {
		s.Email = gitgeneric.DefaultGitCommitEmailAddress
	}

	if s.User == "" {
		s.User = gitgeneric.DefaultGitCommitUserName
	}

	return &Git{
		spec:                   s,
		nativeGitHandler:       &nativeGitHandler,
		workingBranch:          workingBranch,
		workingBranchPrefix:    workingBranchPrefix,
		workingBranchSeparator: workingBranchSeparator,
		pipelineID:             pipelineID,
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
