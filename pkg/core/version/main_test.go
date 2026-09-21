package version

import (
	"testing"
)

func TestIsGreaterThan(t *testing.T) {

	type data struct {
		name                             string
		updatecliBinaryVersion           string
		updatecliManifestRequiredVersion string
		expectedResult                   bool
	}

	dataset := []data{
		{
			name:                             "binary older than the manifest requirement",
			updatecliBinaryVersion:           "1.0.0",
			updatecliManifestRequiredVersion: "1.2.0",
			expectedResult:                   false,
		},
		{
			name:                             "development binary ignores the manifest requirement",
			updatecliBinaryVersion:           "",
			updatecliManifestRequiredVersion: "1.2.0",
			expectedResult:                   true,
		},
		{
			name:                             "manifest without requirement",
			updatecliBinaryVersion:           "1.0.0",
			updatecliManifestRequiredVersion: "",
			expectedResult:                   true,
		},
		{
			name:                             "binary newer than the manifest requirement",
			updatecliBinaryVersion:           "1.2.0",
			updatecliManifestRequiredVersion: "1.0.0",
			expectedResult:                   true,
		},
		{
			name:                             "no version on either side",
			updatecliBinaryVersion:           "",
			updatecliManifestRequiredVersion: "",
			expectedResult:                   true,
		},
		{
			name:                             "release candidate satisfies the version it is a candidate for",
			updatecliBinaryVersion:           "1.0.0-rc.1",
			updatecliManifestRequiredVersion: "1.0.0",
			expectedResult:                   true,
		},
		{
			name:                             "release candidate does not satisfy a later minor",
			updatecliBinaryVersion:           "1.0.0-rc.1",
			updatecliManifestRequiredVersion: "1.1.0",
			expectedResult:                   false,
		},
		{
			name:                             "release candidate satisfies an older requirement",
			updatecliBinaryVersion:           "1.0.0-rc.1",
			updatecliManifestRequiredVersion: "0.9.0",
			expectedResult:                   true,
		},
		{
			name:                             "release candidate with a manifest without requirement",
			updatecliBinaryVersion:           "1.0.0-rc.1",
			updatecliManifestRequiredVersion: "",
			expectedResult:                   true,
		},
		{
			name:                             "build metadata is ignored",
			updatecliBinaryVersion:           "1.0.0+20260918",
			updatecliManifestRequiredVersion: "1.0.0",
			expectedResult:                   true,
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {

			result, err := IsGreaterThan(d.updatecliBinaryVersion, d.updatecliManifestRequiredVersion)

			if err != nil {
				t.Errorf("Unexpected error for %v: %s", d, err)
			}

			if result != d.expectedResult {
				t.Errorf("Unexpected result for %v, got %v", d, result)
			}

		})
	}
}
