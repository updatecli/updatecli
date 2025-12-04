package plugin

import (
	"context"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"

	extism "github.com/extism/go-sdk"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
)

// generate_docer_source_spec is an Extism host function that generates a Docker source spec in YAML format.
// It reads input parameters from the plugin's memory stack, processes them to create the spec,
// and writes the resulting YAML back to the stack.
// The input is expected to be a JSON string containing "image" and "tag" fields.
// The output is a YAML representation of the Docker source spec.
func generate_docker_source_spec(ctx context.Context, p *extism.CurrentPlugin, stack []uint64) {
	input, err := p.ReadString(stack[0])
	if err != nil {
		fmt.Printf("Error reading input string: %v\n", err)
		stack[0] = 0
		return
	}

	inputData := struct {
		Image string `yaml:"image"`
		Tag   string `yaml:"tag"`
	}{}

	err = json.Unmarshal([]byte(input), &inputData)
	if err != nil {
		fmt.Printf("Error unmarshaling input JSON: %v\n", err)
		stack[0] = 0
		return
	}

	auths := map[string]docker.InlineKeyChain{}
	sourceSpec := dockerimage.NewDockerImageSpecFromImage(inputData.Image, inputData.Tag, auths)

	outputBytes, err := yaml.Marshal(sourceSpec)
	if err != nil {
		fmt.Printf("Error marshaling source spec to YAML: %v\n", err)
		stack[0], stack[1] = 0, 0
		return
	}

	output, err := p.WriteString(string(outputBytes))
	if err != nil {
		fmt.Printf("Error writing output bytes: %v\n", err)
		stack[0] = 0
		return
	}

	stack[0] = output
}
