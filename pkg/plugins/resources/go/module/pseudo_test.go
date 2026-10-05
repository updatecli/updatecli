package gomodule

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

const (
	testZeroPseudoVersion  string = "v0.0.0-20250101120000-abcdefabcdef"
	testPseudoVersion      string = "v1.2.4-0.20250101120000-abcdefabcdef"
	testNewerPseudoVersion string = "v0.0.0-20260101120000-abcdefabcdef"
)

func TestVersionsFromPseudoVersion(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		stub    goProxyStub
		// expectedVersion is the version returned by versions()
		expectedVersion string
		// expectedNoNewerVersion is true when nothing newer than the pseudo version is expected
		expectedNoNewerVersion bool
	}{
		{
			name:                   "pseudo version ahead of every published tag isn't downgraded",
			pattern:                ">=" + testPseudoVersion[1:],
			stub:                   goProxyStub{publishedVersions: []string{"v1.2.2", "v1.2.3"}},
			expectedNoNewerVersion: true,
		},
		{
			name:            "release is preferred over a prerelease",
			pattern:         ">=" + testPseudoVersion[1:] + ", 1.2.x-0",
			stub:            goProxyStub{publishedVersions: []string{"v1.2.3", "v1.2.4", "v1.2.5-rc.1", "v1.3.0"}},
			expectedVersion: "v1.2.4",
		},
		{
			name:            "prerelease is used when no release matches",
			pattern:         ">=" + testPseudoVersion[1:],
			stub:            goProxyStub{publishedVersions: []string{"v1.2.3", "v1.3.0-rc.1"}},
			expectedVersion: "v1.3.0-rc.1",
		},
		{
			name:            "releases are ordered semantically",
			pattern:         ">=" + testZeroPseudoVersion[1:],
			stub:            goProxyStub{publishedVersions: []string{"v0.10.0", "v0.9.0"}},
			expectedVersion: "v0.10.0",
		},
		{
			name:            "module without any published version follows its latest commit",
			pattern:         ">=" + testZeroPseudoVersion[1:],
			stub:            goProxyStub{latestVersion: versionInfo{Version: testNewerPseudoVersion}},
			expectedVersion: testNewerPseudoVersion,
		},
		{
			name:            "majoronly ignores minor updates of a pseudo version",
			pattern:         ">=" + testPseudoVersion[1:] + ", >1",
			stub:            goProxyStub{publishedVersions: []string{"v1.2.4", "v1.3.0", "v2.0.0+incompatible"}},
			expectedVersion: "v2.0.0+incompatible",
		},
		{
			name:                   "majoronly without a newer major version",
			pattern:                ">=" + testPseudoVersion[1:] + ", >1",
			stub:                   goProxyStub{publishedVersions: []string{"v1.2.4", "v1.3.0"}},
			expectedNoNewerVersion: true,
		},
		{
			name:                   "majoronly ignores newer commits of a module without any published version",
			pattern:                ">=" + testZeroPseudoVersion[1:] + ", >0",
			stub:                   goProxyStub{latestVersion: versionInfo{Version: testNewerPseudoVersion}},
			expectedNoNewerVersion: true,
		},
		{
			name:            "minoronly ignores patch updates of a pseudo version",
			pattern:         testPseudoVersion + " || >=" + testPseudoVersion[1:] + ", >1.2.x-0 < 2",
			stub:            goProxyStub{publishedVersions: []string{"v1.2.4", "v1.2.5", "v1.3.0", "v2.0.0+incompatible"}},
			expectedVersion: "v1.3.0",
		},
		{
			name:                   "module without any published version and no newer commit",
			pattern:                ">=" + testZeroPseudoVersion[1:],
			stub:                   goProxyStub{latestVersion: versionInfo{Version: "v0.0.0-20240101120000-abcdefabcdef"}},
			expectedNoNewerVersion: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.stub.start(t, Spec{
				VersionFilter: version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: tt.pattern},
			})

			gotVersion, err := g.versions(context.Background())
			if tt.expectedNoNewerVersion {
				require.ErrorIs(t, err, ErrNoVersionNewerThanPseudo)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedVersion, gotVersion)
		})
	}
}

func TestVersionsFromPseudoVersionQueriesEveryProxy(t *testing.T) {
	spec := Spec{
		VersionFilter: version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: ">=" + testPseudoVersion[1:]},
	}

	// The first proxy only serves an older tag, the second one already serves a newer one.
	staleStub := goProxyStub{publishedVersions: []string{"v1.2.3"}}
	stale := staleStub.start(t, spec)

	upToDateStub := goProxyStub{publishedVersions: []string{"v1.2.3", "v1.2.4"}}
	g := upToDateStub.start(t, spec)
	g.Spec.Proxy = stale.Spec.Proxy + "," + g.Spec.Proxy

	gotVersion, err := g.versions(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "v1.2.4", gotVersion)
}

func TestVersionsWithoutPseudoVersionStillFails(t *testing.T) {
	stub := goProxyStub{publishedVersions: []string{"v1.2.3"}}
	g := stub.start(t, Spec{
		VersionFilter: version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: ">=2.0.0"},
	})

	_, err := g.versions(context.Background())
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrNoVersionNewerThanPseudo)
}

func TestVersionsLatestKindOrdersSemantically(t *testing.T) {
	stub := goProxyStub{publishedVersions: []string{"v1.10.0", "v1.9.0"}}
	g := stub.start(t, Spec{
		VersionFilter: version.Filter{Kind: version.LATESTVERSIONKIND},
	})

	gotVersion, err := g.versions(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "v1.10.0", gotVersion)
}

func TestSourceSkipsWithoutVersionNewerThanPseudo(t *testing.T) {
	stub := goProxyStub{publishedVersions: []string{"v1.2.3"}}
	g := stub.start(t, Spec{
		VersionFilter: version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: ">=" + testPseudoVersion[1:]},
	})

	gotResult := result.Source{}
	require.NoError(t, g.Source(context.Background(), pathresolver.Resolver{}, &gotResult))
	assert.Equal(t, result.SKIPPED, gotResult.Result)
}

func TestPseudoVersionLowerBound(t *testing.T) {
	tests := []struct {
		name     string
		filter   version.Filter
		expected string
	}{
		{
			name:     "lower bound",
			filter:   version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: ">=" + testZeroPseudoVersion[1:]},
			expected: testZeroPseudoVersion,
		},
		{
			name:     "lower bound combined with an upper bound",
			filter:   version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: ">=" + testPseudoVersion[1:] + ", 1.2.x-0"},
			expected: testPseudoVersion,
		},
		{
			name:     "minoronly pattern",
			filter:   version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: testPseudoVersion + " || >1.2.x-0 < 2"},
			expected: testPseudoVersion,
		},
		{
			name:     "majoronly pattern",
			filter:   version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: testPseudoVersion + " || >" + testPseudoVersion},
			expected: testPseudoVersion,
		},
		{
			name:   "upper bound only",
			filter: version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: "<" + testPseudoVersion[1:]},
		},
		{
			name:   "upper bound separated from its version by a space",
			filter: version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: "<= " + testZeroPseudoVersion[1:]},
		},
		{
			name:     "lower bound separated from its version by a space",
			filter:   version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: ">= " + testZeroPseudoVersion[1:]},
			expected: testZeroPseudoVersion,
		},
		{
			name:   "regular version",
			filter: version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: ">=1.2.3"},
		},
		{
			name:   "not a semver filter",
			filter: version.Filter{Kind: version.REGEXVERSIONKIND, Pattern: ">=" + testZeroPseudoVersion[1:]},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, pseudoVersionLowerBound(tt.filter))
		})
	}
}
