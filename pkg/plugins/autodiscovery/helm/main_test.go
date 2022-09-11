package helm

import (
	"fmt"
	"testing"

	discoveryConfig "github.com/updatecli/updatecli/pkg/core/pipeline/autodiscovery/config"
	goyaml "gopkg.in/yaml.v3"
)

var (
	expectedPipelines map[string]string = map[string]string{
		"Bump dependency \"minio\" for Helm Chart \"epinio\"": `name: Bump dependency "minio" for Helm Chart "epinio"
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
            files: []
            key: dependencies[0].name
            value: minio
        disablesourceinput: true
targets:
    minio:
        name: Bump Helm Chart dependency "minio" for Helm Chart "epinio"
        kind: helmchart
        spec:
            file: Chart.yaml
            key: dependencies[0].version
            name: epinio
            versionincrement: minor
        sourceid: minio
`,
		"Bump dependency \"kubed\" for Helm Chart \"epinio\"": `name: Bump dependency "kubed" for Helm Chart "epinio"
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
            files: []
            key: dependencies[1].name
            value: kubed
        disablesourceinput: true
targets:
    kubed:
        name: Bump Helm Chart dependency "kubed" for Helm Chart "epinio"
        kind: helmchart
        spec:
            file: Chart.yaml
            key: dependencies[1].version
            name: epinio
            versionincrement: minor
        sourceid: kubed
`,
		"Bump dependency \"epinio-ui\" for Helm Chart \"epinio\"": `name: Bump dependency "epinio-ui" for Helm Chart "epinio"
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
            files: []
            key: dependencies[2].name
            value: epinio-ui
        disablesourceinput: true
targets:
    epinio-ui:
        name: Bump Helm Chart dependency "epinio-ui" for Helm Chart "epinio"
        kind: helmchart
        spec:
            file: Chart.yaml
            key: dependencies[2].version
            name: epinio
            versionincrement: minor
        sourceid: epinio-ui
`,
		"Bump Docker Image \"epinioteam/epinio-ui-qa\" for Helm Chart \"epinio\"": `name: Bump Docker Image "epinioteam/epinio-ui-qa" for Helm Chart "epinio"
sources:
    epinioteam/epinio-ui-qa:
        name: Get latest "epinioteam/epinio-ui-qa" Container tag
        kind: dockerimage
        spec:
            image: epinioteam/epinio-ui-qa
            versionfilter:
                kind: semver
conditions:
    epinioteam/epinio-ui-qa:
        name: Ensure container repository "epinioteam/epinio-ui-qa" is specified
        kind: yaml
        spec:
            file: epinio/values.yaml
            files: []
            key: images.ui.repository
            value: epinioteam/epinio-ui-qa
        disablesourceinput: true
targets:
    epinioteam/epinio-ui-qa:
        name: Bump container image tag for image "epinioteam/epinio-ui-qa" in Chart "epinio"
        kind: helmchart
        spec:
            file: values.yaml
            key: images.ui.tag
            name: epinio
            versionincrement: minor
        sourceid: epinioteam/epinio-ui-qa
`,
		"Bump Docker Image \"splatform/epinio-server\" for Helm Chart \"epinio\"": `name: Bump Docker Image "splatform/epinio-server" for Helm Chart "epinio"
sources:
    splatform/epinio-server:
        name: Get latest "splatform/epinio-server" Container tag
        kind: dockerimage
        spec:
            image: splatform/epinio-server
            versionfilter:
                kind: semver
conditions:
    splatform/epinio-server:
        name: Ensure container repository "splatform/epinio-server" is specified
        kind: yaml
        spec:
            file: epinio/values.yaml
            files: []
            key: image.repository
            value: splatform/epinio-server
        disablesourceinput: true
targets:
    splatform/epinio-server:
        name: Bump container image tag for image "splatform/epinio-server" in Chart "epinio"
        kind: helmchart
        spec:
            file: values.yaml
            key: image.tag
            name: epinio
            versionincrement: minor
        sourceid: splatform/epinio-server
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

	pipelines, err := helm.DiscoverManifests(discoveryConfig.Input{})

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
