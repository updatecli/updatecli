package git

import (
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/tmp"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/commit"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"
	git "github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// Spec contains settings to manipulate a git repository.
type Spec struct {
	// URL specifies the git url
	URL string `yaml:",omitempty" jsonschema:"required"`
	// Username specifies the username for http authentication
	Username string `yaml:",omitempty"`
	// Password specifies the password for http authentication
	Password string `yaml:",omitempty"`
	// Branch specifies the git branch
	Branch string `yaml:",omitempty"`
	// User specifies the git commit author
	User string `yaml:",omitempty"`
	// Email specifies the git commit email
	Email string `yaml:",omitempty"`
	// Directory specifies the directory to use for cloning the repository
	Directory string `yaml:",omitempty"`
	// Force is used during the git push phase to run `git push --force`.
	Force bool `yaml:",omitempty"`
	// CommitMessage contains conventional commit metadata as type or scope, used to generate the final commit message.
	CommitMessage commit.Commit `yaml:",omitempty"`
	// GPG key and passphrased used for commit signing
	GPG sign.GPGSpec `yaml:",omitempty"`
}

type Git struct {
	spec         Spec
	remoteBranch string
}

// New returns a new git object
func New(s Spec) (*Git, error) {
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

	return &Git{
		spec:         s,
		remoteBranch: git.SanitizeBranchName(s.Branch),
	}, nil

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
