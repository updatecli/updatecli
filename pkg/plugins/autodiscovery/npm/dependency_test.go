package npm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
)

// TestRoutineVersionFilterLockOutOfSync checks that a locked version the package.json constraint
// doesn't accept, such as after editing package.json without updating the lock file, doesn't
// bound the constraint, which would leave no version to match.
func TestRoutineVersionFilterLockOutOfSync(t *testing.T) {
	locked := lockedVersions{byName: map[string]string{"axios": "1.3.0"}}

	n := Npm{ageFilter: age.AutodiscoverySpec{Spec: age.Spec{Minimum: "3d"}}}
	_, pattern, _ := n.routineVersionFilter("axios", "~1.2.0", "package.json", true, locked)
	assert.Equal(t, "~1.2.0", pattern)
}

func TestRoutineVersionFilterLowerBound(t *testing.T) {
	locked := lockedVersions{byName: map[string]string{"axios": "1.2.6"}}

	tests := []struct {
		name                    string
		dependencyVersion       string
		isVersionConstraint     bool
		ignoreVersionConstraint bool
		allowDowngrade          bool
		want                    string
	}{
		{
			name:                "range bounded at the locked version",
			dependencyVersion:   "^1.0.0",
			isVersionConstraint: true,
			want:                ">=1.2.6, ^1.0.0",
		},
		{
			name:                "range kept as is when a downgrade is allowed",
			dependencyVersion:   "^1.0.0",
			isVersionConstraint: true,
			allowDowngrade:      true,
			want:                "^1.0.0",
		},
		{
			name:                    "ignored range bounded at the locked version",
			dependencyVersion:       "^1.0.0",
			isVersionConstraint:     true,
			ignoreVersionConstraint: true,
			want:                    ">=1.2.6",
		},
		{
			name:                    "ignored range accepts any version when a downgrade is allowed",
			dependencyVersion:       "^1.0.0",
			isVersionConstraint:     true,
			ignoreVersionConstraint: true,
			allowDowngrade:          true,
			want:                    "*",
		},
		{
			name:              "exact version bounded at itself",
			dependencyVersion: "1.2.6",
			want:              ">=1.2.6",
		},
		{
			name:              "exact version accepts any version when a downgrade is allowed",
			dependencyVersion: "1.2.6",
			allowDowngrade:    true,
			want:              "*",
		},
		{
			name:              "exact prerelease keeps accepting prereleases when a downgrade is allowed",
			dependencyVersion: "2.0.0-rc.2",
			allowDowngrade:    true,
			want:              ">=0.0.0-0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := Npm{ignoreVersionConstraint: tt.ignoreVersionConstraint, ageFilter: age.AutodiscoverySpec{Spec: age.Spec{Minimum: "3d"}, AllowDowngrade: &tt.allowDowngrade}}

			kind, pattern, _ := n.routineVersionFilter("axios", tt.dependencyVersion, "package.json", tt.isVersionConstraint, locked)
			assert.Equal(t, "semver", kind)
			assert.Equal(t, tt.want, pattern)
		})
	}
}

// TestRoutineVersionFilterLockedPrerelease checks that a prerelease in the lock file bounds a
// range without making it accept prereleases.
func TestRoutineVersionFilterLockedPrerelease(t *testing.T) {
	locked := lockedVersions{byName: map[string]string{"axios": "1.5.0-beta.1"}}

	tests := []struct {
		name                    string
		ignoreVersionConstraint bool
		want                    string
	}{
		{
			name: "range bounded at the release of the locked prerelease",
			want: ">=1.5.0, ^1.2.0",
		},
		{
			name:                    "ignored range bounded at the release of the locked prerelease",
			ignoreVersionConstraint: true,
			want:                    ">=1.5.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := Npm{ignoreVersionConstraint: tt.ignoreVersionConstraint, ageFilter: age.AutodiscoverySpec{Spec: age.Spec{Minimum: "3d"}}}

			kind, pattern, _ := n.routineVersionFilter("axios", "^1.2.0", "package.json", true, locked)
			assert.Equal(t, "semver", kind)
			assert.Equal(t, tt.want, pattern)
		})
	}
}

// TestRoutineVersionFilterWithoutAge checks that without an age filter, such as with "age: {}",
// nothing holds back the newest versions, so the patterns are left as they were.
func TestRoutineVersionFilterWithoutAge(t *testing.T) {
	locked := lockedVersions{byName: map[string]string{"axios": "1.2.6"}}
	n := Npm{}

	_, pattern, _ := n.routineVersionFilter("axios", "^1.0.0", "package.json", true, locked)
	assert.Equal(t, "^1.0.0", pattern)

	n.ignoreVersionConstraint = true
	_, pattern, _ = n.routineVersionFilter("axios", "^1.0.0", "package.json", true, locked)
	assert.Equal(t, "*", pattern)
}

// TestRoutineVersionFilterPrereleaseRange checks that a prerelease locked within a range only
// accepting prereleases of a release, which rejects that release, still bounds the range.
func TestRoutineVersionFilterPrereleaseRange(t *testing.T) {
	locked := lockedVersions{byName: map[string]string{"axios": "1.5.0-beta.2"}}

	n := Npm{ageFilter: age.AutodiscoverySpec{Spec: age.Spec{Minimum: "3d"}}}
	_, pattern, _ := n.routineVersionFilter("axios", ">=1.5.0-beta.1 <1.5.0", "package.json", true, locked)
	assert.Equal(t, ">=1.5.0-beta.2, >=1.5.0-beta.1 <1.5.0", pattern)
}

func TestSatisfiesConstraint(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		want       bool
	}{
		{constraint: "^1.2.0", version: "1.5.0", want: true},
		// A range without prerelease is compared to the release of a prerelease
		{constraint: "^1.2.0", version: "1.5.0-beta.1", want: true},
		{constraint: ">=1.5.0-beta.1 <1.5.0", version: "1.5.0-beta.2", want: true},
		{constraint: "~1.2.0", version: "1.3.0", want: false},
		{constraint: "^1.2.0", version: "next", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.constraint+" "+tt.version, func(t *testing.T) {
			assert.Equal(t, tt.want, satisfiesConstraint(tt.constraint, tt.version))
		})
	}
}
