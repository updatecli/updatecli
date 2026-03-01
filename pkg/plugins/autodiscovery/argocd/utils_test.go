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
		"testdata/empty/manifest.yaml",
		"testdata/multi-release/manifest.yaml",
		"testdata/oci-helm-source/manifest.yaml",
		"testdata/sealed-secrets/manifest.yaml",
		"testdata/sealed-secrets_sources/manifest.yaml",
	}

	assert.Equal(t, expectedFiles, gotFiles)
}

func TestListChartDependency(t *testing.T) {

	gotChartData, err := readManifest(
		"testdata/sealed-secrets/manifest.yaml")
	if err != nil {
		t.Errorf("%s\n", err)
	}

	for _, data := range gotChartData {
		expectedChartName := "sealed-secrets"
		if data.Spec.Source.Chart != expectedChartName {
			t.Errorf("Expecting Chart Name %q but got %q", expectedChartName, data.Spec.Source.Chart)
		}
	}

}
