package github

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/shurcooL/githubv4"

	"github.com/updatecli/updatecli/pkg/core/tmp"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/commit"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/app"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/client"

	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

const (
	Kind = "github"

	// Keys shared by the GraphQL query variables and the spec validation.
	keyOwner      = "owner"
	keyName       = "name"
	keyRepository = "repository"
	keyBefore     = "before"
)

/*
"github" defines the specification for a GitHub repository used as an scm.
Updatecli clones the repository, reads files from it, and commits and pushes the changes made by targets.
*/
type Spec struct {
	// "branch" defines the git branch to work on.
	//
	// default:
	//   main
	//
	// remark:
	//   * when the GitHub scm is used by a source or a condition, files are read from this branch.
	//   * when the GitHub scm is used by a target, Updatecli pushes changes to a working branch
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
	// "directory" defines the local path where the git repository is cloned.
	//
	// default:
	//   a directory under the Updatecli temporary directory, such as
	//   "/tmp/updatecli/github/<owner>/<repository>" on Linux.
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
	// "owner" defines the owner of the repository.
	//
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// "repository" defines the name of the repository.
	//
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// "token" defines the token used to authenticate with the GitHub API.
	//
	// remark:
	//   * "token" and "app" are mutually exclusive.
	//   * a token is sensitive, so avoid writing it in the manifest. Read it from an environment
	//     variable with `{{ requiredEnv "GITHUB_TOKEN" }}`, or from a SOPS file with `{{ .github.token }}`.
	//     See https://github.com/getsops/sops
	//   * the environment variable UPDATECLI_GITHUB_TOKEN, or the UPDATECLI_GITHUB_APP_* environment
	//     variables, take precedence over this value.
	//   * when no credential is set, Updatecli falls back to the environment variable GITHUB_TOKEN.
	//
	Token string `yaml:",omitempty"`
	// "url" defines the GitHub URL, to use a GitHub Enterprise instance.
	//
	// default:
	//   github.com
	//
	// remark:
	//   * the scheme "https://" is added when missing.
	//
	// example:
	//   * url: github.example.com
	//
	URL string `yaml:",omitempty"`
	// "username" defines the username used with the token to authenticate with the GitHub API.
	//
	// remark:
	//   * the token is usually enough on its own. A username may be needed for private repositories.
	//
	Username string `yaml:",omitempty"`
	// "user" defines the name used to author commits.
	//
	// default:
	//   updatecli-bot
	//
	User string `yaml:",omitempty"`
	// "gpg" defines the GPG key and passphrase used to sign commits.
	//
	GPG sign.GPGSpec `yaml:",omitempty"`
	// "force" defines whether Updatecli runs `git push --force` when pushing changes.
	//
	// default:
	//   true
	//
	// remark:
	//   * when true, Updatecli also recreates the working branches that diverged from their base branch.
	//   * when "workingbranch" is false and "force" is not set, the GitHub scm returns an error,
	//     to avoid force pushing to "branch" by mistake. Set "force" explicitly to confirm the behaviour.
	//
	Force *bool `yaml:",omitempty"`
	// "commitmessage" defines the settings used to generate commit messages.
	//
	// remark:
	//   * the settings apply to every target using this scm.
	//
	CommitMessage commit.Commit `yaml:",omitempty"`
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
	// "commitusingapi" defines whether Updatecli creates commits with the GitHub GraphQL API instead of git.
	//
	// default:
	//   false
	//
	// remark:
	//   * GitHub signs the commits created this way from a GitHub Actions workflow using the GITHUB_TOKEN.
	//     See https://github.com/updatecli/updatecli/issues/1914
	//
	CommitUsingAPI *bool `yaml:",omitempty"`
	// "app" defines the GitHub App credentials used to authenticate with the GitHub API.
	//
	// remark:
	//   * "app" and "token" are mutually exclusive, and "username" is ignored when "app" is set.
	//   * a GitHub App gives better security and finer permissions than a personal token.
	//   * see https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/authenticating-as-a-github-app-installation
	//
	App *app.Spec `yaml:",omitempty"`
}

