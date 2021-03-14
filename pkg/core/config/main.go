package config

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/olblak/updateCli/pkg/core/engine/condition"
	"github.com/olblak/updateCli/pkg/core/engine/source"
	"github.com/olblak/updateCli/pkg/core/engine/target"
	"gopkg.in/yaml.v3"
)

// Config contains cli configuration
type Config struct {
	Name       string
	Title      string // Title is used for the full pipeline
	Source     source.Source
	PipelineID string // PipelineID allows to identify a full pipeline run, this value is propagated into each target if not defined at that level
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

	// We need to be sure to generate a file checksum before we inject
	// templates values as in some situation those values changes for each run
	pipelineID, err := Checksum(cfgFile)
	if err != nil {
		return err
	}

	switch extension := filepath.Ext(basename); extension {
	case ".tpl", ".tmpl", ".yaml", ".yml":
		t := Template{
			CfgFile:    filepath.Join(dirname, basename),
			ValuesFile: valuesFile,
		}

		err := t.Unmarshal(config)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("file extension not supported: %v", extension)
	}

	// config.PipelineID is required for config.Validate()
	config.PipelineID = pipelineID

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
	for id, t := range config.Targets {
		if len(t.PipelineID) == 0 {
			t.PipelineID = config.PipelineID
		}
		config.Targets[id] = t
	}

	return nil
}

// Checksum return a file checksum using sha256.
func Checksum(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		logrus.Debugf("Can't open file %q", filename)
		return "", err
	}

	defer file.Close()
	hash := sha256.New()

	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
