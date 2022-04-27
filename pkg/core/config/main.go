package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
	// Name defines a pipeline name
	Name string
	// PipelineID allows to identify a full pipeline run, this value is propagated into each target if not defined at that level
	PipelineID string
	// Title is used for the full pipeline
	Title string
	// PullRequests defines the list of Pull Request configuration which need to be managed
	PullRequests map[string]pullrequest.Config
	// SCMs defines the list of repository configuration used to fetch content from.
	SCMs map[string]scm.Config `yaml:"scms"`
	// Sources defines the list of source configuration
	Sources map[string]source.Config
	// Conditions defines the list of condition configuration
	Conditions map[string]condition.Config
	// Targets defines the list of target configuration
	Targets map[string]target.Config
}

// Option contains configuration options such as filepath located on disk,etc.
type Option struct {
	// ManifestFile contains the updatecli manifest full file path
	ManifestFile string
	// ValuesFiles contains the list of updatecli values full file path
	ValuesFiles []string
	// SecretsFiles contains the list of updatecli sops secrets full file path
	SecretsFiles []string
	// DisableTemplating specify if needs to be done
	DisableTemplating bool
}

// Reset reset configuration
func (config *Config) Reset() {
	*config = Config{}
}

// New reads an updatecli configuration file
func New(option Option) (config Config, err error) {

	config.Reset()

	dirname, basename := filepath.Split(option.ManifestFile)

	// We need to be sure to generate a file checksum before we inject
	// templates values as in some situation those values changes for each run
	pipelineID, err := Checksum(option.ManifestFile)
	if err != nil {
		return config, err
	}

	logrus.Infof("Loading Pipeline %q", option.ManifestFile)

	// Load updatecli manifest no matter the file extension
	c, err := os.Open(option.ManifestFile)

	if err != nil {
		return config, err
	}

	defer c.Close()

	content, err := ioutil.ReadAll(c)
	if err != nil {
		return config, err
	}

	if !option.DisableTemplating {
		// Try to template manifest no matter the extension
		// templated manifest must respect its extension before and after templating

		t := Template{
			CfgFile:      filepath.Join(dirname, basename),
			ValuesFiles:  option.ValuesFiles,
			SecretsFiles: option.SecretsFiles,
		}

		content, err = t.New(content)
		if err != nil {
			return config, err
		}

	}
	switch extension := filepath.Ext(basename); extension {
	case ".tpl", ".tmpl", ".yaml", ".yml":

		err = yaml.Unmarshal(content, &config)
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

// SaveOnDisk save an updatecli manifest on disk
func (c *Config) SaveOnDisk(filename string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	defer func() {
		err = file.Close()
		if err != nil {
			logrus.Errorln(err)
		}

	}()

	_, err = file.WriteString(string(data))
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

func (config *Config) validatePullRequests() error {
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
		// p.Validate may modify the object during validation
		// so we want to be sure that we save those modifications
		config.PullRequests[id] = p
	}
	return nil
}

func (config *Config) validateSources() error {
	for id, s := range config.Sources {
		err := s.Validate()
		if err != nil {
			logrus.Errorf("bad parameters for source %q", id)
			return ErrBadConfig
		}

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
		// s.Validate may modify the object during validation
		// so we want to be sure that we save those modifications
		config.Sources[id] = s
	}
	return nil

}

func (config *Config) validateSCMs() error {
	for id, scm := range config.SCMs {
		if err := scm.Validate(); err != nil {
			logrus.Errorf("bad parameter(s) for scm %q", id)
			return err
		}
		// scm.Validate may modify the object during validation
		// so we want to be sure that we save those modification
		config.SCMs[id] = scm
	}
	return nil
}

func (config *Config) validateTargets() error {

	for id, t := range config.Targets {

		err := t.Validate()
		if err != nil {
			logrus.Errorf("bad parameters for target %q", id)
			return ErrBadConfig
		}

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

		// t.Validate may modify the object during validation
		// so we want to be sure that we save those modifications
		config.Targets[id] = t
	}
	return nil
}

func (config *Config) validateConditions() error {
	for id, c := range config.Conditions {
		err := c.Validate()
		if err != nil {
			logrus.Errorf("bad parameters for condition %q", id)
			return ErrBadConfig
		}

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
	return nil
}

// Validate run various validation test on the configuration and update fields if necessary
func (config *Config) Validate() error {

	var errs []error

	err := config.validateConditions()
	if err != nil {
		errs = append(
			errs,
			fmt.Errorf("conditions validation error:\n%s", err))
	}

	err = config.validatePullRequests()
	if err != nil {
		errs = append(
			errs,
			fmt.Errorf("pullrequests validation error:\n%s", err))
	}

	err = config.validateSCMs()
	if err != nil {
		errs = append(
			errs,
			fmt.Errorf("scms validation error:\n%s", err))
	}

	err = config.validateSources()
	if err != nil {
		errs = append(
			errs,
			fmt.Errorf("sources validation error:\n%s", err))
	}

	err = config.validateTargets()
	if err != nil {
		errs = append(
			errs,
			fmt.Errorf("targets validation error:\n%s", err))
	}

	// Concatenate error message
	if len(errs) > 0 {
		err = errs[0]
		for i := range errs {
			if i > 1 {
				err = fmt.Errorf("%s\n%s", err, errs[i])
			}
		}
		return err
	}

	return nil
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
