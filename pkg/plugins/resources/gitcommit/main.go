package gitcommit

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
)

/*
"gitcommit" defines the specification for retrieving and checking Git commit hashes.
It can be used as a "source" or a "condition".
*/
type Spec struct {
	// "path" defines the path of the local Git repository.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * a relative path is resolved from the Updatecli manifest directory.
	//   * "path" overrides the working directory provided by the scm configuration.
	//   * "url" takes precedence over "path".
	//
	Path string `yaml:",omitempty"`
	// "branch" defines the branch whose latest commit hash is returned.
	//
	// compatible:
	//   * source
	//
	// default:
	//   the current HEAD of the repository.
	//
	Branch string `yaml:",omitempty"`
	// "age" defines the minimum or maximum age of a commit to be considered valid.
	//
	// compatible:
	//   * source
	//
	// remark:
	//   * "minimum" and "maximum" accept a duration string such as "24h", "7d", "3w" or "1y".
	//   * the age of a commit is its committer date.
	//   * the newest commit of the branch inside that window is returned.
	//   * the branch history is walked from its tip until a commit matches,
	//     so "depth" must be large enough to reach it.
	//   * when no commit matches the age filter, the source is skipped.
	//
	// example:
	// ```
	//   age:
	//     minimum: 7d
	// ```
	//
	Age age.Spec `yaml:",omitempty"`
	// "hash" defines the commit hash checked by the condition.
	//
	// compatible:
	//   * condition
	//
	// default:
	//   the output of the associated source.
	//
	Hash string `yaml:",omitempty"`
	// "depth" limits the number of commits fetched from the Git repository.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   0, which means no limit
	//
	Depth *int `yaml:",omitempty"`
	// "url" defines the Git repository URL to clone.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * "url" overrides "path" and the working directory provided by the scm configuration.
	//
	// example:
	//   * url: git@github.com:updatecli/updatecli.git
	//   * url: https://github.com/updatecli/updatecli.git
	//
	URL string `yaml:",omitempty"`
	// "username" defines the username used with the HTTP protocol.
	//
	// compatible:
	//   * source
	//   * condition
	//
	Username string `yaml:",omitempty"`
	// "password" defines the password used with the HTTP protocol.
	//
	// compatible:
	//   * source
	//   * condition
	//
	Password string `yaml:",omitempty"`
}

// GitCommit defines a resource of kind "gitcommit".
type GitCommit struct {
	spec             Spec
	nativeGitHandler commitHandler
	directory        string
}

type commitHandler interface {
	GetCommitHash(workingDir, branch string) (string, error)
	IsCommitExist(workingDir, commit string) (bool, error)
	SearchCommit(workingDir, branch string, match func(when time.Time) bool) (gitgeneric.DatedCommit, error)
}

// New returns a newly initialized GitCommit resource.
func New(spec interface{}) (*GitCommit, error) {
	newSpec := Spec{}
	if err := mapstructure.Decode(spec, &newSpec); err != nil {
		return nil, err
	}

	validationErrors := []string{}

	if err := newSpec.Age.Validate(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Return all the validation errors if found any
	if len(validationErrors) > 0 {
		return nil, fmt.Errorf("validation error: the provided manifest configuration has the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	return &GitCommit{
		spec:             newSpec,
		nativeGitHandler: &gitgeneric.GoGit{},
	}, nil
}

// Changelog returns nil because changelogs are not supported by this resource.
func (gc *GitCommit) Changelog(from, to string) *result.Changelogs {
	return nil
}

func (gc *GitCommit) clone() (string, error) {
	g, err := git.New(git.Spec{
		URL:      gc.spec.URL,
		Username: gc.spec.Username,
		Password: gc.spec.Password,
		Depth:    gc.spec.Depth,
	}, "")
	if err != nil {
		return "", err
	}
	return g.Clone()
}

// ReportConfig returns the non-sensitive resource configuration.
func (gc *GitCommit) ReportConfig() interface{} {
	return Spec{
		Path:   gc.spec.Path,
		Branch: gc.spec.Branch,
		Age:    gc.spec.Age,
		Hash:   gc.spec.Hash,
		Depth:  gc.spec.Depth,
		URL:    redact.URL(gc.spec.URL),
	}
}
