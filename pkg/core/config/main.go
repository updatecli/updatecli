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
	"github.com/updatecli/updatecli/pkg/core/pipeline/pullrequest"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/core/result"
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
)

// Config contains cli configuration
type Config struct {
	Name         string
	PipelineID   string                        // PipelineID allows to identify a full pipeline run, this value is propagated into each target if not defined at that level
	Title        string                        // Title is used for the full pipeline
	PullRequests map[string]pullrequest.Config // PullRequests defines the list of Pull Request configuration which need to be managed
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

	logrus.Infof("Loading Pipeline %q", cfgFile)

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
	for id, scm := range config.SCMs {
		if err := scm.Validate(); err != nil {
			logrus.Errorf("bad parameter(s) for scmIDs %q", id)
			return err
		}
	}

	for id, p := range config.PullRequests {
		if err := p.Validate(); err != nil {
			logrus.Errorf("bad parameters for pullrequest %q", id)
			return err
		}

		// Then validate that the pullrequest specifies an existing SCM
		if len(p.ScmID) > 0 {
			if _, ok := config.SCMs[p.ScmID]; !ok {
				logrus.Errorf("The pullrequest %q specifies a scm id %q which does not exist", id, p.ScmID)
				return ErrBadConfig
			}
		}

		// Validate references to other configuration objects
		for _, target := range p.Targets {
			if _, ok := config.Targets[target]; !ok {
				logrus.Errorf("the specified target %q for the pull request %q does not exist", target, id)
				return ErrBadConfig
			}
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
			logrus.Warningf("The directive 'scm' for the source[%q] is now deprecated. Please use the new top level scms syntax", id)
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
		if len(c.SourceID) > 0 {
			if _, ok := config.Sources[c.SourceID]; !ok {
				logrus.Errorf("the specified SourceID %q for condition[id] does not exist", c.SourceID)
				return ErrBadConfig
			}
		}
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

		// Temporary code until we fully remove the old way to configure scm
		// Introduce by https://github.com/updatecli/updatecli/issues/260
		//if c.Scm != nil {
		if len(c.Scm) > 0 {
			generateScmFromLegacyCondition(id, c, config)
		}

		config.Conditions[id] = c
	}

	for id, t := range config.Targets {
		if len(t.PipelineID) == 0 {
			t.PipelineID = config.PipelineID
		}
		if len(t.SourceID) > 0 {
			if _, ok := config.Sources[t.SourceID]; !ok {
				logrus.Errorf("the specified SourceID %q for condition[id] does not exist", t.SourceID)
				return ErrBadConfig
			}
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

		// Temporary code until we fully remove the old way to configure scm
		// Introduce by https://github.com/updatecli/updatecli/issues/260
		//if t.Scm != nil {
		if len(t.Scm) > 0 {
			err := generateScmFromLegacyTarget(id, t, config)
			if err != nil {
				return err
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
				It returns {{ source "<key>" }} if a key exist but still set to zero value,
				then we assume that the value will be set later in the run.
				Otherwise it returns the value.
				This func is design to constantly reevaluate if a configuration changed
			*/

			sourceResult, err := getFieldValueByQuery(data, []string{"Sources", s, "Result"})
			if err != nil {
				return "", err
			}

			switch sourceResult {
			case result.SUCCESS:
				return getFieldValueByQuery(data, []string{"Sources", s, "Output"})
			case result.FAILURE:
				return "", fmt.Errorf("parent source %q failed", s)
			// If the result of the parent source execution is not SUCCESS or FAILURE, then it means it was either skipped or not already run.
			// In this case, the function is return "as it" (literrally) to allow retry later (on a second configuration iteration)
			default:
				return fmt.Sprintf("{{ source %q }}", s), nil
			}
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
