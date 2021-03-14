package config

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/olblak/updateCli/pkg/core/engine/condition"
	"github.com/olblak/updateCli/pkg/core/engine/source"
	"github.com/olblak/updateCli/pkg/core/engine/target"
	"gopkg.in/yaml.v3"
)

var (
	// ErrConfigFileTypeNotSupported is returned when updatecli try to read
	// an unsupported file type.
	ErrConfigFileTypeNotSupported = errors.New("file extension not supported")

	// ErrBadConfig is returned when updatecli try to read
	// a wrong configuration.
	ErrBadConfig = errors.New("wrong updatecli configuration")

	// ErrNoEnvironmentVariableSet is returned when during the templating process,
	// it tries to access en environment variable not set.
	ErrNoEnvironmentVariableSet = errors.New("environment variable doesn't exist")

	// ErrNoKeyDefined is returned when during the templating process, it tries to
	// retrieve a key value which is not defined in the configuration
	ErrNoKeyDefined = errors.New("key not defined in configuration")

	defaultSourceID = "default"
)

// Config contains cli configuration
type Config struct {
	Name       string
	PipelineID string        // PipelineID allows to identify a full pipeline run, this value is propagated into each target if not defined at that level
	Title      string        // Title is used for the full pipeline
	Source     source.Source // **Deprecated** 2021/02/18 Is replaced by Sources, this setting will be deleted in a futur release
	Sources    map[string]source.Source
	Conditions map[string]condition.Condition
	Targets    map[string]target.Target
}

// Reset reset configuration
func (config *Config) Reset() {
	*config = Config{}
}

// New reads an updatecli configuration file
func New(cfgFile, valuesFile string) (config Config, err error) {

	config.Reset()

	dirname, basename := filepath.Split(cfgFile)

	// We need to be sure to generate a file checksum before we inject
	// templates values as in some situation those values changes for each run
	pipelineID, err := Checksum(cfgFile)
	if err != nil {
		return config, err
	}

	switch extension := filepath.Ext(basename); extension {
	case ".tpl", ".tmpl", ".yaml", ".yml":
		t := Template{
			CfgFile:    filepath.Join(dirname, basename),
			ValuesFile: valuesFile,
		}

		err := t.Init(&config)
		if err != nil {
			return config, err
		}

	default:
		logrus.Debugf("file extension '%s' not supported for file '%s'", extension, filepath.Join(dirname, basename))
		return config, ErrConfigFileTypeNotSupported
	}

	// config.PipelineID is required for config.Validate()
	config.PipelineID = pipelineID

	err = config.Validate()
	if err != nil {
		return config, err
	}

	if len(config.Name) == 0 {
		config.Name = strings.ToTitle(basename)
	}

	err = config.Validate()

	return config, err

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

	if config.Source.Kind != "" && (len(config.Sources) > 0) {
		logrus.Errorf("Source and Sources can't be defined at the same time, please use the 'Sources' syntax")
		return ErrBadConfig

	} else if config.Source.Kind != "" && len(config.Sources) == 0 {

		logrus.Warning("Since version 1.2.0, the single source definition is **Deprecated**  and replaced by Sources. This parameter will be deleted in a future release")

		config.Sources = make(map[string]source.Source)
		config.Sources[defaultSourceID] = config.Source
		config.Source = source.Source{}
	}

	for id, c := range config.Conditions {
		// Try to guess SourceID
		if len(c.SourceID) == 0 && len(config.Sources) > 1 {
			logrus.Errorf("{empty 'sourceID' for condition '%s'", id)
			return ErrBadConfig
		} else if len(c.SourceID) == 0 && len(config.Sources) == 1 {
			for id := range config.Sources {
				c.SourceID = id
			}
		}
		config.Conditions[id] = c
	}

	for id, t := range config.Targets {
		if len(t.PipelineID) == 0 {
			t.PipelineID = config.PipelineID
		}
		// Try to guess SourceID
		if len(t.SourceID) == 0 && len(config.Sources) > 1 {
			logrus.Errorf("{empty 'sourceID' for target '%s'", id)
			return ErrBadConfig
		} else if len(t.SourceID) == 0 && len(config.Sources) == 1 {
			for id := range config.Sources {
				t.SourceID = id
			}
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
