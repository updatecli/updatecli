package scaffold

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type scaffoldConfig struct {
	Path    string
	Content string
}

var (

	//go:embed assets/updatecli.d/default.yaml
	assetDefaultConfig string

	//go:embed assets/updatecli.d/_scm.github.yaml
	assetScmGitHubPartialConfig string
	//go:embed assets/updatecli.d/_scm.bitbucket.yaml
	assetScmBitbucketPartialConfig string
	//go:embed assets/updatecli.d/_scm.gitea.yaml
	assetScmGiteaPartialConfig string
	//go:embed assets/updatecli.d/_scm.gitlab.yaml
	assetScmGitlabPartialConfig string
	//go:embed assets/updatecli.d/_scm.stash.yaml
	assetScmStashPartialConfig string

	//go:embed assets/values.yaml
	assetValues string

	//go:embed assets/README.md
	assetReadme string

	//go:embed assets/CHANGELOG.md
	assetChangelog string

	scaffoldConfigs []scaffoldConfig = []scaffoldConfig{
		{
			Path:    "updatecli.d/_scm.bitbucket.yaml",
			Content: assetScmBitbucketPartialConfig,
		},
		{
			Path:    "updatecli.d/_scm.github.yaml",
			Content: assetScmGitHubPartialConfig,
		},
		{
			Path:    "updatecli.d/_scm.gitea.yaml",
			Content: assetScmGiteaPartialConfig,
		},
		{
			Path:    "updatecli.d/_scm.gitlab.yaml",
			Content: assetScmGitlabPartialConfig,
		},
		{
			Path:    "updatecli.d/_scm.stash.yaml",
			Content: assetScmStashPartialConfig,
		},
		{
			Path:    "updatecli.d/default.example.yaml",
			Content: assetDefaultConfig,
		},
		{
			Path:    "values.yaml",
			Content: assetValues,
		},
		{
			Path:    "README.md",
			Content: assetReadme,
		},
		{
			Path:    "CHANGELOG.md",
			Content: assetChangelog,
		},
	}
)

func (s *Scaffold) scaffoldConfig(rootDir string) error {

	for _, cf := range scaffoldConfigs {
		defaultConfigDir := filepath.Join(rootDir, filepath.Dir(cf.Path))

		if _, err := os.Stat(defaultConfigDir); os.IsNotExist(err) {
			err := os.MkdirAll(defaultConfigDir, 0755)
			if err != nil {
				logrus.Errorf("Failed to create config directory: %s", defaultConfigDir)
				continue
			}
		}

		configFilePath := filepath.Join(rootDir, cf.Path)

		// If the config already exist, we don't overwrite it
		if _, err := os.Stat(configFilePath); err == nil {
			logrus.Infof("Skipping, config already exist: %s", configFilePath)
			continue
		}

		f, err := os.Create(configFilePath)
		if err != nil {
			logrus.Errorf("Failed to create config file: %s", configFilePath)
			continue
		}

		_, err = f.Write([]byte(cf.Content))
		if err != nil {
			logrus.Errorf("Failed to write config file: %s", configFilePath)
			f.Close()
			continue
		}
		f.Close()
	}

	return nil
}
