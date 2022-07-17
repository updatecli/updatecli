package helm

import (
	"fmt"
	"testing"

	goyaml "gopkg.in/yaml.v3"
)

func TestSearchFiles(t *testing.T) {

	_, err := searchChartMetadataFiles(
		"/home/olblak/Projects/Epinio/helm-charts",
		[]string{"Chart.yaml", "Chart.yml"})
	if err != nil {
		t.Errorf("%s\n", err)
	}
	t.Fail()
}

func TestListChartDependency(t *testing.T) {

	_, err := getChartMetadata(
		"/home/olblak/Projects/Epinio/helm-charts/chart/epinio/Chart.yaml")
	if err != nil {
		t.Errorf("%s\n", err)
	}
	t.Fail()
}

func TestPipelines(t *testing.T) {

	spec := Spec{
		// Shoud work but doesn't due to final /
		// Must investigate for Windows as well
		//RootDir: "/home/olblak/Projects/Epinio/helm-charts/",
		RootDir: "/home/olblak/Projects/Epinio/helm-charts/",
	}

	helm, err := New(spec, "")

	if err != nil {
		t.Errorf("%v", err)
	}

	pipelines, err := helm.DiscoverManifests(nil)
	//pipelines, err := helm.Manifests(&scm.Config{
	//	Kind: "git",
	//})

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
