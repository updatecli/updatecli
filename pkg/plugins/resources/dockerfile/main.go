package dockerfile

import (
	"fmt"

	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile/mobyparser"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile/simpletextparser"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile/types"
)

// Spec defines a specification for a "dockerfile" resource
// parsed from an updatecli manifest file
type Spec struct {
	// File specifies the dockerimage file path to use and is incompatible with Files
	File string `yaml:",omitempty"`
	// Files specifies the dockerimage file path(s) to use and is incompatible with File
	Files []string `yaml:",omitempty"`
	// Instruction specifies a DockerImage instruction such as ENV
	Instruction types.Instruction `yaml:"instruction,omitempty"`
	// Value specifies the value for a specified Dockerfile instruction.
	Value string `yaml:"value,omitempty"`
	// Stage can be used to further refined the scope
	// For Sources:
	// - If not defined, the last stage will be considered
	// For Condition and Targets:
	// - If not defined, all stages will be considered
	Stage string `yaml:"stage,omitempty"`
}

// Dockerfile defines a resource of kind "dockerfile"
type Dockerfile struct {
	parser           types.DockerfileParser
	spec             Spec
	contentRetriever text.TextRetriever
	files            []string
}

// New returns a reference to a newly initialized Dockerfile object from a Spec
// or an error if the provided Spec triggers a validation error.
func New(spec interface{}) (*Dockerfile, error) {
	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	newParser, err := getParser(newSpec)
	if err != nil {
		return nil, err
	}

	fileList := newSpec.Files
	if newSpec.File != "" {
		if len(newSpec.Files) > 0 {
			return nil, fmt.Errorf("parsing error: spec.file and spec.files are mutually exclusive")
		}
		fileList = append(fileList, newSpec.File)
	}

	newResource := &Dockerfile{
		spec:             newSpec,
		parser:           newParser,
		contentRetriever: &text.Text{},
		files:            fileList,
	}

	return newResource, nil
}

func getParser(spec Spec) (types.DockerfileParser, error) {
	instruction := spec.Instruction
	switch i := instruction.(type) {
	default:
		return nil, fmt.Errorf("parsing error: cannot determine instruction: %v", i)
	case string:
		return mobyparser.MobyParser{
			Instruction: i,
			Value:       spec.Value,
		}, nil
	case map[string]string:
		return simpletextparser.NewSimpleTextDockerfileParser(i)
	case map[string]interface{}:
		// If the YAML parser is typing the map values weakly
		// Then a new map with the correct type has to be constructed by copy
		parsedInstruction := make(map[string]string)
		for k, v := range i {
			stringValue, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("parsing error: cannot determine instruction: %v", i)
			}
			parsedInstruction[k] = stringValue
		}
		return simpletextparser.NewSimpleTextDockerfileParser(parsedInstruction)
	}
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (df *Dockerfile) Changelog(from, to string) *result.Changelogs {
	return nil
}

// ReportConfig returns a cleaned version of the configuration
// to identify the resource without any sensitive information or context specific data.
func (df *Dockerfile) ReportConfig() interface{} {
	return Spec{
		File:        df.spec.File,
		Files:       df.spec.Files,
		Instruction: df.spec.Instruction,
		Value:       df.spec.Value,
		Stage:       df.spec.Stage,
	}
}
