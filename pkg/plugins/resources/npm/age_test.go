package npm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
)

// publishDates mimics the `time` object returned by the npm registry, including the
// "created" and "modified" keys which are not versions.
// The dates are the ones used by the package fixtures, they are old enough for the
// assertions below to remain stable over time.
var publishDates = map[string]string{
	"modified": "2022-12-29T06:38:42.456Z",
	"created":  "2014-08-29T23:08:36.810Z",
	"0.1.0":    "2014-08-29T23:08:36.810Z",
	"0.2.0":    "2014-09-12T20:06:33.167Z",
}

func TestFilterVersionsByAge(t *testing.T) {
	tests := []struct {
		name           string
		versions       []string
		publishDates   map[string]string
		releaseAge     age.Spec
		expectedResult []string
	}{
		{
			name:           "No age filter defined returns every version",
			versions:       []string{"0.1.0", "0.2.0"},
			publishDates:   publishDates,
			releaseAge:     age.Spec{},
			expectedResult: []string{"0.1.0", "0.2.0"},
		},
		{
			name:           "Minimum age matching every version",
			versions:       []string{"0.1.0", "0.2.0"},
			publishDates:   publishDates,
			releaseAge:     age.Spec{Minimum: "1y"},
			expectedResult: []string{"0.1.0", "0.2.0"},
		},
		{
			name:           "Minimum age matching no version",
			versions:       []string{"0.1.0", "0.2.0"},
			publishDates:   publishDates,
			releaseAge:     age.Spec{Minimum: "100y"},
			expectedResult: []string{},
		},
		{
			name:           "Maximum age matching every version",
			versions:       []string{"0.1.0", "0.2.0"},
			publishDates:   publishDates,
			releaseAge:     age.Spec{Maximum: "100y"},
			expectedResult: []string{"0.1.0", "0.2.0"},
		},
		{
			name:           "Maximum age matching no version",
			versions:       []string{"0.1.0", "0.2.0"},
			publishDates:   publishDates,
			releaseAge:     age.Spec{Maximum: "1y"},
			expectedResult: []string{},
		},
		{
			name:           "Minimum and maximum age matching every version",
			versions:       []string{"0.1.0", "0.2.0"},
			publishDates:   publishDates,
			releaseAge:     age.Spec{Minimum: "1y", Maximum: "100y"},
			expectedResult: []string{"0.1.0", "0.2.0"},
		},
		{
			name:           "Version without any release date reported by the registry is ignored",
			versions:       []string{"0.1.0", "0.3.0"},
			publishDates:   publishDates,
			releaseAge:     age.Spec{Minimum: "1y"},
			expectedResult: []string{"0.1.0"},
		},
		{
			name:     "Version with an invalid release date is ignored",
			versions: []string{"0.1.0", "0.2.0"},
			publishDates: map[string]string{
				"0.1.0": "2014-08-29T23:08:36.810Z",
				"0.2.0": "not a date",
			},
			releaseAge:     age.Spec{Minimum: "1y"},
			expectedResult: []string{"0.1.0"},
		},
		{
			name:           "No release date at all reported by the registry",
			versions:       []string{"0.1.0", "0.2.0"},
			publishDates:   nil,
			releaseAge:     age.Spec{Minimum: "1y"},
			expectedResult: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := filterVersionsByAge(tt.versions, tt.publishDates, tt.releaseAge)
			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}

func TestLatestVersionMatchingAge(t *testing.T) {
	tests := []struct {
		name             string
		distTagLatest    string
		orderedVersions  []string
		matchingVersions []string
		expectedResult   string
	}{
		{
			name:             "Dist-tag latest matches the age filter",
			distTagLatest:    "0.2.0",
			orderedVersions:  []string{"0.1.0", "0.2.0"},
			matchingVersions: []string{"0.1.0", "0.2.0"},
			expectedResult:   "0.2.0",
		},
		{
			name:             "Dist-tag latest is too recent, fallback to the previously published version",
			distTagLatest:    "0.3.0",
			orderedVersions:  []string{"0.1.0", "0.2.0", "0.3.0"},
			matchingVersions: []string{"0.1.0", "0.2.0"},
			expectedResult:   "0.2.0",
		},
		{
			name:             "Fallback follows the publication order and not the version order",
			distTagLatest:    "2.0.0",
			orderedVersions:  []string{"1.0.0", "2.0.0", "1.0.1"},
			matchingVersions: []string{"1.0.0", "1.0.1"},
			expectedResult:   "1.0.1",
		},
		{
			name:             "Prereleases are ignored by the fallback",
			distTagLatest:    "2.0.0",
			orderedVersions:  []string{"1.0.0", "1.1.0", "2.0.0-rc.1", "2.0.0"},
			matchingVersions: []string{"1.0.0", "1.1.0", "2.0.0-rc.1"},
			expectedResult:   "1.1.0",
		},
		{
			name:             "Prereleases are used when nothing else matches",
			distTagLatest:    "2.0.0",
			orderedVersions:  []string{"2.0.0-rc.1", "2.0.0-rc.2", "2.0.0"},
			matchingVersions: []string{"2.0.0-rc.1", "2.0.0-rc.2"},
			expectedResult:   "2.0.0-rc.2",
		},
		{
			name:             "Non semantic versions are not treated as prereleases",
			distTagLatest:    "20240101",
			orderedVersions:  []string{"20230101", "20230601", "20240101"},
			matchingVersions: []string{"20230101", "20230601"},
			expectedResult:   "20230601",
		},
		{
			name:             "No version matches the age filter",
			distTagLatest:    "0.2.0",
			orderedVersions:  []string{"0.1.0", "0.2.0"},
			matchingVersions: []string{},
			expectedResult:   "",
		},
		{
			name:             "The registry didn't report any publication order",
			distTagLatest:    "0.3.0",
			orderedVersions:  nil,
			matchingVersions: []string{"0.1.0", "0.2.0"},
			expectedResult:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := latestVersionMatchingAge(tt.distTagLatest, tt.orderedVersions, tt.matchingVersions)
			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}
