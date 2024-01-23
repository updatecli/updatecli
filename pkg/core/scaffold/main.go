package scaffold

import (
	"github.com/sirupsen/logrus"
)

var (
	// defaultPolicyFile is the default policy file name
	defaultPolicyFile = "Policy.yaml"
	// defaultValuesDir is the default values directory name
	defaultValuesDir = "values.d"
	// defaultSecretsDir is the default secrets directory name
	defaultSecretsDir = "secrets.d"
	// configDir is the default config directory name
	configDir = "updatecli.d"
)

// Scaffold is the main structure to scaffold a new Updatecli policy
type Scaffold struct {
	// PolicyFile is the policy file name
	PolicyFile string
	// ValuesDir is the values directory name
	ValuesDir string
	// SecretsDir is the secrets directory name
	SecretsDir string
	// ConfigDir is the config directory name
	ConfigDir string
}

// Init initialize a new scaffold
func (s *Scaffold) Init() {

	setDefaultValues := func(s *string, defaultValue string) {
		if *s == "" {
			*s = defaultValue
		}
	}

	setDefaultValues(&s.ConfigDir, configDir)
	setDefaultValues(&s.PolicyFile, defaultPolicyFile)
	setDefaultValues(&s.SecretsDir, defaultSecretsDir)
	setDefaultValues(&s.ValuesDir, defaultValuesDir)
}

// Run scaffold a new Updatecli policy
func (s *Scaffold) Run(rootDir string) error {
	s.Init()
	logrus.Debugf("Initialize an Updatecli policy")

	err := s.scaffoldPolicy(&PolicySpec{}, rootDir, s.PolicyFile)
	if err != nil {
		return err
	}

	err = s.scaffoldValues(rootDir)
	if err != nil {
		return err
	}

	err = s.scaffoldConfig(rootDir)
	if err != nil {
		return err
	}

	err = s.scaffoldReadme(rootDir)
	if err != nil {
		return err
	}

	err = s.scaffoldChangelog(rootDir)
	if err != nil {
		return err
	}

	return nil
}
