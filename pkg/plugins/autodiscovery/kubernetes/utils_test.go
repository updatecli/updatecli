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
			path: "testdata/success",
			expectedResults: []string{
				"testdata/success/pod.yaml",
			},
		},
		{
			name: "case: all",
			path: "testdata",
			expectedResults: []string{
				"testdata/cronjob/cronjob.yaml",
				"testdata/initContainers/initContainers.yaml",
				"testdata/kustomize/deployment.yaml",
				"testdata/latest/pod.yaml",
				"testdata/prow/prow.yaml",
				"testdata/success/pod.yaml",
				"testdata/template/deployment.yaml",
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
			filepath:       "testdata/success/pod.yaml",
			expectedResult: []string{"ghcr.io/updatecli/updatecli:v0.67.0"},
		},
		{
			name:             "case: template",
			filepath:         "testdata/template/deployment.yaml",
			expectedResult:   []string{""},
			expectedError:    true,
			expectedErrorMsg: "yaml: line 19: could not find expected ':'",
		},
		{
			name:           "case: wrong flavor",
			filepath:       "testdata/prow/prow.yaml",
			expectedResult: []string{},
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

func TestGetProwManifestData(t *testing.T) {

	testdata := []struct {
		name             string
		filepath         string
		expectedResult   []string
		expectedError    bool
		expectedErrorMsg string
	}{
		{
			name:     "case: success",
			filepath: "testdata/prow/prow.yaml",
			expectedResult: []string{
				"ghcr.io/updatecli/updatecli:v0.82.2",
				"ghcr.io/updatecli/updatecli:v0.82.2",
				"ghcr.io/updatecli/updatecli:v0.82.2",
			},
		},
		{
			name:           "case: wrong flavor",
			filepath:       "testdata/success/pod.yaml",
			expectedResult: []string{},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {

			getProwManifestData, err := getProwManifestData(
				tt.filepath)

			if tt.expectedError {
				assert.EqualError(t, err, tt.expectedErrorMsg)
				return
			} else {
				assert.NoError(t, err)
			}

			gotContainers := []string{}
			for _, repo := range getProwManifestData.ProwPreSubmitJobs {
				for _, tests := range repo {
					for _, container := range tests.Spec.Containers {
						gotContainers = append(gotContainers, container.Image)
					}
				}
			}
			for _, repo := range getProwManifestData.ProwPostSubmitJobs {
				for _, tests := range repo {
					for _, container := range tests.Spec.Containers {
						gotContainers = append(gotContainers, container.Image)
					}
				}
			}
			for _, tests := range getProwManifestData.ProwPeriodicJobs {
				for _, container := range tests.Spec.Containers {
					gotContainers = append(gotContainers, container.Image)
				}
			}

			assert.Equal(t, tt.expectedResult, gotContainers)
		})
	}
}
