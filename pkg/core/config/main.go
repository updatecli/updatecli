package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	autodiscovery "github.com/updatecli/updatecli/pkg/core/pipeline/autodiscovery/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/pullrequest"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/version"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

const (
	// LOCALSCMIDENTIFIER defines the scm id used to configure the local scm directory
	LOCALSCMIDENTIFIER string = "local"
	// DefaultConfigFilename defines the default updatecli configuration filename
	DefaultConfigFilename string = "updatecli.yaml"
	// DefaultConfigDirname defines the default updatecli manifest directory
	DefaultConfigDirname string = "updatecli.d"
)

// Config contains cli configuration
type Config struct {
	// filename contains the updatecli manifest filename
	filename string
	// Spec describe an updatecli manifest
	Spec Spec
	// gitHandler holds a git client implementation to manipulate git SCMs
	gitHandler gitgeneric.GitHandler
}

// Spec contains pipeline configuration
type Spec struct {
	// Name defines a pipeline name
	Name string `yaml:",omitempty" jsonschema:"required"`
	// PipelineID allows to identify a full pipeline run, this value is propagated into each target if not defined at that level
	PipelineID string `yaml:",omitempty"`
	// AutoDiscovery defines parameters to the autodiscover feature
	AutoDiscovery autodiscovery.Config `yaml:",omitempty"`
	// Title is used for the full pipeline
	Title string `yaml:",omitempty"`
	// PullRequests defines the list of Pull Request configuration which need to be managed
	PullRequests map[string]pullrequest.Config `yaml:",omitempty"`
	// SCMs defines the list of repository configuration used to fetch content from.
	SCMs map[string]scm.Config `yaml:"scms,omitempty"`
	// Sources defines the list of source configuration
	Sources map[string]source.Config `yaml:",omitempty"`
	// Conditions defines the list of condition configuration
	Conditions map[string]condition.Config `yaml:",omitempty"`
	// Targets defines the list of target configuration
	Targets map[string]target.Config `yaml:",omitempty"`
	// Version specifies the mininum updatecli version compatible with the manifest
	Version string `yaml:",omitempty"`
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
	*config = Config{
		gitHandler: gitgeneric.GoGit{},
	}
}

// New reads an updatecli configuration file
func New(option Option) (config Config, err error) {

	config.Reset()

	config.filename = option.ManifestFile

	dirname, basename := filepath.Split(option.ManifestFile)

	// We need to be sure to generate a file checksum before we inject
	// templates values as in some situation those values changes for each run
	pipelineID, err := FileChecksum(option.ManifestFile)
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

		err = yaml.Unmarshal(content, &config.Spec)
		if err != nil {
			return config, err
		}

	default:
		logrus.Debugf("file extension '%s' not supported for file '%s'", extension, config.filename)
		return config, ErrConfigFileTypeNotSupported
	}

	// config.PipelineID is required for config.Validate()
	if len(config.Spec.PipelineID) == 0 {
		config.Spec.PipelineID = pipelineID
	}

	// By default Set config.Version to the current updatecli version
	if len(config.Spec.Version) == 0 {
		config.Spec.Version = version.Version
	}

	// Ensure there is a local SCM defined as specified
	if err = config.EnsureLocalScm(); err != nil {
		return config, err
	}

	err = config.Validate()
	if err != nil {
		return config, err
	}

	if len(config.Spec.Name) == 0 {
		config.Spec.Name = strings.ToTitle(basename)
	}

	err = config.Validate()

	return config, err

}

