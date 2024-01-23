package scaffold

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var (
	changelogFile     string = "CHANGELOG.md"
	changelogTemplate string = `# CHANGELOG

## 0.1.0

  * Initial release
`
)

func (s *Scaffold) scaffoldChangelog(dirname string) error {

	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		err := os.MkdirAll(dirname, 0755)
		if err != nil {
			return err
		}
	}

	changelogFilePath := filepath.Join(dirname, changelogFile)

	// If the changelog already exist, we don't overwrite it
	if _, err := os.Stat(changelogFilePath); err == nil {
		logrus.Infof("Skipping, changelog already exist: %s", changelogFilePath)
		return nil
	}

	f, err := os.Create(changelogFilePath)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write([]byte(changelogTemplate))
	if err != nil {
		return err
	}

	return nil
}
