package gitcommit

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
)

// Spec defines a specification for a "gitcommit" resource.
type Spec struct {
	// Path specifies a local Git repository path.
	//
	// compatible:
	//   * source
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
	// Depth limits the number of commits fetched from the Git repository.
	//
	// compatible:
	//   * source
	//
	// default:
	//   0 (no limit)
	Depth *int `yaml:",omitempty"`
	// URL specifies the Git repository URL to clone.
	//
	// compatible:
	//   * source
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
	Username string `yaml:",omitempty"`
	// Password specifies the password used with the HTTP protocol.
	//
	// compatible:
	//   * source
	Password string `yaml:",omitempty"`
}

// GitCommit defines a resource of kind "gitcommit".
type GitCommit struct {
	spec             Spec
	nativeGitHandler commitHashFinder
	directory        string
}

type commitHashFinder interface {
	GetCommitHash(workingDir, branch string) (string, error)
}

// New returns a newly initialized GitCommit resource.
func New(spec interface{}) (*GitCommit, error) {
	newSpec := Spec{}
	if err := mapstructure.Decode(spec, &newSpec); err != nil {
		return nil, err
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
		Depth:  gc.spec.Depth,
		URL:    redact.URL(gc.spec.URL),
	}
}
