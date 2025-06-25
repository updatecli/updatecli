package config

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/updatecli/updatecli/pkg/core/cmdoptions"
	"github.com/updatecli/updatecli/pkg/core/pipeline/action"
	"github.com/updatecli/updatecli/pkg/core/pipeline/autodiscovery"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/core/version"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cueformat "cuelang.org/go/cue/format"
)

const (
	// LOCALSCMIDENTIFIER defines the scm id used to configure the local scm directory
	LOCALSCMIDENTIFIER string = "local"
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
	/*
		"name" defines a pipeline name

		example:
			* "name: 'deps: update nodejs version to latest stable'"

		remark:
			* using a short sentence describing the pipeline is a good way to name your pipeline.
			* using conventional commits convention is a good way to name your pipeline.
			* "name" is often used a default values for other configuration such as pullrequest title.
			* "name" shouldn't contain any dynamic information such as source output.
	*/
	Name string `yaml:",omitempty" jsonschema:"required"`
	/*
		"pipelineid" allows to identify a full pipeline run.

		example:
			* "pipelineid: nodejs/dependencies"
			* "pipelineid: gomod/github.com/updatecli/updatecli"
			* "pipelineid: autodiscovery/gomodules/minor"

		remark:
			* "pipelineid" is used to generate uniq branch name for target update relying on scm configuration.
			* The same "pipelineid" may be used by different Updatecli manifest" to ensure they are updated in the same workflow including pullrequest.
	*/
	PipelineID string `yaml:",omitempty"`
	/*
		"autodiscovery" defines the configuration to automatically discover new versions update.

		example:
		---
		autodiscovery:
			scmid: default
			actionid:  default
			groupby: all
			crawlers:
				golang/gomod:
					versionfilter:
					kind: semver
					pattern: patch
		---
	*/
	AutoDiscovery autodiscovery.Config `yaml:",omitempty"`
	/*
		"title" is deprecated, please use "name" instead.
	*/
	Title string `yaml:",omitempty" jsonschema:"-"`
	/*
		!Deprecated in favor of `actions`
	*/
	PullRequests map[string]action.Config `yaml:",omitempty" jsonschema:"-"`
	/*
		"actions" defines the list of action configurations which need to be managed.

		examples:
		---
		actions:
			default:
				kind: github/pullrequest
				scmid: default
				spec:
					automerge: true
					labels:
						- "dependencies"
		---
	*/
	Actions map[string]action.Config `yaml:",omitempty"`
	/*
		"scms" defines the list of repository configuration used to fetch content from.

		examples:
		---
		scms:
			default:
				kind: github
				spec:
					owner: "updatecli"
					repository: "updatecli"
					token: "${{ env "GITHUB_TOKEN" }}"
					branch: "main"
		---

	*/
	SCMs map[string]scm.Config `yaml:"scms,omitempty"`
	/*
		"sources" defines the list of Updatecli source definition.

		example:
		---
		sources:
			# Source to retrieve the latest version of nodejs
			nodejs:
				name: Get latest nodejs version
				kind: json
				spec:
					file: https://nodejs.org/dist/index.json
					key: .(lts!=false).version
		---
	*/
	Sources map[string]source.Config `yaml:",omitempty"`
	/*
		"conditions" defines the list of Updatecli condition definition.

		example:
		---
		conditions:
			container:
				name: Check if Updatecli container image for tag "v0.63.0" exists
				kind: dockerimage
				spec:
					image: "updatecli/updatecli:latest"
					tag: "v0.63.0"
		---
	*/
	Conditions map[string]condition.Config `yaml:",omitempty"`
	/*
		"targets" defines the list of Updatecli target definition.

		example:
		---
		targets:
		  	default:
		     	name: 'ci: update Golangci-lint version to {{ source "default" }}'
		     	kind: yaml
		     	spec:
		         	file: .github/workflows/go.yaml
		         	key: $.jobs.build.steps[2].with.version
		     	scmid: default
		     	sourceid: default
		---
	*/
	Targets map[string]target.Config `yaml:",omitempty"`
	/*
		"version" defines the minimum Updatecli version compatible with the manifest
	*/
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
	// DisableTemplating specifies if needs to be done
	DisableTemplating bool
}

// Reset reset configuration
func (config *Config) Reset() {
	*config = Config{
		gitHandler: &gitgeneric.GoGit{},
	}
}

