package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mitchellh/hashstructure"
	"github.com/sirupsen/logrus"

	"github.com/olblak/updateCli/pkg/core/engine/condition"
	"github.com/olblak/updateCli/pkg/core/engine/source"
	"github.com/olblak/updateCli/pkg/core/engine/target"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config contains cli configuration
type Config struct {
	Name       string
	Source     source.Source
	Conditions map[string]condition.Condition
	Targets    map[string]target.Target
}

// Reset reset configuration
func (config *Config) Reset() {
	config.Source = source.Source{}
	config.Conditions = map[string]condition.Condition{}
	config.Targets = map[string]target.Target{}
}

// ReadFile reads the updatecli configuration file
func (config *Config) ReadFile(cfgFile, valuesFile string) (err error) {

	config.Reset()

	dirname, basename := filepath.Split(cfgFile)

	switch extension := filepath.Ext(basename); extension {
	case ".tpl", ".tmpl":
		t := Template{
			CfgFile:    filepath.Join(dirname, basename),
			ValuesFile: valuesFile,
		}

		err := t.Unmarshal(config)
		if err != nil {
			return err
		}

	case ".yaml", ".yml":
		v := viper.New()

		v.SetEnvPrefix("updatecli")
		v.AutomaticEnv()
		v.SetConfigName(strings.TrimSuffix(basename, filepath.Ext(basename))) // name of config file (without extension)
		v.SetConfigType(strings.Replace(filepath.Ext(basename), ".", "", -1)) // REQUIRED if the config file does not have the extension in the name
		v.AddConfigPath(dirname)                                              // optionally look for config in the working directory
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				return fmt.Errorf("Config file not found")
			}

			return err
		}
		err := v.Unmarshal(&config)
		if err != nil {
			return fmt.Errorf("unable to decode into struct, %v", err)
		}
	default:
		return fmt.Errorf("file extension not supported: %v", extension)
	}

	err = config.Validate()
	if err != nil {
		return err
	}

	return nil

}

// Display shows updatecli configuration including secrets !
func (config *Config) Display() error {
	c, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}
	logrus.Infof("%s", string(c))
	return nil
}

// Validate run various validation test on the configuration and update fields if necessary
func (config *Config) Validate() error {
	pipelineID, err := hashstructure.Hash(config, nil)
	if err != nil {
		return err
	}

	for id, t := range config.Targets {
		if len(t.PipelineID) == 0 {
			t.PipelineID = fmt.Sprintf("%d", pipelineID)
		}
		config.Targets[id] = t
	}

	return nil
}
