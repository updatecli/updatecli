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

/*
"dockerfile" defines the specification for manipulating Dockerfile instructions.
It can be used as a "source", a "condition", or a "target".
*/
type Spec struct {
	// "file" defines the path of the Dockerfile.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "file" and "files" are mutually exclusive.
	//
	// example:
	//   * file: Dockerfile
	//
	File string `yaml:",omitempty"`
	// "files" defines the list of Dockerfile paths.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "file" and "files" are mutually exclusive.
	//   * a source accepts only one file.
	//
	Files []string `yaml:",omitempty"`
	// "instruction" defines the Dockerfile instruction to manipulate.
	//
	// It is either a string or a map with the keys "keyword" and "matcher".
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * in the map form, "keyword" accepts "FROM", "ARG", "ENV" or "LABEL", case insensitive.
	//   * in the map form, "matcher" selects the instruction, such as the image name for "FROM"
	//     or the variable name for "ARG" and "ENV".
	//   * in the map form, the optional key "ignoreUnsetValue" set to true ignores instructions without a value.
	//   * in the map form, a condition only checks that a matching instruction exists.
	//   * the string form does not support a source, which returns an empty value.
	//
	// example:
	// ```
	//   instruction:
	//     keyword: "FROM"
	//     matcher: "alpine"
	// ```
	//
	Instruction types.Instruction `yaml:"instruction,omitempty"`
	// "value" defines the expected value of the Dockerfile instruction.
	//
	// compatible:
	//   * condition
	//
	// remark:
	//   * it is only used when "instruction" is a string.
	//   * the condition fails only when the instruction is missing.
	//     A different value is reported in the logs but does not fail the condition.
	//   * a target always uses the output of the associated source.
	//
	Value string `yaml:"value,omitempty"`
	// "stage" defines the Dockerfile stage to consider.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * it is only used when "instruction" is a map.
	//   * when unset in a source, the last stage is used.
	//   * when unset in a condition or a target, every stage is used.
	//
	// example:
	//   * stage: builder
	//
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
			switch val := v.(type) {
			case string:
				parsedInstruction[k] = val
			case bool:
				parsedInstruction[k] = fmt.Sprintf("%t", val)
			default:
				return nil, fmt.Errorf("parsing error: cannot determine instruction: %v", i)
			}
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