// New reads an updatecli configuration file
func New(option Option) (configs []Config, err error) {

	_, basename := filepath.Split(option.ManifestFile)

	// We need to be sure to generate a file checksum before we inject
	// templates values as in some situation those values changes for each run
	fileChecksum, err := FileChecksum(option.ManifestFile)
	if err != nil {
		return configs, err
	}

	logrus.Infof("Loading Pipeline %q", option.ManifestFile)

	// Load updatecli manifest no matter the file extension
	c, err := os.Open(option.ManifestFile)

	if err != nil {
		return configs, err
	}

	defer c.Close()

	var templatedManifestContent []byte
	rawManifestContent, err := io.ReadAll(c)
	if err != nil {
		return configs, err
	}

	specs := []Spec{}

	isCue := false

	switch extension := filepath.Ext(basename); extension {
	case ".tpl", ".tmpl", ".yaml", ".yml", ".json":
		//
	case ".cue":
		if !cmdoptions.Experimental {
			return configs, fmt.Errorf("cuelang support is experimental, please use '--experimental' flag to enable it")
		}
		isCue = true

	default:
		logrus.Debugf("file extension '%s' not supported for file '%s'", extension, option.ManifestFile)
		return configs, ErrConfigFileTypeNotSupported
	}

	var cueManifest cue.Value
	if !option.DisableTemplating {
		// Try to template manifest no matter the extension
		// templated manifest must respect its extension before and after templating

		cwd, err := os.Getwd()
		if err != nil {
			return configs, err
		}
		fs := os.DirFS(cwd)

		t := Template{
			CfgFile:      option.ManifestFile,
			ValuesFiles:  option.ValuesFiles,
			SecretsFiles: option.SecretsFiles,
			fs:           fs,
		}

		if isCue {
			cueManifest, err = t.NewCueTemplate(rawManifestContent)
		} else {
			templatedManifestContent, err = t.NewStringTemplate(rawManifestContent)
		}
		if err != nil {
			logrus.Errorf("Error while templating %q:\n---\n%s\n---\n\t%s\n", option.ManifestFile, string(rawManifestContent), err.Error())
			return configs, err
		}

		if GolangTemplatingDiff {
			if isCue {
				// Reformat all the CUE sources so that they can be diffed.
				rawManifestContent, err = cueformat.Source(rawManifestContent, cueformat.Simplify())
				if err != nil {
					logrus.Errorf("Error formatting %q for diff: %s", option.ManifestFile, err.Error())
					return configs, err
				}

				node := cueManifest.Syntax(cue.Final(), cue.ErrorsAsValues(true))
				templatedManifestContent, err = cueformat.Node(node, cueformat.Simplify())
				if err != nil {
					logrus.Errorf("Error formatting %q (templated) for diff: %s", option.ManifestFile, err.Error())
					return configs, err
				}
			}

			diff := text.Diff("raw manifest", "templated manifest", string(rawManifestContent), string(templatedManifestContent))
			switch diff {
			case "":
				logrus.Debugln("no Golang templating detected")
			default:
				logrus.Debugf("Golang templating change detected:\n%s\n\n---\n", diff)
			}
		}
	} else if isCue {
		ctx := cuecontext.New()
		cueManifest = ctx.CompileBytes(rawManifestContent)
	}

	if isCue {
		var spec Spec
		err = cueManifest.Decode(&spec)
		if err != nil {
			return configs, err
		}
		specs = append(specs, spec)
	} else {
		err := unmarshalConfigSpec(templatedManifestContent, &specs)
		if err != nil {
			return configs, err
		}
	}

	configs = make([]Config, len(specs))
	for id := range specs {

		configs[id].Reset()
		configs[id].filename = option.ManifestFile
		configs[id].Spec = specs[id]
		// config.PipelineID is required for config.Validate()
		if len(configs[id].Spec.PipelineID) == 0 {
			logrus.Debugln("pipelineid undefined, we'll try to generate one")

			// If pipeline name is defined then we use it to generate a pipeline id
			// there is less change for the pipeline name to change.
			switch configs[id].Spec.Name {
			case "":
				logrus.Debugln("pipeline name undefined, we'll use the manifest file checksum")
				configs[id].Spec.PipelineID = fileChecksum
			default:
				logrus.Debugln("using pipeline name to generate the pipelineid")
				hash := sha256.New()
				hash.Write([]byte(configs[id].Spec.Name))
				configs[id].Spec.PipelineID = fmt.Sprintf("%x", hash.Sum(nil))
			}
		}

		// By default Set config.Version to the current updatecli version
		if len(configs[id].Spec.Version) == 0 {
			configs[id].Spec.Version = version.Version
		}

		// Ensure there is a local SCM defined as specified
		if err = configs[id].EnsureLocalScm(); err != nil {
			continue
		}

		/** Check for deprecated directives **/
		// pullrequests deprecated over actions
		if len(configs[id].Spec.PullRequests) > 0 {
			if len(configs[id].Spec.Actions) > 0 {
				err := fmt.Errorf("the `pullrequests` and `actions` keywords are mutually exclusive. Please use only `actions` as `pullrequests` is deprecated")
				logrus.Errorf("Skipping manifest %q:\n\t%s", option.ManifestFile, err.Error())
				continue
			}

			logrus.Warningf("The `pullrequests` keyword is deprecated in favor of `actions`, please update this manifest. Updatecli will continue the execution while trying to translate `pullrequests` to `actions`.")

			configs[id].Spec.Actions = configs[id].Spec.PullRequests
			configs[id].Spec.PullRequests = nil
		}

		err = configs[id].Validate()
		if err != nil {
			continue
		}

		if len(configs[id].Spec.Name) == 0 {
			configs[id].Spec.Name = strings.ToTitle(basename)
		}

		err = configs[id].Validate()
		if err != nil {
			continue
		}
	}

	return configs, err

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

	onDiskData, err := os.ReadFile(c.filename)
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
	c, err := yaml.Marshal(&config.Spec)
	if err != nil {
		return err
	}
	logrus.Infof("%s", string(c))

	return nil
}

