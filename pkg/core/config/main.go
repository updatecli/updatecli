package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/mitchellh/hashstructure"
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

	// ErrNoEnvironmentVariableSet is returned when during the templating process,
	// it tries to access en environment variable not set.
	ErrNoEnvironmentVariableSet = errors.New("Environment variable doesn't exist")

	// ErrNoKeyDefined is returned when during the templating process, it tries to
	// retrieve a key value which is not defined in the configuration
	ErrNoKeyDefined = errors.New("key not defined in configuration")
)

// Config contains cli configuration
type Config struct {
	Name       string
	Title      string // Title is used for the full pipeline
	Source     source.Source
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

// Update updates its own configuration file
// It uses when the configuration expected a value that
// hasn't been set yet like a source output
func (config *Config) Update() (err error) {
	funcMap := template.FuncMap{
		"requiredEnv": func(s string) (string, error) {
			/*
				Retrieve value from environment variable, return error if not found
			*/
			value := os.Getenv(s)
			if value == "" {
				logrus.Debugf("%s", fmt.Sprintf("no value found for environment variable "+s))
				return "", ErrNoEnvironmentVariableSet
			}
			return value, nil
		},
		"pipeline": func(s string) (string, error) {
			/*
				Retrieve the value of a third location key from
				the updatecli configuration.
				It returns an error if a key doesn't exist
				It returns {{ pipeline "<key>" }} if a key exist but still set to zero value,
				then we assume that the value will be set later in the run.
				Otherwise it returns the value.
				This func is design to constantly reevaluate if a configuration changed
			*/

			var field func(interface{}, []string) (string, error)

			field = func(conf interface{}, query []string) (value string, err error) {
				ValueIface := reflect.ValueOf(conf)

				Field := reflect.Value{}

				switch ValueIface.Kind() {
				case reflect.Ptr:
					// Check if the passed interface is a pointer
					// Create a new type of Iface's Type, so we have a pointer to work with
					// 'dereference' with Elem() and get the field by name
					Field = ValueIface.Elem().FieldByName(query[0])
				case reflect.Map:
					Field = ValueIface.MapIndex(reflect.ValueOf(query[0]))
				case reflect.Struct:
					Field = ValueIface.FieldByName(query[0])
				}

				if !Field.IsValid() {
					logrus.Debugf(
						"Configuration `%s` does not have the field `%s`",
						ValueIface.Type(),
						query[0])
					return "", ErrNoKeyDefined
				}

				if len(query) > 1 {
					value, err = field(Field.Interface(), query[1:])
					if err != nil {
						return "", err
					}

				} else if len(query) == 1 {
					return Field.String(), nil
				}

				return value, nil

			}

			val, err := field(config, strings.Split(s, "."))

			if err != nil {
				return "", err
			}

			if len(val) > 0 {
				return val, nil
			}
			// If we couldn't find a value, then we return the function so we can retry
			// later on.
			return fmt.Sprintf("\"{{ pipeline \"%s\" }}\"", s), nil

		},
	}

	data := *config

	content, err := yaml.Marshal(config)

	if err != nil {
		return err
	}

	tmpl, err := template.New("cfg").Funcs(funcMap).Parse(string(content))

	if err != nil {
		return err
	}

	b := bytes.Buffer{}

	if err := tmpl.Execute(&b, &data); err != nil {
		return err
	}

	err = yaml.Unmarshal(b.Bytes(), &config)
	if err != nil {
		return err
	}

	err = config.Validate()
	if err != nil {
		return err
	}

	return err

}
