package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/shurcooL/githubv4"

	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/core/tmp"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/commit"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"

	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// Spec represents the configuration input
type Spec struct {
	//  "branch" defines the git branch to work on.
	//
	//  compatible:
	//    * scm
	//
	//  default:
	//    main
	//
	//  remark:
	//    depending on which resource references the GitHub scm, the behavior will be different.
	//
	//    If the scm is linked to a source or a condition (using scmid), the branch will be used to retrieve
	//    file(s) from that branch.
	//
	//    If the scm is linked to target then Updatecli creates a new "working branch" based on the branch value.
	//    The working branch created by Updatecli looks like "updatecli_<pipelineID>".
	//    The working branch can be disabled using the "workingBranch" parameter set to false.
	Branch string `yaml:",omitempty"`
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
	//    /tmp/updatecli/github/<owner>/<repository>
	Directory string `yaml:",omitempty"`
	//  "email" defines the email used to commit changes.
	//
	//  compatible:
	//    * scm
	//
	//  default:
	//    default set to your global git configuration
	Email string `yaml:",omitempty"`
	//  "owner" defines the owner of a repository.
	//
	//  compatible:
	//    * scm
	Owner string `yaml:",omitempty" jsonschema:"required"`
	//  "repository" specifies the name of a repository for a specific owner.
	//
	//  compatible:
	//    * scm
	Repository string `yaml:",omitempty" jsonschema:"required"`
	//	"token" specifies the credential used to authenticate with GitHub API.
	//
	//	compatible:
	//		* scm
	Token string `yaml:",omitempty" jsonschema:"required"`
	//  "url" specifies the default github url in case of GitHub enterprise
	//
	//  compatible:
	//    * scm
	//
	//  default:
	//    github.com
	//
	//  remark:
	//    A token is a sensitive information, it's recommended to not set this value directly in the configuration file
	//    but to use an environment variable or a SOPS file.
	//
	//    The value can be set to `{{ requiredEnv "GITHUB_TOKEN"}}` to retrieve the token from the environment variable `GITHUB_TOKEN`
	//	  or `{{ .github.token }}` to retrieve the token from a SOPS file.
	//
	//	  For more information, about a SOPS file, please refer to the following documentation:
	//    https://github.com/getsops/sops
	//
	URL string `yaml:",omitempty"`
	//  "username" specifies the username used to authenticate with GitHub API.
	//
	//  compatible:
	//    * scm
	//
	//  remark:
	//    the token is usually enough to authenticate with GitHub API. Needed when working with GitHub private repositories.
	Username string `yaml:",omitempty"`
	//  "user" specifies the user associated with new git commit messages created by Updatecli
	//
	//  compatible:
	//    * scm
	User string `yaml:",omitempty"`
	//  "gpg" specifies the GPG key and passphrased used for commit signing
	//
	//  compatible:
	//    * scm
	GPG sign.GPGSpec `yaml:",omitempty"`
	//	"force" is used during the git push phase to run `git push --force`.
	//
	//	compatible:
	//    * scm
	//
	//	default:
	//    false
	//
	//  remark:
	//    When force is set to true, Updatecli also recreates the working branches that
	//    diverged from their base branch.
	Force *bool `yaml:",omitempty"`
	//	"commitMessage" is used to generate the final commit message.
	//
	//	compatible:
	//		* scm
	//
	//	remark:
	//		it's worth mentioning that the commit message settings is applied to all targets linked to the same scm.
	CommitMessage commit.Commit `yaml:",omitempty"`
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
	//  "commitUsingApi" defines if Updatecli should use GitHub GraphQL API to create the commit.
	//  When set to `true`, a commit created from a GitHub action using the GITHUB_TOKEN will automatically be signed by GitHub.
	//  More info on https://github.com/updatecli/updatecli/issues/1914
	//
	//  compatible:
	//	  * scm
	//
	//  default: false
	CommitUsingAPI *bool `yaml:",omitempty"`
}

