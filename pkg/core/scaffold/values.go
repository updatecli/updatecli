package scaffold

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var (
	valuesTemplate string = `---
# Values.yaml contains settings that be used from Updatecli manifest.
# scm:
#   default:
#     user: updatecli-bot
#     email: updatecli-bot@updatecli.io
#     owner: github_owner
#     repository: github_repository
#     username: "updatecli-bot"
#     branch: main
`
)

func (s *Scaffold) scaffoldValues(dirname string) error {

	dirname = filepath.Join(dirname)

	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		err := os.MkdirAll(dirname, 0755)
		if err != nil {
			return err
		}
	}

	valuesFilePath := filepath.Join(dirname, s.ValuesFile)

	if _, err := os.Stat(valuesFilePath); err == nil {
		logrus.Infof("Skipping, values already exist: %s", valuesFilePath)
		return nil
	}

	f, err := os.Create(valuesFilePath)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write([]byte(valuesTemplate))
	if err != nil {
		return err
	}

	return nil
}
