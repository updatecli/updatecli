package dockerfile

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/plugins/dockerfile/mobyparser"
	"github.com/updatecli/updatecli/pkg/plugins/dockerfile/simpletextparser"
	"github.com/updatecli/updatecli/pkg/plugins/dockerfile/types"
)

// Dockerfile is struct that holds parameters for a "Dockerfile" kind from updatecli's parsing (YAML, command and flags)
type Dockerfile struct {
	File        string            `yaml:"file"`
	Instruction types.Instruction `yaml:"instruction"`
	Value       string            `yaml:"value"`
	parser      types.DockerfileParser
	DryRun      bool
	messages    []string
}

func (d *Dockerfile) SetParser() error {
	var i interface{}
	var err error

	i = d.Instruction
	switch i := i.(type) {
	default:
		return fmt.Errorf("Parsing Error: cannot determine instruction: %v.", i)
	case string:
		d.parser = mobyparser.MobyParser{
			Instruction: i,
			Value:       d.Value,
		}
		return nil
	case map[string]string:
		d.parser, err = simpletextparser.NewSimpleTextDockerfileParser(i)
		return err
	case map[string]interface{}:
		// If the YAML parser is typing the map values weakly
		// Then a new map with the correct type has to be constructed by copy
		parsedInstruction := make(map[string]string)
		for k, v := range i {
			stringValue, ok := v.(string)
			if !ok {
				return fmt.Errorf("Parsing Error: cannot determine instruction: %v.", i)
			}
			parsedInstruction[k] = stringValue
		}
		d.parser, err = simpletextparser.NewSimpleTextDockerfileParser(parsedInstruction)
		return err
	}
}