func (config *Config) validateActions() error {
	for id, a := range config.Spec.Actions {
		if err := a.Validate(); err != nil {
			logrus.Errorf("bad parameters for action %q", id)
			return err
		}

		// Then validate that the action specifies an existing SCM
		if len(a.ScmID) > 0 {
			if _, ok := config.Spec.SCMs[a.ScmID]; !ok {
				logrus.Errorf("The action %q specifies a scm id %q which does not exist", id, a.ScmID)
				return ErrBadConfig
			}
		}

		// a.Validate may modify the object during validation
		// so we want to be sure that we save those modifications
		config.Spec.Actions[id] = a
	}
	return nil
}

func (config *Config) validateAutodiscovery() error {
	// Then validate that the action specifies an existing SCM
	if len(config.Spec.AutoDiscovery.ScmId) > 0 {
		if _, ok := config.Spec.SCMs[config.Spec.AutoDiscovery.ScmId]; !ok {
			logrus.Errorf("The autodiscovery specifies a scm id %q which does not exist",
				config.Spec.AutoDiscovery.ScmId)
			return ErrBadConfig
		}
	}

	if err := config.Spec.AutoDiscovery.GroupBy.Validate(); err != nil {
		return ErrBadConfig
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
				logrus.Errorf("the specified sourceid %q for condition[id] does not exist", t.SourceID)
				return ErrBadConfig
			}
		}

		if t.DisableConditions && len(t.DeprecatedConditionIDs) > 0 {
			logrus.Errorf("target %q has 'disableconditions' set to true and 'conditionids' defined (%v), it's not possible to disable conditions and define conditions at the same time", id, t.DeprecatedConditionIDs)
			return ErrBadConfig
		}

		undefinedConditions := []string{}
		for _, conditionID := range t.DeprecatedConditionIDs {
			if _, ok := config.Spec.Conditions[conditionID]; !ok {
				undefinedConditions = append(undefinedConditions, conditionID)
			}
		}

		if len(undefinedConditions) > 0 {
			logrus.Errorf("target %q has undefined conditionids: %v", id, undefinedConditions)
			return ErrBadConfig
		}

		// Only check/guess the sourceID if the user did not disable it (default is enabled)
		if !t.DisableSourceInput {
			// Try to guess SourceID
			if len(t.SourceID) == 0 && len(config.Spec.Sources) > 1 {

				logrus.Errorf("empty 'sourceid' for target %q", id)
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
				logrus.Errorf("the specified sourceid %q for condition[id] does not exist", c.SourceID)
				return ErrBadConfig
			}
		}
		// Only check/guess the sourceID if the user did not disable it (default is enabled)
		if !c.DisableSourceInput {
			// Try to guess SourceID
			if len(c.SourceID) == 0 && len(config.Spec.Sources) > 1 {
				logrus.Errorf("The condition %q has an empty 'sourceid' attribute.", id)
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

	if config.Spec.Title != "" {
		logrus.Warningf("title is deprecated, please use name instead")
	}

	err = config.validateConditions()
	if err != nil {
		errs = append(
			errs,
			fmt.Errorf("conditions validation error:\n%s", err))
	}

	err = config.validateActions()
	if err != nil {
		errs = append(
			errs,
			fmt.Errorf("actions validation error:\n%s", err))
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

	content, err := yaml.Marshal(config.Spec)
	if err != nil {
		return err
	}

	tmpl, err := template.New("cfg").Funcs(updatecliRuntimeFuncMap(data)).Parse(string(content))
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
