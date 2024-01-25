package scaffold

import (
	"github.com/sirupsen/logrus"
)

var (
	// defaultPolicyFile is the default policy file name
	defaultPolicyFile = "Policy.yaml"
	// defaultSecretsDir is the default secrets directory name
	defaultSecretsDir = "secrets.d"
	// defaultConfigDir is the default config directory name
	defaultConfigDir = "updatecli.d"
	// defaultValuesFile is the default values file name
	defaultValuesFile = "values.yaml"
)

// Scaffold is the main structure to scaffold a new Updatecli policy
type Scaffold struct {
	// PolicyFile is the policy file name
	PolicyFile string
	// ValuesFile is the values directory name
	ValuesFile string
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

	setDefaultValues(&s.ConfigDir, defaultConfigDir)
	setDefaultValues(&s.PolicyFile, defaultPolicyFile)
	setDefaultValues(&s.SecretsDir, defaultSecretsDir)
	setDefaultValues(&s.ValuesFile, defaultValuesFile)
}

// Run scaffold a new Updatecli policy
func (s *Scaffold) Run(rootDir string) error {
	s.Init()
	logrus.Debugf("Initialize an Updatecli policy")

	err := s.scaffoldPolicy(&PolicySpec{}, rootDir)
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
