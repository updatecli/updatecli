package azuredevops

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/tmp"
	azdoclient "github.com/updatecli/updatecli/pkg/plugins/resources/azuredevops/client"
	"github.com/updatecli/updatecli/pkg/plugins/resources/azuredevops/credential"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/commit"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
)

const (
	// Kind defines the SCM kind for Azure DevOps.
	Kind = "azuredevops"
)

/*
"azuredevops" defines the specification for an Azure DevOps git repository used as an scm.
Updatecli clones the repository, reads files from it, and commits and pushes the changes made by targets.
*/
type Spec struct {
	azdoclient.Spec `yaml:",inline,omitempty"`
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
	//   "/tmp/updatecli/azuredevops/<project>/<repository>" on Linux.
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
	//   * when "workingbranch" is false and "force" is not set, the Azure DevOps scm returns an error,
	//     to avoid force pushing to "branch" by mistake. Set "force" explicitly to confirm the behavior.
	//
	Force *bool `yaml:",omitempty"`
	// "gpg" defines the GPG key and passphrase used to sign commits.
	//
	GPG sign.GPGSpec `yaml:",omitempty"`
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
	//   * when the Azure DevOps scm is used by a source or a condition, files are read from this branch.
	//   * when the Azure DevOps scm is used by a target, Updatecli pushes changes to a working branch
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

// AzureDevOps contains settings to interact with Azure DevOps.
type AzureDevOps struct {
	force                  bool
	Spec                   Spec
	client                 azdoclient.Client
	pipelineID             string
	nativeGitHandler       gitgeneric.GitHandler
	workingBranch          bool
	workingBranchPrefix    string
	workingBranchSeparator string
}

// New returns a new valid Azure DevOps object.
func New(spec interface{}, pipelineID string) (*AzureDevOps, error) {
	var s Spec
	var clientSpec azdoclient.Spec

	// mapstructure.Decode cannot handle embedded fields
	// hence we decode it in two steps
	err := mapstructure.Decode(spec, &clientSpec)
	if err != nil {
		return &AzureDevOps{}, err
	}

	err = mapstructure.Decode(spec, &s)
	if err != nil {
		return &AzureDevOps{}, err
	}

	if clientSpec.Organization == "" {
		clientSpec.Organization = s.Organization
	}

	if clientSpec.URL == "" {
		clientSpec.URL = s.URL
	}

	if clientSpec.Project == "" {
		clientSpec.Project = s.Project
	}

	if clientSpec.Repository == "" {
		clientSpec.Repository = s.Repository
	}

	usernameFromEnv, tokenFromEnv := credential.GetCredentialsFromEnv()
	if clientSpec.Username == "" {
		clientSpec.Username = s.Username

		if usernameFromEnv != "" {
			logrus.Debugf("using Azure DevOps username from environment variable %s", credential.ENVIRONMENT_AZURE_DEVOPS_USERNAME)
			clientSpec.Username = usernameFromEnv
		}
	}

	if clientSpec.Token == "" {
		clientSpec.Token = s.Token
		if tokenFromEnv != "" {
			logrus.Debugf("using Azure DevOps token from environment variable %s", credential.ENVIRONMENT_AZURE_DEVOPS_TOKEN)
			clientSpec.Token = tokenFromEnv
		}
	}

	err = clientSpec.Sanitize()
	if err != nil {
		return &AzureDevOps{}, err
	}

	s.Spec = clientSpec

	err = s.Validate()
	if err != nil {
		return &AzureDevOps{}, err
	}

	if s.Directory == "" {
		s.Directory = path.Join(tmp.Directory, Kind, s.Project, s.Repository)
	}

	if s.Branch == "" {
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

	if force && !workingBranch && s.Force == nil {
		errorMsg := fmt.Sprintf(`
Better safe than sorry.

Updatecli may be pushing unwanted changes to the branch %q.

The Azure DevOps scm plugin has by default the force option set to true,
The scm force option set to true means that Updatecli is going to run "git push --force"
Some target plugin, like the shell one, run "git commit -A" to catch all changes done by that target.

If you know what you are doing, please set the force option to true in your configuration file to ignore this error message.
`, s.Branch)

		logrus.Errorln(errorMsg)
		return nil, errors.New("unclear configuration, better safe than sorry")
	}

	if s.Email == "" {
		s.Email = gitgeneric.DefaultGitCommitEmailAddress
	}

	if s.User == "" {
		s.User = gitgeneric.DefaultGitCommitUserName
	}

	if s.Username == "" && s.Token != "" {
		s.Username = gitgeneric.DefaultGitCommitUserName
	}

	c, err := azdoclient.New(clientSpec)
	if err != nil {
		return &AzureDevOps{}, err
	}

	nativeGitHandler := gitgeneric.GoGit{}

	azdo := AzureDevOps{
		force:                  force,
		Spec:                   s,
		client:                 c,
		pipelineID:             pipelineID,
		nativeGitHandler:       &nativeGitHandler,
		workingBranch:          workingBranch,
		workingBranchPrefix:    workingBranchPrefix,
		workingBranchSeparator: workingBranchSeparator,
	}

	return &azdo, nil
}

func (s *Spec) Validate() error {
	return s.Spec.Validate()
}

func (a *AzureDevOps) repositoryURL() string {
	return azdoclient.GitURL(a.Spec.URL, a.Spec.Organization, a.Spec.Project, a.Spec.Repository)
}

// GetBranches returns the source, working and target branches.
func (a *AzureDevOps) GetBranches() (sourceBranch, workingBranch, targetBranch string) {
	sourceBranch = a.Spec.Branch
	workingBranch = a.Spec.Branch
	targetBranch = a.Spec.Branch

	if len(a.pipelineID) > 0 && a.workingBranch {
		workingBranch = a.nativeGitHandler.SanitizeBranchName(
			strings.Join([]string{a.workingBranchPrefix, targetBranch, a.pipelineID}, a.workingBranchSeparator))
	}

	return sourceBranch, workingBranch, targetBranch
}

// Clone runs `git clone`.
func (a *AzureDevOps) Clone() (string, error) {

	// This global is shared across all SCMs/pipelines in the same process,
	// Clone() mutates the global go-git setting transport.UnsupportedCapabilities and then need to be restored.
	// otherwise it can cause surprising behavior (and data races) when multiple pipelines run or
	// when other SCMs clone/fetch after an Azure DevOps clone.
	currentUnsupportedCapabilities := transport.UnsupportedCapabilities
	defer func() {
		transport.UnsupportedCapabilities = currentUnsupportedCapabilities
	}()

	transport.UnsupportedCapabilities = []capability.Capability{
		capability.ThinPack,
	}

	// Source:
	//   * https://github.com/go-git/go-git/blob/master/_examples/azure_devops/main.go
	// Azure DevOps requires capabilities multi_ack / multi_ack_detailed,
	// which are not fully implemented and by default are included in
	// transport.UnsupportedCapabilities.
	//
	// The initial clone operations require a full download of the repository,
	// and therefore those unsupported capabilities are not as crucial, so
	// by removing them from that list allows for the first clone to work
	// successfully.
	//
	// Additional fetches will yield issues, therefore work always from a clean
	// clone until those capabilities are fully supported.
	//
	// New commits and pushes against a remote worked without any issues.
	transport.UnsupportedCapabilities = []capability.Capability{
		capability.ThinPack,
	}

	err := a.nativeGitHandler.Clone(
		a.Spec.Username,
		a.Spec.Token,
		a.GetURL(),
		a.GetDirectory(),
		a.Spec.Submodules,
		a.Spec.Depth,
		a.Spec.Branch,
		a.Spec.SingleBranch != nil && *a.Spec.SingleBranch,
	)
	if err != nil {
		logrus.Errorf("failed cloning Azure DevOps repository %q", redact.URL(a.GetURL()))
		return "", err
	}

	return a.Spec.Directory, nil
}

// Checkout creates and then uses a temporary git branch.
func (a *AzureDevOps) Checkout() error {
	sourceBranch, workingBranch, _ := a.GetBranches()

	return a.nativeGitHandler.Checkout(
		a.Spec.Username,
		a.Spec.Token,
		sourceBranch,
		workingBranch,
		a.Spec.Directory,
		a.force,
		a.Spec.Depth,
	)
}

// Commit runs `git commit`.
func (a *AzureDevOps) Commit(ctx context.Context, message string) error {
	commitMessage, err := a.Spec.CommitMessage.Generate(message)
	if err != nil {
		return err
	}

	err = a.nativeGitHandler.Commit(
		a.Spec.User,
		a.Spec.Email,
		commitMessage,
		a.GetDirectory(),
		a.Spec.GPG.SigningKey,
		a.Spec.GPG.Passphrase,
	)
	if err != nil {
		return err
	}

	if a.Spec.CommitMessage.IsSquash() {
		sourceBranch, workingBranch, _ := a.GetBranches()
		if err = a.nativeGitHandler.SquashCommit(a.GetDirectory(), sourceBranch, workingBranch, gitgeneric.SquashCommitOptions{
			IncludeCommitTitles: true,
			Message:             commitMessage,
			SigninKey:           a.Spec.GPG.SigningKey,
			SigninPassphrase:    a.Spec.GPG.Passphrase,
		}); err != nil {
			return err
		}
	}

	return nil
}

// Add runs `git add`.
func (a *AzureDevOps) Add(files []string) error {
	return a.nativeGitHandler.Add(files, a.Spec.Directory)
}

// CleanWorkingBranch checks if the working branch is diverged from the target branch and removes it if not.
func (a *AzureDevOps) CleanWorkingBranch() (bool, error) {
	_, workingBranch, targetBranch := a.GetBranches()

	if workingBranch == targetBranch {
		logrus.Infof("Skipping cleaning working branch %q on %q (same as target branch)\n", workingBranch, redact.URL(a.GetURL()))
		return false, nil
	}

	isSimilarBranch, err := a.nativeGitHandler.IsSimilarBranch(workingBranch, targetBranch, a.GetDirectory())
	if err != nil {
		return false, fmt.Errorf("failed to compare working branch %q with target branch %q: %w", workingBranch, targetBranch, err)
	}

	if isSimilarBranch {
		if err = a.nativeGitHandler.DeleteBranch(workingBranch, a.GetDirectory(), a.Spec.Username, a.Spec.Token); err != nil {
			return false, fmt.Errorf("failed to delete working branch %q from %q: %w", workingBranch, a.GetDirectory(), err)
		}
		return true, nil
	}

	return false, nil
}

// GetDirectory returns the local git repository path.
func (a *AzureDevOps) GetDirectory() (directory string) {
	return a.Spec.Directory
}

// GetURL returns an Azure DevOps git URL.
func (a *AzureDevOps) GetURL() string {
	return a.repositoryURL()
}

// IsRemoteBranchUpToDate checks if the local working branch is up to date with the remote branch.
func (a *AzureDevOps) IsRemoteBranchUpToDate() (bool, error) {
	sourceBranch, workingBranch, _ := a.GetBranches()

	return a.nativeGitHandler.IsLocalBranchSyncedWithRemote(
		sourceBranch,
		workingBranch,
		a.Spec.Username,
		a.Spec.Token,
		a.GetDirectory(),
	)
}

// Clean deletes the local working directory.
func (a *AzureDevOps) Clean() error {
	return os.RemoveAll(a.Spec.Directory)
}

// IsRemoteWorkingBranchExist checks if the remote working branch exists.
func (a *AzureDevOps) IsRemoteWorkingBranchExist() (bool, error) {
	_, workingBranch, _ := a.GetBranches()

	return a.nativeGitHandler.IsRemoteBranchExist(
		workingBranch,
		a.Spec.Username,
		a.Spec.Token,
		a.GetDirectory(),
	)
}

// Push runs `git push`.
func (a *AzureDevOps) Push() (bool, error) {
	return a.nativeGitHandler.Push(
		a.Spec.Username,
		a.Spec.Token,
		a.GetDirectory(),
		a.force,
	)
}

// PushTag pushes tags.
func (a *AzureDevOps) PushTag(tag string) error {
	return a.nativeGitHandler.PushTag(
		tag,
		a.Spec.Username,
		a.Spec.Token,
		a.GetDirectory(),
		a.force,
	)
}

// PushBranch pushes branches.
func (a *AzureDevOps) PushBranch(branch string) error {
	return a.nativeGitHandler.PushBranch(
		branch,
		a.Spec.Username,
		a.Spec.Token,
		a.GetDirectory(),
		a.force,
	)
}

// GetChangedFiles returns a list of changed files.
func (a *AzureDevOps) GetChangedFiles(workingDir string) ([]string, error) {
	return a.nativeGitHandler.GetChangedFiles(workingDir)
}
