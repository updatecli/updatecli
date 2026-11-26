package pyproject

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestNewAgeNeverDowngrades(t *testing.T) {
	testdata := []struct {
		name string
		spec Spec
		// expectedPatterns maps each dependency getting a manifest to the pattern of its source
		// versionfilter. numpy, which uv.lock resolves at several versions, never gets one, as a
		// single upgrade can't target every fork.
		expectedPatterns map[string]string
	}{
		{
			name: "the default age bounds the constraint at the version uv.lock resolves",
			expectedPatterns: map[string]string{
				"requests": ">=2.32.5, >=2.28",
				// Resolved from another registry, still the version in use
				"corp-utils": ">=1.4.0",
			},
		},
		{
			name: "a user versionfilter is bounded at the version uv.lock resolves",
			spec: Spec{VersionFilter: version.Filter{Kind: "semver", Pattern: "minor"}},
			expectedPatterns: map[string]string{
				"requests":   ">=2.32.5, 2.x",
				"corp-utils": ">=1.4.0, 1.x",
			},
		},
		{
			name: "a user pep440 versionfilter is bounded at the version uv.lock resolves",
			spec: Spec{VersionFilter: version.Filter{Kind: "pep440", Pattern: "<3"}},
			expectedPatterns: map[string]string{
				"requests":   ">=2.32.5, <3",
				"corp-utils": ">=1.4.0, <3",
			},
		},
		{
			name: "allowdowngrade keeps the constraint as declared",
			spec: Spec{Age: &age.AutodiscoverySpec{Spec: age.Spec{Minimum: "7d"}, AllowDowngrade: new(true)}},
			expectedPatterns: map[string]string{
				"requests":   ">=2.28",
				"corp-utils": "*",
			},
		},
		{
			name: "without an age filter, the constraint is kept as declared",
			spec: Spec{Age: &age.AutodiscoverySpec{}},
			expectedPatterns: map[string]string{
				"requests":   ">=2.28",
				"corp-utils": "*",
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New(tt.spec, "testdata/locked_versions", "", "")
			require.NoError(t, err)
			p.uvAvailable = true

			manifests, err := p.DiscoverManifests()
			require.NoError(t, err)
			assertSourcePatterns(t, manifests, tt.expectedPatterns)
		})
	}
}

func TestNewAgeNeverDowngradesPEP440Versions(t *testing.T) {
	testdata := []struct {
		name string
		spec Spec
		// expectedPatterns maps each dependency getting a manifest to the pattern of its source
		// versionfilter. A dependency missing from it must be skipped.
		expectedPatterns map[string]string
	}{
		{
			name: "local, post and pre-release versions bound the declared constraint",
			expectedPatterns: map[string]string{
				// A specifier can't name a local version
				"torch": ">=2.1.0, >=2.0",
				"foo":   ">=1.0.post1, >=1.0",
				// A pre-release in use doesn't make newer pre-releases match
				"bar": ">=2.0, >=1.0",
			},
		},
		{
			name: "a semver versionfilter skips the versions in use it can't parse",
			spec: Spec{VersionFilter: version.Filter{Kind: "semver", Pattern: "*"}},
			expectedPatterns: map[string]string{
				"torch": ">=2.1.0+cpu",
			},
		},
		{
			name: "allowdowngrade doesn't need a bound",
			spec: Spec{
				VersionFilter: version.Filter{Kind: "semver", Pattern: "*"},
				Age:           &age.AutodiscoverySpec{Spec: age.Spec{Minimum: "7d"}, AllowDowngrade: new(true)},
			},
			expectedPatterns: map[string]string{
				"torch": "*",
				"foo":   "*",
				"bar":   "*",
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New(tt.spec, "testdata/locked_pep440_versions", "", "")
			require.NoError(t, err)
			p.uvAvailable = true

			manifests, err := p.DiscoverManifests()
			require.NoError(t, err)
			assertSourcePatterns(t, manifests, tt.expectedPatterns)
		})
	}
}

func TestNewAgeAllowDowngrade(t *testing.T) {
	// "allowdowngrade" alone keeps the default age
	p, err := New(map[string]interface{}{"age": map[string]interface{}{"allowdowngrade": true}}, "testdata/no_lockfile", "", "")
	require.NoError(t, err)
	assert.Equal(t, defaultMinimumReleaseAge, p.ageFilter.Minimum)
	assert.True(t, p.ageFilter.MayDowngrade())

	// "age: {}" disables the default, so nothing is held back and nothing is bounded
	p, err = New(map[string]interface{}{"age": map[string]interface{}{}}, "testdata/no_lockfile", "", "")
	require.NoError(t, err)
	assert.True(t, p.ageFilter.IsZero())
	assert.False(t, p.ageFilter.NeverDowngrade())
}