// GitHub contains settings to interact with GitHub
type Github struct {
	force bool
	// Spec contains inputs coming from updatecli configuration
	Spec             Spec
	pipelineID       string
	client           GitHubClient
	nativeGitHandler gitgeneric.GitHandler
	workingBranch    bool
	commitUsingApi   bool
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

	if s.URL == "" {
		s.URL = "github.com"
	}

	if !strings.HasPrefix(s.URL, "https://") && !strings.HasPrefix(s.URL, "http://") {
		s.URL = "https://" + s.URL
	}

	if s.Username == "" {
		s.Username = "oauth2"
	}

	// Initialize github client
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: s.Token},
	)

	clientContext := context.WithValue(
		context.Background(),
		oauth2.HTTPClient,
		httpclient.NewRetryClient().(*http.Client))

	httpClient := oauth2.NewClient(clientContext, src)

	nativeGitHandler := gitgeneric.GoGit{}

	// By default, we create a working branch but if for some reason we don't want to create it
	// Then we also need to update the force safeguard to avoid force pushing on the main branch.
	workingBranch := true
	if s.WorkingBranch != nil {
		workingBranch = *s.WorkingBranch
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

	g := Github{
		force:            force,
		Spec:             s,
		pipelineID:       pipelineID,
		nativeGitHandler: &nativeGitHandler,
		workingBranch:    workingBranch,
		commitUsingApi:   commitUsingApi,
	}

	if strings.HasSuffix(s.URL, "github.com") {
		g.client = githubv4.NewClient(httpClient)
	} else {
		// For GH enterprise the GraphQL API path is /api/graphql
		// Cf https://docs.github.com/en/enterprise-cloud@latest/graphql/guides/managing-enterprise-accounts#3-setting-up-insomnia-to-use-the-github-graphql-api-with-enterprise-accounts
		graphqlURL, err := url.JoinPath(s.URL, "/api/graphql")
		if err != nil {
			return nil, err
		}
		g.client = githubv4.NewEnterpriseClient(graphqlURL, httpClient)
	}

	g.setDirectory()

	return &g, nil
}

// Validate verifies if mandatory GitHub parameters are provided and return false if not.
func (s *Spec) Validate() (errs []error) {
	required := []string{}

	if len(s.Token) == 0 {
		required = append(required, "token")
	}

	if len(s.Owner) == 0 {
		required = append(required, "owner")
	}

	if len(s.Repository) == 0 {
		required = append(required, "repository")
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

func (g *Github) queryRepository(sourceBranch string, workingBranch string) (*Repository, error) {
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
	}

	variables := map[string]interface{}{
		"owner":         githubv4.String(g.Spec.Owner),
		"name":          githubv4.String(g.Spec.Repository),
		"qualifiedName": githubv4.String(sourceBranch),
		"headRef":       githubv4.String(workingBranch),
	}

	err := g.client.Query(context.Background(), &query, variables)

	if err != nil {
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
func (g *Github) queryHeadOid(workingBranch string) (*RepositoryRef, error) {
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
	}

	variables := map[string]interface{}{
		"owner":         githubv4.String(g.Spec.Owner),
		"name":          githubv4.String(g.Spec.Repository),
		"qualifiedName": githubv4.String(workingBranch),
	}

	err := g.client.Query(context.Background(), &query, variables)
	if err != nil {
		logrus.Errorf("err - %s", err)
		return nil, err
	}

	headOid := ""
	if query.Repository.Ref != nil {
		headOid = query.Repository.Ref.Target.Oid
	}

	return &RepositoryRef{
		ID:               query.Repository.ID,
		HeadOid:          headOid,
		DefaultBranchOid: query.Repository.DefaultBranchRef.Target.Oid,
	}, nil
}

type refQuery struct {
	CreateRef struct {
		Ref struct {
			Name string
		}
	} `graphql:"createRef(input:$input)"`
}

func (g *Github) createBranch(branchName string, repositoryId string, headOid string) error {
	var query refQuery

	input := githubv4.CreateRefInput{
		RepositoryID: repositoryId,
		Name:         githubv4.String(fmt.Sprintf("refs/heads/%s", branchName)),
		Oid:          githubv4.GitObjectID(headOid),
	}

	if err := g.client.Mutate(context.Background(), &query, input, nil); err != nil {
		return err
	}
	return nil
}
