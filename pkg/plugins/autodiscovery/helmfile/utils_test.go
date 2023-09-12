package helmfile

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchFiles(t *testing.T) {

	gotFiles, err := searchHelmfileFiles(
		"test/testdata/helmfile.d", DefaultFilePattern[:])
	if err != nil {
		t.Errorf("%s\n", err)
	}

	pwd, err := os.Getwd()
	require.NoError(t, err)

	expectedFiles := []string{
		path.Join(pwd, "test/testdata/helmfile.d/cik8s.yaml"),
	}

	assert.Equal(t, expectedFiles, gotFiles)

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
