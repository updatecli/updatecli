package argocd

import (
	"testing"
)

func TestSearchFiles(t *testing.T) {

	gotFiles, err := searchArgoCDFiles(
		"testdata", ArgoCDFilePatterns[:])
	if err != nil {
		t.Errorf("%s\n", err)
	}

	expectedFile := "testdata/sealed-secrets/manifest.yaml"
	if gotFiles[0] != expectedFile {
		t.Errorf("Expecting file %q but got %q", expectedFile, gotFiles[0])
	}
}

func TestListChartDependency(t *testing.T) {

	gotChartData, err := loadApplicationData(
		"testdata/sealed-secrets/manifest.yaml")
	if err != nil {
		t.Errorf("%s\n", err)
	}
	expectedChartName := "sealed-secrets"
	if gotChartData.Spec.Source.Chart != expectedChartName {
		t.Errorf("Expecting Chart Name %q but got %q", expectedChartName, gotChartData.Spec.Source.Chart)
	}
}
