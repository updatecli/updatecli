package maven

import (
	"fmt"
	"testing"

	discoveryConfig "github.com/updatecli/updatecli/pkg/core/pipeline/autodiscovery/config"
	goyaml "gopkg.in/yaml.v3"
)

var (
	expectedPipelines map[string]string = map[string]string{}
)

func TestDiscoverManifests(t *testing.T) {

	spec := Spec{
		RootDir: "testdata",
	}

	maven, err := New(spec, "")

	if err != nil {
		t.Errorf("%v", err)
	}

	pipelines, err := maven.DiscoverManifests(discoveryConfig.Input{})

	if err != nil {
		t.Errorf("%v", err)
	}

	for _, pipeline := range pipelines {

		output, err := goyaml.Marshal(pipeline)

		if err != nil {
			t.Errorf("%v", err)
		}

		if string(output) != expectedPipelines[pipeline.Name] {
			fmt.Println("Wrong result")
			fmt.Printf("Expected:\n>>>\n%v\n>>>\n\n", expectedPipelines[pipeline.Name])
			fmt.Printf("Got:\n<<<\n%v\n<<<\n", string(output))
			t.Fail()
		}
	}
}
