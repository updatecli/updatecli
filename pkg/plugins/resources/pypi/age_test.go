package pypi

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// daysAgo returns an ISO 8601 upload time the given number of days in the past.
func daysAgo(days int) string {
	return time.Now().Add(-time.Duration(days) * 24 * time.Hour).UTC().Format(time.RFC3339)
}

// agedPackageData returns a payload where 2.31.0 was released two days ago,
// 2.30.0 twenty days ago, and 2.29.0 a year ago.
// 2.30.0 has a second file uploaded yesterday, which must not restart its cooldown.
func agedPackageData() string {
	return fmt.Sprintf(`{
  "info": {"name": "requests", "version": "2.31.0", "project_urls": {}},
  "releases": {
    "2.29.0": [{"yanked": false, "upload_time_iso_8601": %q}],
    "2.30.0": [
      {"yanked": false, "upload_time_iso_8601": %q},
      {"yanked": false, "upload_time_iso_8601": %q}
    ],
    "2.31.0": [{"yanked": false, "upload_time_iso_8601": %q}],
    "2.32.0rc1": [{"yanked": false, "upload_time_iso_8601": %q}]
  }
}`, daysAgo(365), daysAgo(20), daysAgo(1), daysAgo(2), daysAgo(30))
}

// recentPackageDataAt returns a payload where every release is the given number of days old.
func recentPackageDataAt(days int) string {
	return fmt.Sprintf(`{
  "info": {"name": "requests", "version": "2.31.0", "project_urls": {}},
  "releases": {
    "2.30.0": [{"yanked": false, "upload_time_iso_8601": %q}],
    "2.31.0": [{"yanked": false, "upload_time_iso_8601": %q}]
  }
}`, daysAgo(days), daysAgo(days))
}

// recentPackageData returns a payload where every release is a day old.
func recentPackageData() string {
	return fmt.Sprintf(`{
  "info": {"name": "requests", "version": "2.31.0", "project_urls": {}},
  "releases": {
    "2.30.0": [{"yanked": false, "upload_time_iso_8601": %q}],
    "2.31.0": [{"yanked": false, "upload_time_iso_8601": %q}]
  }
}`, daysAgo(1), daysAgo(1))
}

