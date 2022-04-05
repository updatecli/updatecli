package dockerfile

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile/mobyparser"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile/simpletextparser"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile/types"
)

// Spec defines a specification for a "dockerfile" resource
// parsed from an updatecli manifest file
type Spec struct {
	// File specifies the dockerimage file such as Dockerfile
	File string `yaml:"file"`
	// Instruction specifies a DockerImage instruction such as ENV
	Instruction types.Instruction `yaml:"instruction"`
	// Value specifies the value for a specified Dockerfile instruction.
	Value string `yaml:"value"`
}

// Dockerfile defines a resource of kind "dockerfile"
type Dockerfile struct {
	parser           types.DockerfileParser
	spec             Spec
	contentRetriever text.TextRetriever
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

	newResource := &Dockerfile{
		spec:             newSpec,
		parser:           newParser,
		contentRetriever: &text.Text{},
	}

	return newResource, nil
}

func getParser(spec Spec) (types.DockerfileParser, error) {
	instruction := spec.Instruction
	switch i := instruction.(type) {
	default:
		return nil, fmt.Errorf("Parsing Error: cannot determine instruction: %v.", i)
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
				return nil, fmt.Errorf("Parsing Error: cannot determine instruction: %v.", i)
			}
			parsedInstruction[k] = stringValue
		}
		return simpletextparser.NewSimpleTextDockerfileParser(parsedInstruction)
	}
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (df *Dockerfile) Changelog() string {
	return ""
}
