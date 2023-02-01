package version

import (
	"fmt"
	"testing"
)

func TestIsGreaterThan(t *testing.T) {

	type data struct {
		updatecliBinaryVersion           string
		updatecliManifestRequiredVersion string
		expectedResult                   bool
	}

	dataset := []data{
		{
			updatecliBinaryVersion:           "1.0.0",
			updatecliManifestRequiredVersion: "1.2.0",
			expectedResult:                   false,
		},
		{
			updatecliBinaryVersion:           "",
			updatecliManifestRequiredVersion: "1.2.0",
			expectedResult:                   true,
		},
		{
			updatecliBinaryVersion:           "1.0.0",
			updatecliManifestRequiredVersion: "",
			expectedResult:                   true,
		},
		{
			updatecliBinaryVersion:           "1.2.0",
			updatecliManifestRequiredVersion: "1.0.0",
			expectedResult:                   true,
		},
		{
			updatecliBinaryVersion:           "",
			updatecliManifestRequiredVersion: "",
			expectedResult:                   true,
		},
	}

	for _, d := range dataset {

		result, err := IsGreaterThan(d.updatecliBinaryVersion, d.updatecliManifestRequiredVersion)

		if err != nil {
			fmt.Println(err)
			t.Error()
		}

		if result != d.expectedResult {
			t.Errorf("Unexpected result for %v, got %v", d, result)
		}

	}
}
