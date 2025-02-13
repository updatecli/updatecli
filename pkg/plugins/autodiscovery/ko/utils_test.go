package ko

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
				"testdata/success/.ko.yaml",
			},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			gotFiles, err := searchKosFiles(
				tt.path, DefaultKoFiles[:])
			if err != nil {
				t.Errorf("%s\n", err)
			}

			assert.Equal(t, tt.expectedResults, gotFiles)
		})
	}
}

func TestGetContainerManifestData(t *testing.T) {

	testdata := []struct {
		name             string
		filepath         string
		expectedResult   []string
		expectedError    bool
		expectedErrorMsg string
	}{
		{
			name:           "case: success",
			filepath:       "testdata/success/.ko.yaml",
			expectedResult: []string{"golang:1.19.0"},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {

			gotKubernetesManifestData, err := getKoManifestData(
				tt.filepath)

			if tt.expectedError {
				assert.EqualError(t, err, tt.expectedErrorMsg)
				return
			} else {
				assert.NoError(t, err)
			}

			gotContainers := []string{}
			for _, container := range gotKubernetesManifestData.BaseImageOverrides {
				gotContainers = append(gotContainers, container)
			}

			assert.Equal(t, tt.expectedResult, gotContainers)
		})
	}
}