func TestSourceAge(t *testing.T) {
	tests := []struct {
		name           string
		spec           Spec
		mockedBody     string
		expectedResult string
		expectedStatus string
		expectedError  bool
	}{
		{
			name:           "latest falls back to the highest stable release older than the minimum",
			spec:           Spec{Name: "requests", Age: age.Spec{Minimum: "7d"}},
			mockedBody:     agedPackageData(),
			expectedResult: "2.30.0",
			expectedStatus: result.SUCCESS,
		},
		{
			name:           "latest is used when it matches the age filter",
			spec:           Spec{Name: "requests", Age: age.Spec{Minimum: "1d"}},
			mockedBody:     agedPackageData(),
			expectedResult: "2.31.0",
			expectedStatus: result.SUCCESS,
		},
		{
			name: "pep440 filter only sees releases older than the minimum",
			spec: Spec{
				Name:          "requests",
				Age:           age.Spec{Minimum: "7d"},
				VersionFilter: version.Filter{Kind: "pep440", Pattern: ">=2.0"},
			},
			mockedBody:     agedPackageData(),
			expectedResult: "2.30.0",
			expectedStatus: result.SUCCESS,
		},
		{
			name: "maximum discards releases that are too old",
			spec: Spec{
				Name:          "requests",
				Age:           age.Spec{Maximum: "30d"},
				VersionFilter: version.Filter{Kind: "pep440", Pattern: "<2.31"},
			},
			mockedBody:     agedPackageData(),
			expectedResult: "2.30.0",
			expectedStatus: result.SUCCESS,
		},
		{
			name:           "source is skipped while every release is cooling down",
			spec:           Spec{Name: "requests", Age: age.Spec{Minimum: "7d"}},
			mockedBody:     recentPackageData(),
			expectedStatus: result.SKIPPED,
		},
		{
			// 2.31.0 is the only release matching the filter and it is two days old.
			name: "source is skipped while only the releases matching the version filter are cooling down",
			spec: Spec{
				Name:          "requests",
				Age:           age.Spec{Minimum: "7d"},
				VersionFilter: version.Filter{Kind: "pep440", Pattern: "==2.31.0"},
			},
			mockedBody:     agedPackageData(),
			expectedStatus: result.SKIPPED,
		},
		{
			name: "dev releases don't count as cooling down",
			spec: Spec{
				Name:          "testpkg",
				Age:           age.Spec{Minimum: "7d"},
				VersionFilter: version.Filter{Kind: "semver", Pattern: "*"},
			},
			mockedBody: fmt.Sprintf(`{
  "info": {"name": "testpkg", "version": "1.0.dev3", "project_urls": {}},
  "releases": {"1.0.dev3": [{"yanked": false, "upload_time_iso_8601": %q}]}
}`, daysAgo(1)),
			expectedError: true,
		},
		{
			name: "a version filter matching nothing at all still fails",
			spec: Spec{
				Name:          "requests",
				Age:           age.Spec{Minimum: "7d"},
				VersionFilter: version.Filter{Kind: "pep440", Pattern: ">=5.0"},
			},
			mockedBody:    agedPackageData(),
			expectedError: true,
		},
		{
			// 1.0.0 and 1.0.0.post1 both normalize to "1.0.0", only the old one may be returned.
			name: "a recent post release normalized like an old release isn't returned",
			spec: Spec{
				Name:          "testpkg",
				Age:           age.Spec{Minimum: "7d"},
				VersionFilter: version.Filter{Kind: "semver", Pattern: "*"},
			},
			mockedBody: fmt.Sprintf(`{
  "info": {"name": "testpkg", "version": "1.0.0.post1", "project_urls": {}},
  "releases": {
    "1.0.0": [{"yanked": false, "upload_time_iso_8601": %q}],
    "1.0.0.post1": [{"yanked": false, "upload_time_iso_8601": %q}]
  }
}`, daysAgo(365), daysAgo(1)),
			expectedResult: "1.0.0",
			expectedStatus: result.SUCCESS,
		},
		{
			name: "a yanked latest release is still reported as yanked with an age filter",
			spec: Spec{Name: "requests", Age: age.Spec{Minimum: "7d"}},
			mockedBody: fmt.Sprintf(`{
  "info": {"name": "requests", "version": "2.31.0", "project_urls": {}},
  "releases": {
    "2.30.0": [{"yanked": false, "upload_time_iso_8601": %q}],
    "2.31.0": [{"yanked": true, "upload_time_iso_8601": %q}]
  }
}`, daysAgo(365), daysAgo(300)),
			expectedError: true,
		},
		{
			// Releases older than the maximum never match again, so it isn't a cooldown either.
			name:          "every release older than the maximum fails instead of skipping",
			spec:          Spec{Name: "requests", Age: age.Spec{Maximum: "1y"}},
			mockedBody:    recentPackageDataAt(730),
			expectedError: true,
		},
		{
			name: "releases matching the versionfilter only older than the maximum fail",
			spec: Spec{
				Name:          "requests",
				Age:           age.Spec{Maximum: "1y"},
				VersionFilter: version.Filter{Kind: "pep440", Pattern: "==2.29.0"},
			},
			mockedBody:    agedPackageData(),
			expectedError: true,
		},
		{
			// 3.0.0 was uploaded out of order, the fallback stays below the registry latest.
			name: "latest fallback never goes above the latest version reported by the registry",
			spec: Spec{Name: "testpkg", Age: age.Spec{Minimum: "7d"}},
			mockedBody: fmt.Sprintf(`{
  "info": {"name": "testpkg", "version": "2.0.0", "project_urls": {}},
  "releases": {
    "1.9.0": [{"yanked": false, "upload_time_iso_8601": %q}],
    "2.0.0": [{"yanked": false, "upload_time_iso_8601": %q}],
    "3.0.0": [{"yanked": false, "upload_time_iso_8601": %q}]
  }
}`, daysAgo(60), daysAgo(1), daysAgo(30)),
			expectedResult: "1.9.0",
			expectedStatus: result.SUCCESS,
		},
		{
			// Never ending, so it isn't a cooldown to skip.
			name:          "a registry reporting no upload time at all fails",
			spec:          Spec{Name: "requests", Age: age.Spec{Minimum: "7d"}},
			mockedBody:    existingPackageData,
			expectedError: true,
		},
		{
			name: "releases without upload time are ignored when others have one",
			spec: Spec{Name: "requests", Age: age.Spec{Minimum: "7d"}},
			mockedBody: fmt.Sprintf(`{
  "info": {"name": "requests", "version": "2.31.0", "project_urls": {}},
  "releases": {
    "2.30.0": [{"yanked": false, "upload_time_iso_8601": %q}],
    "2.31.0": [{"yanked": false}]
  }
}`, daysAgo(365)),
			expectedResult: "2.30.0",
			expectedStatus: result.SUCCESS,
		},
		{
			// 2.0.0 and 2.0.0.post1 normalize alike, the latest one is compared as published.
			name: "latest cooling down falls back even when an older release normalizes alike",
			spec: Spec{Name: "testpkg", Age: age.Spec{Minimum: "7d"}},
			mockedBody: fmt.Sprintf(`{
  "info": {"name": "testpkg", "version": "2.0.0.post1", "project_urls": {}},
  "releases": {
    "2.0.0": [{"yanked": false, "upload_time_iso_8601": %q}],
    "2.0.0.post1": [{"yanked": false, "upload_time_iso_8601": %q}]
  }
}`, daysAgo(30), daysAgo(1)),
			expectedResult: "2.0.0",
			expectedStatus: result.SUCCESS,
		},
		{
			name: "latest matching is returned even when an older release normalizes alike",
			spec: Spec{Name: "testpkg", Age: age.Spec{Minimum: "7d"}},
			mockedBody: fmt.Sprintf(`{
  "info": {"name": "testpkg", "version": "2.0.0.post1", "project_urls": {}},
  "releases": {
    "2.0.0": [{"yanked": false, "upload_time_iso_8601": %q}],
    "2.0.0.post1": [{"yanked": false, "upload_time_iso_8601": %q}]
  }
}`, daysAgo(60), daysAgo(30)),
			expectedResult: "2.0.0.post1",
			expectedStatus: result.SUCCESS,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New(tt.spec)
			require.NoError(t, err)

			p.webClient = GetMockClient(pypiDefaultURL, "", tt.mockedBody, 200)

			gotResult := result.Source{}
			err = p.Source(context.Background(), pathresolver.Resolver{}, &gotResult)
			if tt.expectedError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, gotResult.Result)
			assert.Equal(t, tt.expectedResult, gotResult.Information)
		})
	}
}