// GitHub contains settings to interact with GitHub
type Github struct {
	force bool
	// Spec contains inputs coming from updatecli configuration
	Spec                   Spec
	pipelineID             string
	client                 client.Client
	nativeGitHandler       gitgeneric.GitHandler
	workingBranch          bool
	workingBranchPrefix    string
	workingBranchSeparator string
	commitUsingApi         bool
	token                  oauth2.TokenSource
	username               string
	URL                    string
}

// Repository contains GitHub repository data
type Repository struct {
	ID          string
	Name        string
	Owner       string
	ParentID    string
	ParentName  string
	ParentOwner string
	Status      string
}

type RepositoryRef struct {
	ID               string
	HeadOid          string
	DefaultBranchOid string
}

// New returns a new valid GitHub object.
func New(s Spec, pipelineID string) (*Github, error) {
	var err error

	errs := s.Validate()

	if len(errs) > 0 {
		strErrs := []string{}
		for _, err := range errs {
			strErrs = append(strErrs, err.Error())
		}
		return &Github{}, fmt.Errorf("%s", strings.Join(strErrs, "\n"))
	}

	if s.Directory == "" {
		s.Directory = path.Join(tmp.Directory, "github", s.Owner, s.Repository)
	}

	nativeGitHandler := gitgeneric.GoGit{}

	// By default, we create a working branch but if for some reason we don't want to create it
	// Then we also need to update the force safeguard to avoid force pushing on the main branch.
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

	commitUsingApi := false
	if s.CommitUsingAPI != nil {
		commitUsingApi = *s.CommitUsingAPI
	}

	if force {
		if !workingBranch && s.Force == nil {
			errorMsg := fmt.Sprintf(`
Better safe than sorry.

Updatecli may be pushing unwanted changes to the branch %q.

The GitHub scm plugin has by default the force option set to true,
The scm force option set to true means that Updatecli is going to run "git push --force"
Some target plugin, like the shell one, run "git commit -A" to catch all changes done by that target.

If you know what you are doing, please set the force option to true in your configuration file to ignore this error message.
`, s.Branch)

			logrus.Errorln(errorMsg)
			return nil, errors.New("unclear configuration, better safe than sorry")

		}
	}

	if s.Email == "" {
		s.Email = gitgeneric.DefaultGitCommitEmailAddress
	}

	if s.User == "" {
		s.User = gitgeneric.DefaultGitCommitUserName
	}

	g := Github{
		force:                  force,
		Spec:                   s,
		pipelineID:             pipelineID,
		nativeGitHandler:       &nativeGitHandler,
		workingBranch:          workingBranch,
		workingBranchPrefix:    workingBranchPrefix,
		workingBranchSeparator: workingBranchSeparator,
		commitUsingApi:         commitUsingApi,
	}

	clientConfig, err := client.New(g.Spec.Username, g.Spec.Token, g.Spec.App, g.Spec.URL)
	if err != nil {
		return &Github{}, fmt.Errorf("creating GitHub client: %w", err)
	}

	if clientConfig != nil {
		g.client = clientConfig.Client
		g.username = clientConfig.Username
		g.token = clientConfig.TokenSource
		g.URL = clientConfig.URL
	}

	g.setDirectory()

	return &g, nil
}

