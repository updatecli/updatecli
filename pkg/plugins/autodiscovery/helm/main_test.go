package helm

import (
	"fmt"
	"testing"

	goyaml "gopkg.in/yaml.v3"
)

var (
	expectedPipelines map[string]string = map[string]string{
		"epinio-minio": `name: epinio-minio
sources:
    minio:
        name: Get latest "minio" Helm Chart Version
        kind: helmchart
        spec:
            name: minio
            url: https://charts.min.io/
conditions:
    minio:
        name: Ensure dependency "minio" is specified
        kind: yaml
        spec:
            file: epinio/Chart.yaml
            key: dependencies[0].name
            value: minio
        disablesourceinput: true
targets:
    minio:
        name: Bump chart dependency "minio" in Chart "epinio"
        kind: helmchart
        spec:
            file: Chart.yaml
            key: dependencies[0].version
            name: epinio
            versionincrement: minor
        sourceid: minio
`,
		"epinio-kubed": `name: epinio-kubed
sources:
    kubed:
        name: Get latest "kubed" Helm Chart Version
        kind: helmchart
        spec:
            name: kubed
            url: https://charts.appscode.com/stable
conditions:
    kubed:
        name: Ensure dependency "kubed" is specified
        kind: yaml
        spec:
            file: epinio/Chart.yaml
            key: dependencies[1].name
            value: kubed
        disablesourceinput: true
targets:
    kubed:
        name: Bump chart dependency "kubed" in Chart "epinio"
        kind: helmchart
        spec:
            file: Chart.yaml
            key: dependencies[1].version
            name: epinio
            versionincrement: minor
        sourceid: kubed
`,
		"epinio-epinio-ui": `name: epinio-epinio-ui
sources:
    epinio-ui:
        name: Get latest "epinio-ui" Helm Chart Version
        kind: helmchart
        spec:
            name: epinio-ui
            url: https://epinio.github.io/helm-charts
conditions:
    epinio-ui:
        name: Ensure dependency "epinio-ui" is specified
        kind: yaml
        spec:
            file: epinio/Chart.yaml
            key: dependencies[2].name
            value: epinio-ui
        disablesourceinput: true
targets:
    epinio-ui:
        name: Bump chart dependency "epinio-ui" in Chart "epinio"
        kind: helmchart
        spec:
            file: Chart.yaml
            key: dependencies[2].version
            name: epinio
            versionincrement: minor
        sourceid: epinio-ui
`}
)

func TestDiscoverManifests(t *testing.T) {

	spec := Spec{
		RootDir: "testdata/chart",
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

		if string(output) != expectedPipelines[pipeline.Name] {
			fmt.Println("Wrong result")
			fmt.Printf("Expected:\n>>>\n%q\n>>>\n\n", expectedPipelines[pipeline.Name])
			fmt.Printf("Got:\n<<<\n%q\n<<<\n", output)
			t.Fail()
		}

	}

}
