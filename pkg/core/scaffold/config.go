package scaffold

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var (
	configFile     string = "default.yaml"
	configTemplate string = `---
name: Default pipeline name

## scms defines the source control management system to interact with.
# scms:
#   default:
#     kind: github
#     spec:
#       owner: {{ .scm.default.owner }}
#       repository: {{ .scm.default.repository }}
#       branch: {{ .scm.default.branch }}
#       user: {{ .scm.default.user }}
#       email: {{ .scm.default.email }}
#       username: {{ .scm.default.username }}
#       token: {{ requiredEnv "UPDATECLI_GITHUB_TOKEN" }}

## actions defines what to do when a target with the same scmid is modified.
# actions:
#   default:
#     kind: "github/pullrequest"
#     scmid: "default"
#     spec:
#       automerge: false
#       labels:
#         - "dependencies"

## sources defines where to find the information.
# sources:
#   default:
#     name: "Short source description"
#     scmid: default
#     kind: specify the source plugin to use
#       spec:
#         # Specify the source plugin specific configuration

## conditions defines when to executes a target
# conditions:
#   default:
#     name: "Short condition description"
#     kind: specify the condition plugin to use
#     scmid: default
#     spec:
#       # Specify the condition plugin specific configuration

## targets defines where to apply the changes.
# targets:
#   default:
#     name: "Short target description"
#     kind: specify the target plugin to use
#     scmid: default
#     spec:
#       # Specify the target plugin specific configuration
`
)

func (s *Scaffold) scaffoldConfig(rootDir string) error {

	configDir = filepath.Join(rootDir, s.ConfigDir)

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			return err
		}
	}

	configFilePath := filepath.Join(configDir, configFile)

	// If the config already exist, we don't overwrite it
	if _, err := os.Stat(configFilePath); err == nil {
		logrus.Infof("Skipping, config already exist: %s", configFilePath)
		return nil
	}

	f, err := os.Create(configFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(configTemplate))
	if err != nil {
		return err
	}

	return nil
}
