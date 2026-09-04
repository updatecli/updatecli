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

// Spec defines a specification for a "gitcommit" resource.
type Spec struct {
	// Path specifies a local Git repository path.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remarks:
	//   * Path overrides the working directory provided by an SCM configuration.
	Path string `yaml:",omitempty"`
	// Branch specifies the branch whose latest commit hash is returned.
	//
	// compatible:
	//   * source
	//
	// default:
	//   The repository's current HEAD branch.
	Branch string `yaml:",omitempty"`
	// Age defines the minimum or maximum age of a commit to be considered valid.
	// It accepts a duration string (e.g., "24h", "7d", "3w", "1y").
	// The age of a commit is its committer date, and the newest commit of the branch
	// falling inside that window is returned.
	//
	// compatible:
	//   * source
	//
	// remarks:
	//   * The branch history is walked from its tip until a commit matches, so Depth
	//     must be large enough to reach it.
	Age age.Spec `yaml:",omitempty"`
	// Hash specifies the commit hash checked by the condition.
	//
	// compatible:
	//   * condition
	//
	// default:
	//   The source output.
	Hash string `yaml:",omitempty"`
	// Depth limits the number of commits fetched from the Git repository.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   0 (no limit)
	Depth *int `yaml:",omitempty"`
	// URL specifies the Git repository URL to clone.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// example:
	//   * git@github.com:updatecli/updatecli.git
	//   * https://github.com/updatecli/updatecli.git
	//
	// remarks:
	//   * URL overrides both Path and the working directory provided by an SCM configuration.
	URL string `yaml:",omitempty"`
	// Username specifies the username used with the HTTP protocol.
	//
	// compatible:
	//   * source
	//   * condition
	Username string `yaml:",omitempty"`
	// Password specifies the password used with the HTTP protocol.
	//
	// compatible:
	//   * source
	//   * condition
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