// Validate verifies if mandatory GitHub parameters are provided and return false if not.
func (s *Spec) Validate() (errs []error) {
	required := []string{}

	if s.App != nil && len(s.Token) > 0 {
		errs = append(errs, fmt.Errorf("you cannot use both token and app authentication methods"))
	} else if s.App != nil {
		if err := s.App.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("app configuration is invalid: %w", err))
		}
	}

	if len(s.Owner) == 0 {
		required = append(required, keyOwner)
	}

	if len(s.Repository) == 0 {
		required = append(required, keyRepository)
	}

	if len(required) > 0 {
		errs = append(errs, fmt.Errorf("github parameter(s) required: [%v]", strings.Join(required, ",")))
	}

	return errs
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
	if childGHSpec.Force != nil {
		gs.Force = childGHSpec.Force
	}
	if childGHSpec.GPG != (sign.GPGSpec{}) {
		gs.GPG = childGHSpec.GPG
	}
	if childGHSpec.Owner != "" {
		gs.Owner = childGHSpec.Owner
	}
	// PullRequest is deprecated so not merging it
	if childGHSpec.Repository != "" {
		gs.Repository = childGHSpec.Repository
	}
	if childGHSpec.Token != "" {
		gs.Token = childGHSpec.Token
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

	if childGHSpec.App != nil {
		gs.App = &app.Spec{
			ClientID:       childGHSpec.App.ClientID,
			PrivateKey:     childGHSpec.App.PrivateKey,
			PrivateKeyPath: childGHSpec.App.PrivateKeyPath,
			InstallationID: childGHSpec.App.InstallationID,
			ExpirationTime: childGHSpec.App.ExpirationTime,
		}
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
	if os.Getenv(fmt.Sprintf("%s%s", prefix, "OWNER")) != "" {
		gs.Owner = os.Getenv(fmt.Sprintf("%s%s", prefix, "OWNER"))
	}
	if os.Getenv(fmt.Sprintf("%s%s", prefix, "REPOSITORY")) != "" {
		gs.Repository = os.Getenv(fmt.Sprintf("%s%s", prefix, "REPOSITORY"))
	}
	if os.Getenv(fmt.Sprintf("%s%s", prefix, "TOKEN")) != "" {
		gs.Token = os.Getenv(fmt.Sprintf("%s%s", prefix, "TOKEN"))
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

func (g *Github) setDirectory() {
	if _, err := os.Stat(g.Spec.Directory); os.IsNotExist(err) {

		err := os.MkdirAll(g.Spec.Directory, 0755)
		if err != nil {
			logrus.Errorf("err - %s", err)
		}
	}
}

func (g *Github) queryRepository(ctx context.Context, sourceBranch string, workingBranch string, retry int) (*Repository, error) {
	rateLimit, err := queryRateLimit(g.client, ctx)
	logrus.Debugln(rateLimit)
	if err != nil {
		if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) {
			if retry < client.MaxRetry {
				logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, client.MaxRetry)
				rateLimit.Pause()
				return g.queryRepository(ctx, sourceBranch, workingBranch, retry+1)
			}
			return nil, errors.New(ErrAPIRateLimitExceededFinalAttempt)
		}
		return nil, fmt.Errorf("unable to query GitHub API rate limit: %w", err)
	}

	/*
			   query($owner: String!, $name: String!) {
			       repository(owner: $owner, name: $name){
			           id
			           name
		               owner {
		                   login
		               }
			           parent {
		                   id
		                   name
		                   owner {
		                       login
		                   }
			           }
			       }
			   }
	*/

	var query struct {
		Repository struct {
			ID    string
			Name  string
			Owner struct {
				Login string
			}

			Ref *struct {
				Name    string
				Compare struct {
					Status string
				} `graphql:"compare(headRef: $headRef)"`
			} `graphql:"ref(qualifiedName: $qualifiedName)"`

			Parent *struct {
				ID    string
				Name  string
				Owner struct {
					Login string
				}
			}
		} `graphql:"repository(owner: $owner, name: $name)"`
		RateLimit RateLimit
	}

	variables := map[string]interface{}{
		keyOwner:        githubv4.String(g.Spec.Owner),
		keyName:         githubv4.String(g.Spec.Repository),
		"qualifiedName": githubv4.String(sourceBranch),
		"headRef":       githubv4.String(workingBranch),
	}

	err = g.client.Query(ctx, &query, variables)
	if err != nil {
		if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) {
			if retry < client.MaxRetry {
				logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, client.MaxRetry)
				query.RateLimit.Pause()
				return g.queryRepository(ctx, sourceBranch, workingBranch, retry+1)
			}
			return nil, errors.New(ErrAPIRateLimitExceededFinalAttempt)
		}
		return nil, err
	}

	parentID := ""
	parentName := ""
	parentOwner := ""
	if query.Repository.Parent != nil {
		parentID = query.Repository.Parent.ID
		parentName = query.Repository.Parent.Name
		parentOwner = query.Repository.Parent.Owner.Login
	}

	status := ""
	if query.Repository.Ref != nil {
		status = query.Repository.Ref.Compare.Status
	}

	result := &Repository{
		ID:          query.Repository.ID,
		Name:        query.Repository.Name,
		Owner:       query.Repository.Owner.Login,
		ParentID:    parentID,
		ParentName:  parentName,
		ParentOwner: parentOwner,
		Status:      status,
	}

	return result, nil
}

