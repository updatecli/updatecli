package argocd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchFiles(t *testing.T) {

	gotFiles, err := searchArgoCDFiles(
		"testdata", ArgoCDFilePatterns[:])
	if err != nil {
		t.Errorf("%s\n", err)
	}

	expectedFiles := []string{
		"testdata/oci-helm-source/manifest.yaml",
		"testdata/sealed-secrets/manifest.yaml",
		"testdata/sealed-secrets_sources/manifest.yaml",
	}

	assert.Equal(t, gotFiles, expectedFiles)
}

func TestListChartDependency(t *testing.T) {

	gotChartData, err := readManifest(
		"testdata/sealed-secrets/manifest.yaml")
	if err != nil {
		t.Errorf("%s\n", err)
	}
	expectedChartName := "sealed-secrets"
	if gotChartData.Spec.Source.Chart != expectedChartName {
		t.Errorf("Expecting Chart Name %q but got %q", expectedChartName, gotChartData.Spec.Source.Chart)
	}
}
