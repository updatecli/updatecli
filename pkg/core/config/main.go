package config

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/pullRequest"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
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

	// ErrNotAllowedTemplatedKey is returned when
	// we are planning to template at runtime unauthorized keys such map key
	ErrNotAllowedTemplatedKey = errors.New("not allowed templated key")

	defaultSourceID = "default"
)

// Config contains cli configuration
type Config struct {
	Name         string
	PipelineID   string                        // PipelineID allows to identify a full pipeline run, this value is propagated into each target if not defined at that level
	Title        string                        // Title is used for the full pipeline
	Source       source.Config                 // **Deprecated** 2021/02/18 Is replaced by Sources, this setting will be deleted in a future release
	PullRequests map[string]pullRequest.Config // PullRequests defines the list of Pull Request configuration which need to be managed
	SCMs         map[string]scm.Config         `yaml:"scms"` // SCMs defines the list of repository configuration used to fetch content from.
	Sources      map[string]source.Config      // Sources defines the list of source configuration
	Conditions   map[string]condition.Config   // Conditions defines the list of condition configuration
	Targets      map[string]target.Config      // Targets defines the list of target configuration
}

// Reset reset configuration
func (config *Config) Reset() {
	*config = Config{}
}

// New reads an updatecli configuration file
func New(cfgFile string, valuesFiles, secretsFiles []string) (config Config, err error) {

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
			CfgFile:      filepath.Join(dirname, basename),
			ValuesFiles:  valuesFiles,
			SecretsFiles: secretsFiles,
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

		logrus.Warning("Since version 0.2.0, the single source definition is **Deprecated**  and replaced by Sources. This parameter will be deleted in a future release")

		config.Sources = make(map[string]source.Config)
		config.Sources[defaultSourceID] = config.Source
		config.Source = source.Config{}
	}

	for _, scm := range config.SCMs {
		if err := scm.Validate(); err != nil {
			return err
		}
	}

	for _, p := range config.PullRequests {
		if err := p.Validate(); err != nil {
			return err
		}
	}

	for id, s := range config.Sources {
		if IsTemplatedString(id) {
			logrus.Errorf("sources key %q contains forbidden go template instruction", id)
			return ErrNotAllowedTemplatedKey
		}

		// Temporary code until we fully remove the old way to configure scm
		// Introduce by https://github.com/updatecli/updatecli/issues/260
		if len(s.Scm) > 0 {
			logrus.Warningln("source.Scm is now deprecated. Please use the new syntax")
			if len(s.SCMID) == 0 {
				if _, ok := config.SCMs["source_"+id]; !ok {
					for kind, spec := range s.Scm {
						if config.SCMs == nil {
							config.SCMs = make(map[string]scm.Config, 1)
						}
						config.SCMs["source_"+id] = scm.Config{
							Kind: kind,
							Spec: spec}
					}
				}
				s.SCMID = "source_" + id
			} else {
				logrus.Warning("source.SCMID is also defined, ignoring source.Scm")
			}
			s.Scm = map[string]interface{}{}
			config.Sources[id] = s
		}
	}

	for id, c := range config.Conditions {

		// Only check/guess the sourceID if the user did not disable it (default is enabled)
		if !c.DisableSourceInput {
			// Try to guess SourceID
			if len(c.SourceID) == 0 && len(config.Sources) > 1 {
				logrus.Errorf("The condition %q has an empty 'sourceID' attribute.", id)
				return ErrBadConfig
			} else if len(c.SourceID) == 0 && len(config.Sources) == 1 {
				for id := range config.Sources {
					c.SourceID = id
				}
			}
		}

		if IsTemplatedString(id) {
			logrus.Errorf("condition key %q contains forbidden go template instruction", id)
			return ErrNotAllowedTemplatedKey
		}

		// Temporary code until we fully remove the old way to confgigure scm
		// Introduce by https://github.com/updatecli/updatecli/issues/260
		//if c.Scm != nil {
		if len(c.Scm) > 0 {
			logrus.Warningln("condition.Scm is now deprecated. Please use the new syntax")
			if len(c.SCMID) == 0 {
				if _, ok := config.SCMs["condition_"+id]; !ok {
					for kind, spec := range c.Scm {

						if config.SCMs == nil {
							config.SCMs = make(map[string]scm.Config, 1)
						}

						config.SCMs["condition_"+id] = scm.Config{
							Kind: kind,
							Spec: spec}

					}
				}
				c.SCMID = "condition_" + id
				// Ensure we clean this deprecated variable
			} else {
				logrus.Warning("condition.SCMID is also defined, ignoring condition.Scm")
			}
			c.Scm = map[string]interface{}{}
		}

		config.Conditions[id] = c
	}

	for id, t := range config.Targets {
		if len(t.PipelineID) == 0 {
			t.PipelineID = config.PipelineID
		}
		// Try to guess SourceID
		if len(t.SourceID) == 0 && len(config.Sources) > 1 {

			logrus.Errorf("empty 'sourceID' for target %q", id)
			return ErrBadConfig
		} else if len(t.SourceID) == 0 && len(config.Sources) == 1 {
			for id := range config.Sources {
				t.SourceID = id
			}
		}

		if IsTemplatedString(id) {
			logrus.Errorf("target key %q contains forbidden go template instruction", id)
			return ErrNotAllowedTemplatedKey
		}

		// Temporary code until we fully remove the old way to confgigure scm
		// Introduce by https://github.com/updatecli/updatecli/issues/260
		//if t.Scm != nil {
		if len(t.Scm) > 0 {
			logrus.Warningln("target.Scm is now deprecated. Please use the new syntax")
			if len(t.SCMID) == 0 {
				if _, ok := config.SCMs["target_"+id]; !ok {
					for kind, spec := range t.Scm {
						if config.SCMs == nil {
							config.SCMs = make(map[string]scm.Config, 1)
						}

						config.SCMs["target_"+id] = scm.Config{
							Kind: kind,
							Spec: spec,
						}

						// Temporary code until we fully deprecated the old pullRequest syntax
						// at this stage we can only have github pullrequest so to not break backward
						// compatibility, we automatically add a pullRequest configuration in case of github scm
						if kind == "github" {
							if config.PullRequests == nil {
								config.PullRequests = make(map[string]pullRequest.Config, 1)
							}
							config.PullRequests["target_"+id] = pullRequest.Config{
								Kind:             kind,
								Spec:             spec,
								DependsOnTargets: []string{id},
							}
						}

					}
				}
				t.SCMID = "target_" + id
			} else {
				logrus.Warning("target.SCMID is also defined, ignoring target.Scm")
			}
			t.Scm = map[string]interface{}{}
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

// Update updates its own configuration file
// It's used when the configuration expected a value defined a runtime
func (config *Config) Update(data interface{}) (err error) {
	funcMap := template.FuncMap{
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

			val, err := getFieldValueByQuery(data, strings.Split(s, "."))

			if err != nil {
				return "", err
			}

			if len(val) > 0 {
				return val, nil
			}
			// If we couldn't find a value, then we return the function so we can retry
			// later on.
			return fmt.Sprintf("{{ pipeline %q }}", s), nil

		},
		"source": func(s string) (string, error) {
			/*
				Retrieve the value of a third location key from
				the updatecli contex.
				It returns an error if a key doesn't exist
				It returns {{ pipeline "<key>" }} if a key exist but still set to zero value,
				then we assume that the value will be set later in the run.
				Otherwise it returns the value.
				This func is design to constantly reevaluate if a configuration changed
			*/

			val, err := getFieldValueByQuery(
				data, []string{"Sources", s, "Output"})

			if err != nil {
				return "", err
			}

			if len(val) > 0 {
				return val, nil
			}
			// If we couldn't find a value, then we return the function so we can retry
			// later on.
			return fmt.Sprintf("{{ source %q }}", s), nil

		},
	}

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

// IsTemplatedString test if a string contains go template information
func IsTemplatedString(s string) bool {
	if len(s) == 0 {
		return false
	}

	leftDelimiterFound := false

	for _, val := range strings.SplitAfter(s, "{{") {
		if strings.Contains(val, "{{") {
			leftDelimiterFound = true
			continue
		}
		if strings.Contains(val, "}}") && leftDelimiterFound {
			return true
		}
	}

	return false
}

func getFieldValueByQuery(conf interface{}, query []string) (value string, err error) {
	ValueIface := reflect.ValueOf(conf)

	Field := reflect.Value{}

	// We want to be able to use case insensitive key
	insensitiveQuery := []string{
		query[0],
		strings.ToLower(query[0]),
		strings.Title(strings.ToLower(query[0])),
		strings.Title(query[0]),
		strings.ToTitle(query[0]),
	}

	switch ValueIface.Kind() {
	case reflect.Ptr:
		// Check if the passed interface is a pointer
		// Create a new type of Iface's Type, so we have a pointer to work with
		// 'dereference' with Elem() and get the field by name
		//Field = ValueIface.Elem().FieldByName(query[0])

		for _, q := range insensitiveQuery {
			Field = ValueIface.Elem().FieldByName(q)
			if Field.IsValid() {
				query[0] = q
				break
			}
		}
	case reflect.Map:
		// We want to be able to use case insensitive key
		for _, q := range insensitiveQuery {
			Field = ValueIface.MapIndex(reflect.ValueOf(q))
			if Field.IsValid() {
				query[0] = q
				break
			}
		}
	case reflect.Struct:
		// We want to be able to use case insensitive key
		for _, q := range insensitiveQuery {
			Field = ValueIface.FieldByName(q)
			if Field.IsValid() {
				break
			}
		}
	}

	// Means that despite the different case sensitive key, we couldn't find it
	if !Field.IsValid() {
		logrus.Debugf(
			"Configuration `%s` does not have the field `%s`",
			ValueIface.Type(),
			query[0])
		return "", ErrNoKeyDefined
	}

	if len(query) > 1 {
		value, err = getFieldValueByQuery(Field.Interface(), query[1:])
		if err != nil {
			return "", err
		}

	} else if len(query) == 1 {
		return Field.String(), nil
	}

	return value, nil

}

// GetChangelogTitle try to guess a specific target based on various information available for
// a specific job
func (config *Config) GetChangelogTitle(ID string, fallback string) (title string) {

	if len(config.Title) > 0 {
		// If a pipeline title has been defined, then use it for pull request title
		title = fmt.Sprintf("[updatecli] %s",
			config.Title)

	} else if len(config.Targets) == 1 && len(config.Targets[ID].Name) > 0 {
		// If we only have one target then we can use it as fallback.
		// Reminder, map in golang are not sorted so the order can't be kept between updatecli run
		title = fmt.Sprintf("[updatecli] %s", config.Targets[ID].Name)
	} else {
		// At the moment, we don't have an easy way to describe what changed
		// I am still thinking to a better solution.
		logrus.Warning("**Fallback** Please add a title to you configuration using the field 'title: <your pipeline>'")
		title = fmt.Sprintf("[updatecli][%s] Bump version to %s",
			config.Sources[config.Targets[ID].SourceID].Kind,
			fallback)
	}
	return title
}
