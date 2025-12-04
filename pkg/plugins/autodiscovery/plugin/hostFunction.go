package plugin

import (
	"context"
	"encoding/json"

	extism "github.com/extism/go-sdk"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
)

// generate_docker_source_spec is an Extism host function that generates a Docker source spec in YAML format.
// It reads input parameters from the plugin's memory stack, processes them to create the spec,
// and writes the resulting YAML back to the stack.
// The input is expected to be a JSON string containing "image" and "tag" fields.
// The output is a json representation of the Docker source spec versionfilter and tagfilter.
// Such as:
//
//	{
//	  "versionfilter": {
//	    "kind": "semver",
//	    "pattern": "*"
//	  }
//	  "tagfilter": "^v[0-9]+\\.[0-9]+\\.[0-9]+$"
//	}
func generate_docker_source_spec(ctx context.Context, p *extism.CurrentPlugin, stack []uint64) {
	input, err := p.ReadString(stack[0])
	if err != nil {
		stack[0] = 0
		return
	}

	inputData := struct {
		Image string `json:"image"`
		Tag   string `json:"tag"`
	}{}

	err = json.Unmarshal([]byte(input), &inputData)
	if err != nil {
		stack[0] = 0
		return
	}

	auths := map[string]docker.InlineKeyChain{}
	sourceSpec := dockerimage.NewDockerImageSpecFromImage(inputData.Image, inputData.Tag, auths)

	if sourceSpec == nil {
		stack[0] = 0
		return
	}

	outputBytes, err := json.Marshal(sourceSpec)
	if err != nil {
		stack[0], stack[1] = 0, 0
		return
	}

	output, err := p.WriteString(string(outputBytes))
	if err != nil {
		stack[0] = 0
		return
	}

	stack[0] = output
}
