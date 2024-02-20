package fleet

import (
	"testing"
)

func TestSearchFiles(t *testing.T) {

	gotFiles, err := searchFleetBundleFiles(
		"testdata/fleet.d", FleetBundleFiles[:])
	if err != nil {
		t.Errorf("%s\n", err)
	}

	expectedFile := "testdata/fleet.d/grafana/fleet.yaml"
	if gotFiles[0] != expectedFile {
		t.Errorf("Expecting file %q but got %q", expectedFile, gotFiles[0])
	}
}

func TestListChartDependency(t *testing.T) {

	gotFleetBundleData, err := getFleetBundleData(
		"testdata/fleet.d/grafana/fleet.yaml")
	if err != nil {
		t.Errorf("%s\n", err)
	}
	expectedBundleName := "grafana"
	if gotFleetBundleData.Helm.Chart != expectedBundleName {
		t.Errorf("Expecting Chart Name %q but got %q", expectedBundleName, gotFleetBundleData.Helm.Chart)
	}
}