// Returns Git object ID of the latest commit on the branch and the default branch
// of the repository.
func (g *Github) queryHeadOid(ctx context.Context, workingBranch string, retry int) (*RepositoryRef, error) {
	rateLimit, err := queryRateLimit(g.client, ctx)
	logrus.Debugln(rateLimit)
	if err != nil {
		if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) {
			if retry < client.MaxRetry {
				logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, client.MaxRetry)
				rateLimit.Pause()
				return g.queryHeadOid(ctx, workingBranch, retry+1)
			}
			return nil, errors.New(ErrAPIRateLimitExceededFinalAttempt)
		}
		return nil, fmt.Errorf("unable to query GitHub API rate limit: %w", err)
	}

	var query struct {
		Repository struct {
			ID    string
			Name  string
			Owner struct {
				Login string
			}

			DefaultBranchRef *struct {
				Name   string
				Target struct {
					Oid string
				}
			}

			Ref *struct {
				Name   string
				Target struct {
					Oid string
				}
			} `graphql:"ref(qualifiedName: $qualifiedName)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
		RateLimit RateLimit
	}

	variables := map[string]interface{}{
		keyOwner:        githubv4.String(g.Spec.Owner),
		keyName:         githubv4.String(g.Spec.Repository),
		"qualifiedName": githubv4.String(workingBranch),
	}

	err = g.client.Query(ctx, &query, variables)
	if err != nil {
		if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) {
			if retry < client.MaxRetry {
				logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, client.MaxRetry)
				query.RateLimit.Pause()
				return g.queryHeadOid(ctx, workingBranch, retry+1)
			}
			return nil, errors.New(ErrAPIRateLimitExceededFinalAttempt)
		}
		return nil, fmt.Errorf("unable to query GitHub API: %w", err)
	}

	headOid := ""
	if query.Repository.Ref != nil {
		headOid = query.Repository.Ref.Target.Oid
	}

	defaultBranchOid := ""
	if query.Repository.DefaultBranchRef != nil {
		defaultBranchOid = query.Repository.DefaultBranchRef.Target.Oid
	}

	return &RepositoryRef{
		ID:               query.Repository.ID,
		HeadOid:          headOid,
		DefaultBranchOid: defaultBranchOid,
	}, nil
}

type refQuery struct {
	CreateRef struct {
		Ref struct {
			Name string
		}
	} `graphql:"createRef(input:$input)"`
}

// createBranch creates a new branch named branchName from the commit headOid
// using the GitHub GraphQL API.
func (g *Github) createBranch(ctx context.Context, branchName string, repositoryId string, headOid string, retry int) error {
	var query refQuery

	rateLimit, err := queryRateLimit(g.client, ctx)
	logrus.Debugln(rateLimit)
	if err != nil {
		if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) {
			if retry < client.MaxRetry {
				logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, client.MaxRetry)
				rateLimit.Pause()
				return g.createBranch(ctx, branchName, repositoryId, headOid, retry+1)
			}
			return errors.New(ErrAPIRateLimitExceededFinalAttempt)
		}
		return fmt.Errorf("unable to query GitHub API rate limit: %w", err)
	}

	input := githubv4.CreateRefInput{
		RepositoryID: repositoryId,
		Name:         githubv4.String(fmt.Sprintf("refs/heads/%s", branchName)),
		Oid:          githubv4.GitObjectID(headOid),
	}

	if err := g.client.Mutate(ctx, &query, input, nil); err != nil {
		return err
	}
	return nil
}
