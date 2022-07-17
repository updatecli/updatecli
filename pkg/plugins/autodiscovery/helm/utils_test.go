package helm

import (
	"testing"
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
