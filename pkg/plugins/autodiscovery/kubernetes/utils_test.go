package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchFiles(t *testing.T) {

	testdata := []struct {
		name            string
		path            string
		expectedResults []string
	}{
		{
			name: "case: success",
			path: "test/testdata/success",
			expectedResults: []string{
				"test/testdata/success/pod.yaml",
			},
		},
		{
			name: "case: all",
			path: "test/testdata",
			expectedResults: []string{
				"test/testdata/kustomize/deployment.yaml",
				"test/testdata/latest/pod.yaml",
				"test/testdata/success/pod.yaml",
				"test/testdata/template/deployment.yaml",
			},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			gotFiles, err := searchKubernetesFiles(
				tt.path, DefaultKubernetesFiles[:])
			if err != nil {
				t.Errorf("%s\n", err)
			}

			assert.Equal(t, tt.expectedResults, gotFiles)
		})
	}
}

func TestGetKubernetesManifestData(t *testing.T) {

	testdata := []struct {
		name             string
		filepath         string
		expectedResult   []string
		expectedError    bool
		expectedErrorMsg string
	}{
		{
			name:           "case: success",
			filepath:       "test/testdata/success/pod.yaml",
			expectedResult: []string{"ghcr.io/updatecli/updatecli:v0.67.0"},
		},
		{
			name:             "case: template",
			filepath:         "test/testdata/template/deployment.yaml",
			expectedResult:   []string{""},
			expectedError:    true,
			expectedErrorMsg: "yaml: line 19: could not find expected ':'",
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {

			gotKubernetesManifestData, err := getKubernetesManifestData(
				tt.filepath)

			if tt.expectedError {
				assert.EqualError(t, err, tt.expectedErrorMsg)
				return
			} else {
				assert.NoError(t, err)
			}

			gotContainers := []string{}
			for _, container := range gotKubernetesManifestData.Spec.Containers {
				gotContainers = append(gotContainers, container.Image)
			}

			assert.Equal(t, tt.expectedResult, gotContainers)
		})
	}
}
