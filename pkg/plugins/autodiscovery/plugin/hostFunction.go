package plugin

import (
	"context"
	"encoding/json"

	extism "github.com/extism/go-sdk"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
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
	logrus.Debugf("Calling generate_docker_source_spec wasm host function")
	input, err := p.ReadString(stack[0])
	if err != nil {
		logrus.Errorf("reading input string: %s", err)
		stack[0] = 0
		return
	}

	inputData := struct {
		Image string `json:"image"`
		Tag   string `json:"tag"`
	}{}

	err = json.Unmarshal([]byte(input), &inputData)
	if err != nil {
		logrus.Errorf("unmarshaling input string: %s", err)
		stack[0] = 0
		return
	}

	auths := map[string]docker.InlineKeyChain{}
	sourceSpec := dockerimage.NewDockerImageSpecFromImage(inputData.Image, inputData.Tag, auths)

	if sourceSpec == nil {
		logrus.Errorf("no source spec generated for image %q and tag %q", inputData.Image, inputData.Tag)
		stack[0] = 0
		return
	}

	outputBytes, err := json.Marshal(sourceSpec)
	if err != nil {
		logrus.Errorf("marshaling output string: %s", err)
		stack[0], stack[1] = 0, 0
		return
	}

	output, err := p.WriteString(string(outputBytes))
	if err != nil {
		logrus.Errorf("writing output string: %s", err)
		stack[0] = 0
		return
	}

	stack[0] = output
}

// versionfilter_greater_than_pattern is an Extism host function that modifies a version filter pattern
// to be greater than a specified pattern.
// It reads input parameters from the plugin's memory stack, processes them to create the new pattern,
// and writes the resulting pattern back to the stack.
// The input is expected to be a JSON string containing "versionfilter" and "pattern" fields.
// The output is a JSON string containing the modified "pattern".
func versionfilter_greater_than_pattern(ctx context.Context, p *extism.CurrentPlugin, stack []uint64) {
	logrus.Debugf("Calling versionfilter_greater_than_pattern wasm host function")
	input, err := p.ReadString(stack[0])

	if err != nil {
		logrus.Errorf("reading input string: %s", err)
		stack[0] = 0
		return
	}

	inputData := struct {
		VersionFilter version.Filter `json:"versionfilter"`
		Pattern       string         `json:"pattern"`
	}{}

	err = json.Unmarshal([]byte(input), &inputData)
	if err != nil {
		logrus.Errorf("unmarshaling input string: %s", err)
		stack[0] = 0
		return
	}

	newPattern, err := inputData.VersionFilter.GreaterThanPattern(
		inputData.Pattern,
	)
	if err != nil {
		logrus.Errorf("generating greater than pattern: %s", err)
		stack[0] = 0
		return
	}

	outputData := struct {
		Pattern string `json:"pattern"`
	}{
		Pattern: newPattern,
	}

	outputBytes, err := json.Marshal(outputData)
	if err != nil {
		logrus.Errorf("marshaling output string: %s", err)
		stack[0] = 0
		return
	}

	output, err := p.WriteString(string(outputBytes))
	if err != nil {
		logrus.Errorf("writing output string: %s", err)
		stack[0] = 0
		return
	}

	stack[0] = output
}