// IsManifestDifferentThanOnDisk checks if an Updatecli manifest in memory is the same than the one on disk
func (c *Config) IsManifestDifferentThanOnDisk() (bool, error) {

	buf := bytes.NewBufferString("")

	encoder := yaml.NewEncoder(buf)

	defer func() {
		err := encoder.Close()
		if err != nil {
			logrus.Errorln(err)
		}
	}()

	encoder.SetIndent(YAMLSetIdent)

	err := encoder.Encode(c.Spec)

	if err != nil {
		return false, err
	}

	data := buf.Bytes()

	onDiskData, err := ioutil.ReadFile(c.filename)
	if err != nil {
		return false, err
	}

	if string(onDiskData) == string(data) {
		logrus.Infof("%s No Updatecli manifest change required", result.SUCCESS)
		return false, nil
	}

	edits := myers.ComputeEdits(span.URIFromPath(c.filename), string(onDiskData), string(data))
	diff := fmt.Sprint(gotextdiff.ToUnified(c.filename+"(old)", c.filename+"(updated)", string(onDiskData), edits))

	logrus.Infof("%s Updatecli manifest change required\n%s", result.ATTENTION, diff)

	return true, nil

}

// SaveOnDisk saves an updatecli manifest to disk
func (c *Config) SaveOnDisk() error {

	file, err := os.OpenFile(c.filename, os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	defer func() {
		err = file.Close()
		if err != nil {
			logrus.Errorln(err)
		}
	}()

	encoder := yaml.NewEncoder(file)

	defer func() {
		err = encoder.Close()
		if err != nil {
			logrus.Errorln(err)
		}
	}()

	// Must be set to 4 so we align with the default behavior when
	// we marshaled the inmemory yaml to compare with the one from disk
	encoder.SetIndent(YAMLSetIdent)
	err = encoder.Encode(c.Spec)

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
	for id, p := range config.Spec.PullRequests {
		if err := p.Validate(); err != nil {
			logrus.Errorf("bad parameters for pullrequest %q", id)
			return err
		}

		// Then validate that the pullrequest specifies an existing SCM
		if len(p.ScmID) > 0 {
			if _, ok := config.Spec.SCMs[p.ScmID]; !ok {
				logrus.Errorf("The pullrequest %q specifies a scm id %q which does not exist", id, p.ScmID)
				return ErrBadConfig
			}
		}

		// p.Validate may modify the object during validation
		// so we want to be sure that we save those modifications
		config.Spec.PullRequests[id] = p
	}
	return nil
}

func (config *Config) validateAutodiscovery() error {
	// Then validate that the pullrequest specifies an existing SCM
	if len(config.Spec.AutoDiscovery.ScmId) > 0 {
		if _, ok := config.Spec.SCMs[config.Spec.AutoDiscovery.ScmId]; !ok {
			logrus.Errorf("The autodiscovery specifies a scm id %q which does not exist",
				config.Spec.AutoDiscovery.ScmId)
			return ErrBadConfig
		}
	}
	return nil
}

func (config *Config) validateSources() error {
	for id, s := range config.Spec.Sources {
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
				if _, ok := config.Spec.SCMs["source_"+id]; !ok {
					for kind, spec := range s.Scm {
						if config.Spec.SCMs == nil {
							config.Spec.SCMs = make(map[string]scm.Config, 1)
						}
						config.Spec.SCMs["source_"+id] = scm.Config{
							Kind: kind,
							Spec: spec}
					}
				}
				s.SCMID = "source_" + id
			} else {
				logrus.Warning("source.SCMID is also defined, ignoring source.Scm")
			}
			s.Scm = map[string]interface{}{}
			config.Spec.Sources[id] = s
		}
		// s.Validate may modify the object during validation
		// so we want to be sure that we save those modifications
		config.Spec.Sources[id] = s
	}
	return nil

}

func (config *Config) validateSCMs() error {
	for id, scmConfig := range config.Spec.SCMs {
		if err := scmConfig.Validate(); err != nil {
			logrus.Errorf("bad parameter(s) for scm %q", id)
			return err
		}
	}
	return nil
}

