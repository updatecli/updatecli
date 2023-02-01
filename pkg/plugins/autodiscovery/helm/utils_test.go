package helm

import (
	"testing"
)

func TestSearchFiles(t *testing.T) {

	gotFiles, err := searchChartFiles(
		"test/testdata/chart", ChartValidFiles[:])
	if err != nil {
		t.Errorf("%s\n", err)
	}
	expectedFile := "test/testdata/chart/epinio/Chart.yaml"

	if len(gotFiles) == 0 {
		t.Errorf("Expecting file %q but got none", expectedFile)
		return
	}

	if gotFiles[0] != expectedFile {
		t.Errorf("Expecting file %q but got %q", expectedFile, gotFiles[0])
	}
}

func TestListChartDependency(t *testing.T) {

	gotChartMetadata, err := getChartMetadata(
		"test/testdata/chart/epinio/Chart.yaml")
	if err != nil {
		t.Errorf("%s\n", err)
	}
	expectedChartName := "epinio"
	if gotChartMetadata.Name != expectedChartName {
		t.Errorf("Expecting Chart Name %q but got %q", expectedChartName, gotChartMetadata.Name)
	}
}
