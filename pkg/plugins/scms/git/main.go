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

// Git contains settings to manipulate a git repository.
type Git struct {
	URL           string
	Username      string
	Password      string
	Branch        string
	remoteBranch  string
	User          string
	Email         string
	Directory     string
	Version       string
	Force         bool          // Force is used during the git push phase to run `git push --force`.
	CommitMessage commit.Commit // CommitMessage contains conventional commit metadata as type or scope, used to generate the final commit message.
	GPG           sign.GPGSpec  // GPG key and passphrased used for commit signing
}

func New(g Git) (*Git, error) {
	var err error
	if len(g.Directory) == 0 {
		g.Directory, err = newDirectory(g.URL)
		if err != nil {
			return nil, err
		}
	}

	if len(g.Branch) == 0 {
		g.Branch = "main"
	}

	g.remoteBranch = git.SanitizeBranchName(g.Branch)
	return &g, nil

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