func (config *Config) validateTargets() error {

	for id, t := range config.Spec.Targets {

		err := t.Validate()
		if err != nil {
			logrus.Errorf("bad parameters for target %q", id)
			return ErrBadConfig
		}

		if len(t.SourceID) > 0 {
			if _, ok := config.Spec.Sources[t.SourceID]; !ok {
				logrus.Errorf("the specified SourceID %q for condition[id] does not exist", t.SourceID)
				return ErrBadConfig
			}
		}

		// Only check/guess the sourceID if the user did not disable it (default is enabled)
		if !t.DisableSourceInput {
			// Try to guess SourceID
			if len(t.SourceID) == 0 && len(config.Spec.Sources) > 1 {

				logrus.Errorf("empty 'sourceID' for target %q", id)
				return ErrBadConfig
			} else if len(t.SourceID) == 0 && len(config.Spec.Sources) == 1 {
				for id := range config.Spec.Sources {
					t.SourceID = id
				}
			}
		}

		if IsTemplatedString(id) {
			logrus.Errorf("target key %q contains forbidden go template instruction", id)
			return ErrNotAllowedTemplatedKey
		}

		// t.Validate may modify the object during validation
		// so we want to be sure that we save those modifications
		config.Spec.Targets[id] = t

		// Temporary code until we fully remove the old way to configure scm
		// Introduce by https://github.com/updatecli/updatecli/issues/260
		//if t.Scm != nil {
		if len(t.Scm) > 0 {
			err := generateScmFromLegacyTarget(id, config)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func (config *Config) validateConditions() error {
	for id, c := range config.Spec.Conditions {
		err := c.Validate()
		if err != nil {
			logrus.Errorf("bad parameters for condition %q", id)
			return ErrBadConfig
		}

		if len(c.SourceID) > 0 {
			if _, ok := config.Spec.Sources[c.SourceID]; !ok {
				logrus.Errorf("the specified SourceID %q for condition[id] does not exist", c.SourceID)
				return ErrBadConfig
			}
		}
		// Only check/guess the sourceID if the user did not disable it (default is enabled)
		if !c.DisableSourceInput {
			// Try to guess SourceID
			if len(c.SourceID) == 0 && len(config.Spec.Sources) > 1 {
				logrus.Errorf("The condition %q has an empty 'sourceID' attribute.", id)
				return ErrBadConfig
			} else if len(c.SourceID) == 0 && len(config.Spec.Sources) == 1 {
				for id := range config.Spec.Sources {
					c.SourceID = id
				}
			}
		}

		if IsTemplatedString(id) {
			logrus.Errorf("condition key %q contains forbidden go template instruction", id)
			return ErrNotAllowedTemplatedKey
		}

		config.Spec.Conditions[id] = c

		// Temporary code until we fully remove the old way to configure scm
		// Introduce by https://github.com/updatecli/updatecli/issues/260
		//if c.Scm != nil {
		if len(c.Scm) > 0 {
			generateScmFromLegacyCondition(id, config)
		}

	}
	return nil
}

// Validate run various validation test on the configuration and update fields if necessary
func (config *Config) Validate() error {

	var errs []error
	err := config.ValidateManifestCompatibility()
	if err != nil {
		errs = append(
			errs,
			fmt.Errorf("updatecli version compatibility error:\n%s", err))
	}

	err = config.validateConditions()
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

	err = config.validateAutodiscovery()
	if err != nil {
		errs = append(
			errs,
			fmt.Errorf("autodiscovery validation error:\n%s", err))
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

func (config *Config) ValidateManifestCompatibility() error {
	isCompatibleUpdatecliVersion, err := version.IsGreaterThan(
		version.Version,
		config.Spec.Version)

	if err != nil {
		return fmt.Errorf("pipeline %q - %q", config.Spec.Name, err)
	}

	// Ensure that the current updatecli version is compatible with the manifest
	if !isCompatibleUpdatecliVersion {
		return fmt.Errorf("pipeline %q requires Updatecli version greater than %q, skipping", config.Spec.Name, config.Spec.Version)
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

	content, err := yaml.Marshal(config.Spec)
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

	err = yaml.Unmarshal(b.Bytes(), &config.Spec)
	if err != nil {
		return err
	}

	err = config.Validate()
	if err != nil {
		return err
	}

	return err
}
