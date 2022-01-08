package dockerfile

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/plugins/dockerfile/mobyparser"
	"github.com/updatecli/updatecli/pkg/plugins/dockerfile/simpletextparser"
	"github.com/updatecli/updatecli/pkg/plugins/dockerfile/types"
)

// Spec defines a specification for a "dockerfile" resource
// parsed from an updatecli manifest file
type Spec struct {
	File        string            `yaml:"file"`
	Instruction types.Instruction `yaml:"instruction"`
	Value       string            `yaml:"value"`
	DryRun      bool
}

// Dockerfile defines a resource of kind "dockerfile"
type Dockerfile struct {
	parser   types.DockerfileParser
	messages []string
	spec     Spec
}

// New returns a reference to a newly initialized Dockerfile object from a Spec
// or an error if the provided Spec triggers a validation error.
func New(newSpec Spec) (*Dockerfile, error) {
	newParser, err := getParser(newSpec)
	if err != nil {
		return nil, err
	}
	newResource := &Dockerfile{
		spec:   newSpec,
		parser: newParser,
	}

	return newResource, nil
}

func getParser(spec Spec) (types.DockerfileParser, error) {
	var instruction interface{}

	instruction = spec.Instruction
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
