package fleet

import (
	"fmt"
	"testing"

	goyaml "gopkg.in/yaml.v3"
)

var (
	expectedPipelines map[string]string = map[string]string{
		"grafana-grafana": `name: grafana-grafana
sources:
    grafana:
        name: Get latest "grafana" Helm Chart Version
        kind: helmchart
        spec:
            name: grafana
            url: https://grafana.github.io/helm-charts
conditions:
    grafana-name:
        name: Ensure Helm chart name "grafana" is specified
        kind: yaml
        spec:
            file: grafana/fleet.yaml
            key: helm.chart
            value: grafana
        disablesourceinput: true
    grafana-repository:
        name: Ensure Helm chart repository "https://grafana.github.io/helm-charts" is specified
        kind: yaml
        spec:
            file: grafana/fleet.yaml
            key: helm.repo
            value: https://grafana.github.io/helm-charts
        disablesourceinput: true
targets:
    grafana:
        name: Bump chart "grafana" from Fleet bundle "grafana"
        kind: yaml
        spec:
            file: grafana/fleet.yaml
            key: helm.version
        sourceid: grafana
`}
)

func TestDiscoverManifests(t *testing.T) {

	spec := Spec{
		RootDir: "testdata/fleet.d",
	}

	helm, err := New(spec, "")

	if err != nil {
		t.Errorf("%v", err)
	}

	pipelines, err := helm.DiscoverManifests(nil, "", nil, "")

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
			fmt.Printf("Expected:\n>>>\n%q\n>>>\n\n", expectedPipelines[pipeline.Name])
			fmt.Printf("Got:\n<<<\n%q\n<<<\n", output)
			t.Fail()
		}

	}

}
