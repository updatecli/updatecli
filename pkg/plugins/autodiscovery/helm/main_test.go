package helm

import (
	"fmt"
	"testing"

	goyaml "gopkg.in/yaml.v3"
)

func TestDiscoverManifests(t *testing.T) {

	spec := Spec{
		RootDir: "/home/olblak/Projects/Epinio/helm-charts",
	}

	helm, err := New(spec, "")

	if err != nil {
		t.Errorf("%v", err)
	}

	pipelines, err := helm.DiscoverManifests(nil)

	if err != nil {
		t.Errorf("%v", err)
	}

	for _, pipeline := range pipelines {

		output, err := goyaml.Marshal(pipeline)

		if err != nil {
			t.Errorf("%v", err)
		}

		fmt.Printf("%v\n", string(output))
	}

	t.Fail()
}
