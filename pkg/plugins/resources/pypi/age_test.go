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
		name                 string
		spec                 Spec
		mockedBody           string
		expectedResult       string
		expectedStatus       string
		expectedError        bool
		expectedErrorMessage string
		expectedDescription  string
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
			// Like the npm resource, the age filter discarding every release skips the source.
			name:           "every release older than the maximum skips the source",
			spec:           Spec{Name: "requests", Age: age.Spec{Maximum: "1y"}},
			mockedBody:     recentPackageDataAt(730),
			expectedStatus: result.SKIPPED,
		},
		{
			name: "releases matching the versionfilter only older than the maximum skip the source",
			spec: Spec{
				Name:          "requests",
				Age:           age.Spec{Maximum: "1y"},
				VersionFilter: version.Filter{Kind: "pep440", Pattern: "==2.29.0"},
			},
			mockedBody:     agedPackageData(),
			expectedStatus: result.SKIPPED,
		},
		{
			name: "a development latest release ignored by the versionfilter kind is reported as such",
			spec: Spec{Name: "testpkg"},
			mockedBody: `{
  "info": {"name": "testpkg", "version": "2.0.dev3", "project_urls": {}},
  "releases": {
    "1.0.0": [{"yanked": false}],
    "2.0.dev3": [{"yanked": false}]
  }
}`,
			expectedError:        true,
			expectedErrorMessage: "is a development release",
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
			// Both normalize to "1.0.0" for semver, the base release is always the one returned.
			// The highest by PEP 440 order, so the newest build, on every run.
			name: "releases normalizing alike give the highest post release",
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
}`, daysAgo(60), daysAgo(30)),
			expectedResult: "1.0.0.post1",
			expectedStatus: result.SUCCESS,
		},
		{
			// Like npm, go/module and maven, so an age filter set by default doesn't break the pipeline.
			name:                "a registry reporting no upload time at all skips the source",
			spec:                Spec{Name: "requests", Age: age.Spec{Minimum: "7d"}},
			mockedBody:          existingPackageData,
			expectedStatus:      result.SKIPPED,
			expectedDescription: "reports no upload time",
		},
		{
			name: "a registry reporting no upload time at all skips a source with a version filter",
			spec: Spec{
				Name:          "requests",
				Age:           age.Spec{Minimum: "7d"},
				VersionFilter: version.Filter{Kind: "pep440", Pattern: ">=2.0"},
			},
			mockedBody:     existingPackageData,
			expectedStatus: result.SKIPPED,
		},
		{
			// The age filter isn't what prevents a match, so it fails like it would without age.
			name: "a version filter matching nothing still fails while every release is cooling down",
			spec: Spec{
				Name:          "requests",
				Age:           age.Spec{Minimum: "7d"},
				VersionFilter: version.Filter{Kind: "pep440", Pattern: ">=5.0"},
			},
			mockedBody:    recentPackageData(),
			expectedError: true,
		},
		{
			// A version filter bounded at the version in use, such as one generated by an
			// autodiscovery crawler, must not fall back below it: 2.30.0 matches age but not the filter.
			name: "a version filter bounded at a cooling down version skips rather than falling back below it",
			spec: Spec{
				Name:          "requests",
				Age:           age.Spec{Minimum: "7d"},
				VersionFilter: version.Filter{Kind: "semver", Pattern: ">=2.31.0"},
			},
			mockedBody:     agedPackageData(),
			expectedStatus: result.SKIPPED,
		},
		{
			// No release passes the age filter and the only one the filter matches has no upload time.
			name: "a version filter matching only undated releases skips while none passes the age filter",
			spec: Spec{
				Name:          "requests",
				Age:           age.Spec{Minimum: "7d"},
				VersionFilter: version.Filter{Kind: "pep440", Pattern: ">=2.31"},
			},
			mockedBody: fmt.Sprintf(`{
  "info": {"name": "requests", "version": "2.31.0", "project_urls": {}},
  "releases": {
    "2.30.0": [{"yanked": false, "upload_time_iso_8601": %q}],
    "2.31.0": [{"yanked": false}]
  }
}`, daysAgo(1)),
			expectedStatus: result.SKIPPED,
		},
		{
			// 2.30.0 passes the age filter, so an undated 2.31.0 is a registry gap, not a hold.
			name: "a version filter matching only undated releases fails while another passes the age filter",
			spec: Spec{
				Name:          "requests",
				Age:           age.Spec{Minimum: "7d"},
				VersionFilter: version.Filter{Kind: "pep440", Pattern: ">=2.31"},
			},
			mockedBody: fmt.Sprintf(`{
  "info": {"name": "requests", "version": "2.31.0", "project_urls": {}},
  "releases": {
    "2.30.0": [{"yanked": false, "upload_time_iso_8601": %q}],
    "2.31.0": [{"yanked": false}]
  }
}`, daysAgo(365)),
			expectedError: true,
		},
		{
			// Like npm with the kind "latest": nothing to fall back to is an age hold.
			name: "an undated latest release without any older release matching the age filter skips",
			spec: Spec{Name: "requests", Age: age.Spec{Minimum: "7d"}},
			mockedBody: fmt.Sprintf(`{
  "info": {"name": "requests", "version": "2.31.0", "project_urls": {}},
  "releases": {
    "2.30.0": [{"yanked": false, "upload_time_iso_8601": %q}],
    "2.31.0": [{"yanked": false}]
  }
}`, daysAgo(1)),
			expectedStatus: result.SKIPPED,
		},
		{
			// A registry only reporting the older "upload_time" field, in UTC without zone.
			name: "the legacy upload time is used when the ISO 8601 one is missing",
			spec: Spec{Name: "requests", Age: age.Spec{Minimum: "7d"}},
			mockedBody: fmt.Sprintf(`{
  "info": {"name": "requests", "version": "2.31.0", "project_urls": {}},
  "releases": {
    "2.30.0": [{"yanked": false, "upload_time": %q}],
    "2.31.0": [{"yanked": false, "upload_time": %q}]
  }
}`, time.Now().Add(-365*24*time.Hour).UTC().Format(legacyUploadTimeLayout), time.Now().Add(-24*time.Hour).UTC().Format(legacyUploadTimeLayout)),
			expectedResult: "2.30.0",
			expectedStatus: result.SUCCESS,
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
				if tt.expectedErrorMessage != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorMessage)
				}
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, gotResult.Result)
			assert.Equal(t, tt.expectedResult, gotResult.Information)
			if tt.expectedDescription != "" {
				assert.Contains(t, gotResult.Description, tt.expectedDescription)
			}
		})
	}
}

func TestConditionAge(t *testing.T) {
	tests := []struct {
		name            string
		version         string
		versionFilter   version.Filter
		age             age.Spec
		mockedBody      string
		expectedPass    bool
		expectedMessage string
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
			// The versionfilter only selects the source version, the condition checks existence.
			name:          "a versionfilter matching nothing doesn't make the condition fail",
			version:       "2.30.0",
			versionFilter: version.Filter{Kind: "pep440", Pattern: ">=99"},
			mockedBody:    agedPackageData(),
			expectedPass:  true,
		},
		{
			name:          "a semver versionfilter still finds a recent release as cooling down",
			version:       "2.31.0",
			versionFilter: version.Filter{Kind: "semver", Pattern: "*"},
			mockedBody:    agedPackageData(),
			expectedPass:  false,
		},
		{
			// 1.0.0 is cooling down, the old 1.0.0.post1 normalizes to the same semver version.
			name:          "a cooling down release doesn't pass through another one normalized alike",
			version:       "1.0.0",
			versionFilter: version.Filter{Kind: "semver", Pattern: "*"},
			mockedBody: fmt.Sprintf(`{
  "info": {"name": "requests", "version": "1.0.0.post1", "project_urls": {}},
  "releases": {
    "1.0.0": [{"yanked": false, "upload_time_iso_8601": %q}],
    "1.0.0.post1": [{"yanked": false, "upload_time_iso_8601": %q}]
  }
}`, daysAgo(1), daysAgo(30)),
			expectedPass:    false,
			expectedMessage: "exists but doesn't match the age filter yet",
		},
		{
			// Only 1.0.0.post1 exists, so the normalized "1.0.0" designates it, as without age filter.
			name:          "a normalized version only published as a cooling down post release is reported as cooling down",
			version:       "1.0.0",
			versionFilter: version.Filter{Kind: "semver", Pattern: "*"},
			mockedBody: fmt.Sprintf(`{
  "info": {"name": "requests", "version": "1.0.0.post1", "project_urls": {}},
  "releases": {
    "0.9.0": [{"yanked": false, "upload_time_iso_8601": %q}],
    "1.0.0.post1": [{"yanked": false, "upload_time_iso_8601": %q}]
  }
}`, daysAgo(30), daysAgo(1)),
			expectedPass:    false,
			expectedMessage: "doesn't match the age filter yet",
		},
		{
			name:          "a normalized version only published as a post release matching the age filter passes",
			version:       "1.0.0",
			versionFilter: version.Filter{Kind: "semver", Pattern: "*"},
			mockedBody: fmt.Sprintf(`{
  "info": {"name": "requests", "version": "1.0.0.post1", "project_urls": {}},
  "releases": {
    "0.9.0": [{"yanked": false, "upload_time_iso_8601": %q}],
    "1.0.0.post1": [{"yanked": false, "upload_time_iso_8601": %q}]
  }
}`, daysAgo(30), daysAgo(20)),
			expectedPass: true,
		},
		{
			// "1.0" normalizes to "1.0.0" like "1.0.post1", so the base release is the one checked.
			name:          "a two part base release is preferred over its cooling down post release",
			version:       "1.0.0",
			versionFilter: version.Filter{Kind: "semver", Pattern: "*"},
			mockedBody: fmt.Sprintf(`{
  "info": {"name": "requests", "version": "1.0.post1", "project_urls": {}},
  "releases": {
    "1.0": [{"yanked": false, "upload_time_iso_8601": %q}],
    "1.0.post1": [{"yanked": false, "upload_time_iso_8601": %q}]
  }
}`, daysAgo(30), daysAgo(1)),
			expectedPass: true,
		},
		{
			name:    "a release without upload time is reported as such, not as missing",
			version: "2.31.0",
			mockedBody: fmt.Sprintf(`{
  "info": {"name": "requests", "version": "2.31.0", "project_urls": {}},
  "releases": {
    "2.30.0": [{"yanked": false, "upload_time_iso_8601": %q}],
    "2.31.0": [{"yanked": false}]
  }
}`, daysAgo(365)),
			expectedPass:    false,
			expectedMessage: "reports no upload time",
		},
		{
			// "1.0.0-beta.2" is the normalized form of the "1.0b2" release published yesterday.
			name:          "a normalized version of a cooling down release is reported as cooling down",
			version:       "1.0.0-beta.2",
			versionFilter: version.Filter{Kind: "semver", Pattern: "*"},
			mockedBody: fmt.Sprintf(`{
  "info": {"name": "requests", "version": "0.9.0", "project_urls": {}},
  "releases": {
    "0.9.0": [{"yanked": false, "upload_time_iso_8601": %q}],
    "1.0b2": [{"yanked": false, "upload_time_iso_8601": %q}]
  }
}`, daysAgo(365), daysAgo(1)),
			expectedPass:    false,
			expectedMessage: "doesn't match the age filter yet",
		},
		{
			name:          "a normalized version of a release matching the age filter passes",
			version:       "1.0.0-beta.2",
			versionFilter: version.Filter{Kind: "semver", Pattern: "*"},
			mockedBody: fmt.Sprintf(`{
  "info": {"name": "requests", "version": "0.9.0", "project_urls": {}},
  "releases": {
    "0.9.0": [{"yanked": false, "upload_time_iso_8601": %q}],
    "1.0b2": [{"yanked": false, "upload_time_iso_8601": %q}]
  }
}`, daysAgo(365), daysAgo(30)),
			expectedPass: true,
		},
		{
			name:            "a release older than the maximum is reported as such",
			version:         "2.29.0",
			age:             age.Spec{Maximum: "30d"},
			mockedBody:      agedPackageData(),
			expectedPass:    false,
			expectedMessage: "older than the age filter allows",
		},
		{
			name:            "a registry reporting no upload time at all fails without error",
			version:         "2.31.0",
			mockedBody:      existingPackageData,
			expectedPass:    false,
			expectedMessage: "reports no upload time",
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
			releaseAge := tt.age
			if releaseAge.IsZero() {
				releaseAge = age.Spec{Minimum: "7d"}
			}
			p, err := New(Spec{Name: "requests", Version: tt.version, Age: releaseAge, VersionFilter: tt.versionFilter})
			require.NoError(t, err)

			p.webClient = GetMockClient(pypiDefaultURL, "", tt.mockedBody, 200)

			pass, message, err := p.Condition(context.Background(), "", nil, pathresolver.Resolver{})
			require.NoError(t, err)
			assert.Equal(t, tt.expectedPass, pass)
			if tt.expectedMessage != "" {
				assert.Contains(t, message, tt.expectedMessage)
			}
		})
	}
}

func TestReleasesOrder(t *testing.T) {
	rels := releases{
		"1.10.0":      newRelease("1.10.0", "1.10.0", ageMatching),
		"1.9.0":       newRelease("1.9.0", "1.9.0", ageMatching),
		"1.9.0.post1": newRelease("1.9.0.post1", "1.9.0", ageMatching),
		"1.9.0rc1":    newRelease("1.9.0rc1", "1.9.0-rc.1", ageMatching),
		"2.0.0":       newRelease("2.0.0", "2.0.0", ageTooRecent),
	}

	assert.Equal(t, []string{"1.9.0rc1", "1.9.0", "1.9.0.post1", "1.10.0"}, rels.versions(ageMatching))
	assert.Equal(t, []string{"1.9.0-rc.1", "1.9.0", "1.10.0", "2.0.0"}, rels.candidates(ageMatching, ageTooRecent))
	assert.Equal(t, "1.9.0.post1", rels.original("1.9.0"))
}

func TestSourceRegexPicksNewestRelease(t *testing.T) {
	// The regex kind takes the last match as the newest, so "1.10.0" must come after "1.9.0".
	p, err := New(Spec{Name: "testpkg", VersionFilter: version.Filter{Kind: "regex", Pattern: `^1\.`}})
	require.NoError(t, err)

	p.webClient = GetMockClient(pypiDefaultURL, "", `{
  "info": {"name": "testpkg", "version": "1.10.0", "project_urls": {}},
  "releases": {
    "1.9.0": [{"yanked": false}],
    "1.10.0": [{"yanked": false}],
    "1.2.0": [{"yanked": false}]
  }
}`, 200)

	gotResult := result.Source{}
	require.NoError(t, p.Source(context.Background(), pathresolver.Resolver{}, &gotResult))
	assert.Equal(t, "1.10.0", gotResult.Information)
}

func TestSourceRegexSeesShortReleasesAsPublished(t *testing.T) {
	p, err := New(Spec{Name: "testpkg", VersionFilter: version.Filter{Kind: "regex", Pattern: `^2\.0$`}})
	require.NoError(t, err)

	p.webClient = GetMockClient(pypiDefaultURL, "", `{
  "info": {"name": "testpkg", "version": "2.0", "project_urls": {}},
  "releases": {"1.0": [{"yanked": false}], "2.0": [{"yanked": false}]}
}`, 200)

	gotResult := result.Source{}
	require.NoError(t, p.Source(context.Background(), pathresolver.Resolver{}, &gotResult))
	assert.Equal(t, "2.0", gotResult.Information)
}

func TestLatestVersionMatchingAge(t *testing.T) {
	tests := []struct {
		name     string
		versions []string
		latest   string
		expected string
	}{
		{name: "highest stable below latest", versions: []string{"1.0", "1.10", "1.9", "2.0"}, latest: "2.0", expected: "1.10"},
		{name: "stable preferred over a higher pre release", versions: []string{"1.0", "1.1rc1"}, latest: "2.0", expected: "1.0"},
		{name: "pre release when no stable one is below latest", versions: []string{"1.0a1", "1.0b2"}, latest: "2.0", expected: "1.0b2"},
		{name: "pre release of latest itself", versions: []string{"1.0a1", "2.0rc1"}, latest: "2.0", expected: "2.0rc1"},
		{name: "stable preferred over a pre release of latest", versions: []string{"1.0", "2.0rc1"}, latest: "2.0", expected: "1.0"},
		{name: "no post release of latest", versions: []string{"2.0.post1"}, latest: "2.0", expected: ""},
		{name: "nothing above latest", versions: []string{"3.0"}, latest: "2.0", expected: ""},
		{name: "no version", latest: "2.0", expected: ""},
		{name: "public version below a local latest", versions: []string{"0.9", "1.0", "1.1"}, latest: "1.0+abc", expected: "1.0"},
		{name: "no local version above a local latest", versions: []string{"0.9", "1.0+zzz"}, latest: "1.0+abc", expected: "0.9"},
		{name: "latest isn't a PEP 440 version", versions: []string{"1.0"}, latest: "2004d1-final", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var matching []release
			for _, v := range tt.versions {
				matching = append(matching, newRelease(v, v, ageMatching))
			}

			got, err := latestVersionMatchingAge(matching, newRelease(tt.latest, tt.latest, ageTooRecent))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestSourceLatestNotPEP440Skips(t *testing.T) {
	p, err := New(Spec{Name: "testpkg", Age: age.Spec{Minimum: "7d"}})
	require.NoError(t, err)

	p.webClient = GetMockClient(pypiDefaultURL, "", fmt.Sprintf(`{
  "info": {"name": "testpkg", "version": "2004d1-final", "project_urls": {}},
  "releases": {
    "1.0": [{"yanked": false, "upload_time_iso_8601": %q}],
    "2004d1-final": [{"yanked": false, "upload_time_iso_8601": %q}]
  }
}`, daysAgo(365), daysAgo(1)), 200)

	gotResult := result.Source{}
	require.NoError(t, p.Source(context.Background(), pathresolver.Resolver{}, &gotResult))
	assert.Equal(t, result.SKIPPED, gotResult.Result)
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