func TestInstalledVersion(t *testing.T) {
	locked := discoveredFile{installed: lockedVersions{"requests": {"2.32.5"}, "numpy": {"2.1.0", "1.26.4", "10.0.0rc1"}}}

	testdata := []struct {
		name     string
		file     discoveredFile
		dep      pythonDependency
		expected string
	}{
		{name: "locked version", file: locked, dep: pythonDependency{Name: "Requests", Constraint: ">=2.28"}, expected: "2.32.5"},
		{name: "highest locked version", file: locked, dep: pythonDependency{Name: "numpy", Constraint: ">=1.20"}, expected: "10.0.0rc1"},
		{name: "highest locked version without constraint", file: locked, dep: pythonDependency{Name: "numpy"}, expected: "10.0.0rc1"},
		// A declaration under an environment marker only gets the versions within its constraint
		{name: "highest locked version within the constraint", file: locked, dep: pythonDependency{Name: "numpy", Constraint: ">=1.24,<2"}, expected: "1.26.4"},
		// The constraint then stays unbounded, even when pinned
		{name: "no locked version within the constraint", file: locked, dep: pythonDependency{Name: "requests", Constraint: "<2"}, expected: ""},
		{name: "no locked version within the pin", file: locked, dep: pythonDependency{Name: "requests", Constraint: "==2.0.0"}, expected: ""},
		// Unlike vulnerability mode, which then leaves the package out, see TestCurrentVersionWithUvLock
		{name: "package uv.lock doesn't resolve falls back to its pin", file: locked, dep: pythonDependency{Name: "click", Constraint: "==8.1.7"}, expected: "8.1.7"},
		{name: "without uv.lock, the pin is the version in use", dep: pythonDependency{Name: "click", Constraint: "==8.1.7"}, expected: "8.1.7"},
		{name: "without uv.lock nor pin, no version in use", dep: pythonDependency{Name: "click", Constraint: ">=8"}, expected: ""},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.file.installedVersion(tt.dep))
		})
	}
}

func TestRoutineVersionFilter(t *testing.T) {
	testdata := []struct {
		name string
		// filter is the user versionfilter, under the default age.
		filter     version.Filter
		dependency string
		installed  lockedVersions
		// expectedPattern is the source pattern, "" when the dependency must be skipped.
		expectedPattern string
	}{
		{
			// The declared version of "foo<3.0" is 3.0, an upper bound
			name:            "a pep440 filter isn't bounded at the declared version",
			filter:          version.Filter{Kind: "pep440", Pattern: "<3.0"},
			dependency:      "foo<3.0",
			expectedPattern: "<3.0",
		},
		{
			name:            "a pep440 filter is bounded at a pinned version",
			filter:          version.Filter{Kind: "pep440", Pattern: "<3.0"},
			dependency:      "foo==2.5.0",
			expectedPattern: ">=2.5.0, <3.0",
		},
		{
			name:       "a relative semver pattern without a version to start from is skipped",
			filter:     version.Filter{Kind: "semver", Pattern: "minor"},
			dependency: "foo<3.0",
		},
		{
			name:            "a semver pattern which isn't relative applies as is",
			filter:          version.Filter{Kind: "semver", Pattern: ">=1.0.0"},
			dependency:      "foo<3.0",
			expectedPattern: ">=1.0.0",
		},
		{
			name:            "a relative semver pattern starts from the constraint's lower bound",
			filter:          version.Filter{Kind: "semver", Pattern: "minor"},
			dependency:      "foo>=2.28,<3",
			expectedPattern: ">=2.28.0, 2.x",
		},
		{
			name:            "a post release lower bound starts from its release",
			filter:          version.Filter{Kind: "semver", Pattern: "minor"},
			dependency:      "foo>=2.28.post1",
			expectedPattern: ">=2.28.0, 2.x",
		},
		{
			name:       "a lower bound semver can't parse, such as one with an epoch, skips a relative pattern",
			filter:     version.Filter{Kind: "semver", Pattern: "minor"},
			dependency: "foo>=1!2.0",
		},
		{
			name:            "the installed version is the start",
			filter:          version.Filter{Kind: "semver", Pattern: "minor"},
			dependency:      "foo<3.0",
			installed:       lockedVersions{"foo": {"2.5.0"}},
			expectedPattern: ">=2.5.0, 2.x",
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New(Spec{VersionFilter: tt.filter}, "testdata/no_lockfile", "", "")
			require.NoError(t, err)
			require.True(t, p.ageFilter.NeverDowngrade())

			dep, err := parsePEP508(tt.dependency)
			require.NoError(t, err)

			kind, pattern, _, ok := p.routineVersionFilter(discoveredFile{installed: tt.installed}, dep)
			if tt.expectedPattern == "" {
				assert.False(t, ok, "the dependency must be skipped, not get the pattern %q", pattern)
				return
			}
			require.True(t, ok)
			assert.Equal(t, tt.filter.Kind, kind)
			assert.Equal(t, tt.expectedPattern, pattern)
		})
	}
}
