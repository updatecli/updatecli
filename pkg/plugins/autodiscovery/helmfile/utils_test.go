package helmfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchFiles(t *testing.T) {

	gotFiles, err := searchHelmfileFiles(
		"test/testdata/helmfile.d", DefaultFilePattern[:])
	if err != nil {
		t.Errorf("%s\n", err)
	}
	expectedFile := "test/testdata/helmfile.d/cik8s.yaml"

	if len(gotFiles) == 0 {
		t.Errorf("Expecting file %q but got none", expectedFile)
		return
	}

	if gotFiles[0] != expectedFile {
		t.Errorf("Expecting file %q but got %q", expectedFile, gotFiles[0])
	}
}

func TestListChartDependency(t *testing.T) {

	gotMetadata, err := getHelmfileMetadata(
		"test/testdata/helmfile.d/cik8s.yaml")
	if err != nil {
		t.Errorf("%s\n", err)
	}
	expectedReleases := []release{
		{
			Name:    "datadog",
			Chart:   "datadog/datadog",
			Version: "3.1.3",
		},
		{
			Name:    "docker-registry-secrets",
			Chart:   "jenkins-infra/docker-registry-secrets",
			Version: "0.1.0",
		},
		{
			Name:    "jenkins-agents",
			Chart:   "jenkins-infra/jenkins-kubernetes-agents",
			Version: "",
		},
		{
			Name:    "myOCIChart",
			Chart:   "myOCIRegistry/myOCIChart",
			Version: "0.1.0",
		},
	}

	expectedRepositories := []repository{
		{
			Name: "autoscaler",
			URL:  "https://kubernetes.github.io/autoscaler",
		},
		{
			Name: "datadog",
			URL:  "https://helm.datadoghq.com",
		},
		{
			Name: "eks",
			URL:  "https://aws.github.io/eks-charts",
		},
		{
			Name: "jenkins-infra",
			URL:  "https://jenkins-infra.github.io/helm-charts",
		},
		{
			Name: "myOCIRegistry",
			URL:  "myregistry.azurecr.io",
			OCI:  true,
		},
	}

	assert.Equal(t, expectedReleases, gotMetadata.Releases)
	assert.Equal(t, expectedRepositories, gotMetadata.Repositories)
}