func TestConditionAge(t *testing.T) {
	tests := []struct {
		name         string
		version      string
		mockedBody   string
		expectedPass bool
	}{
		{
			name:         "release older than the minimum passes",
			version:      "2.30.0",
			mockedBody:   agedPackageData(),
			expectedPass: true,
		},
		{
			name:         "release newer than the minimum fails",
			version:      "2.31.0",
			mockedBody:   agedPackageData(),
			expectedPass: false,
		},
		{
			name:         "cooldown running on every release fails without error",
			version:      "2.31.0",
			mockedBody:   recentPackageData(),
			expectedPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New(Spec{Name: "requests", Version: tt.version, Age: age.Spec{Minimum: "7d"}})
			require.NoError(t, err)

			p.webClient = GetMockClient(pypiDefaultURL, "", tt.mockedBody, 200)

			pass, _, err := p.Condition(context.Background(), "", nil, pathresolver.Resolver{})
			require.NoError(t, err)
			assert.Equal(t, tt.expectedPass, pass)
		})
	}
}

func TestReleaseDate(t *testing.T) {
	files := []releaseFile{
		{UploadTime: "2024-03-10T00:00:00Z"},
		{UploadTime: "2024-03-01T00:00:00Z", Yanked: true},
		{UploadTime: "2024-03-05T00:00:00Z"},
		{UploadTime: "not a date"},
	}

	got, found := releaseDate(files)
	require.True(t, found)
	assert.Equal(t, "2024-03-05T00:00:00Z", got.Format(time.RFC3339))

	_, found = releaseDate([]releaseFile{{}})
	assert.False(t, found)
}

func TestNewInvalidAge(t *testing.T) {
	_, err := New(Spec{Name: "requests", Age: age.Spec{Minimum: "seven days"}})
	assert.Error(t, err)
}
