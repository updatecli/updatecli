package helm

import (
	"testing"
)

func TestSearchFiles(t *testing.T) {

	gotFiles, err := searchChartFiles(
		"testdata/chart", ChartValidFiles[:])
	if err != nil {
		t.Errorf("%s\n", err)
	}

	expectedFile := "testdata/chart/epinio/Chart.yaml"
	if gotFiles[0] != expectedFile {
		t.Errorf("Expecting file %q but got %q", expectedFile, gotFiles[0])
	}
}

func TestListChartDependency(t *testing.T) {

	gotChartMetadata, err := getChartMetadata(
		"testdata/chart/epinio/Chart.yaml")
	if err != nil {
		t.Errorf("%s\n", err)
	}
	expectedChartName := "epinio"
	if gotChartMetadata.Name != expectedChartName {
		t.Errorf("Expecting Chart Name %q but got %q", expectedChartName, gotChartMetadata.Name)
	}
}
