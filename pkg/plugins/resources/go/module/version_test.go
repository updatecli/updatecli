package gomodule

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

const testModule string = "example.com/mymodule"

// daysAgo returns a release date as reported by a Go module proxy.
func daysAgo(days int) string {
	return time.Now().Add(-time.Duration(days) * 24 * time.Hour).Format(time.RFC3339)
}

// goProxyStub serves the subset of the goproxy protocol used by that resource, as
// described on https://go.dev/ref/mod#goproxy-protocol
type goProxyStub struct {
	// publishedVersions is served by the "@v/list" endpoint.
	publishedVersions []string
	// releaseDates maps a version to its release date, served by "@v/<version>.info".
	releaseDates map[string]string
	// latestVersion is served by the "@latest" endpoint.
	latestVersion versionInfo
	// requestedPaths records every requested path, to assert on the requests really sent.
	requestedPaths []string
}

// start returns a GoModule querying that stub instead of a real Go module proxy.
func (s *goProxyStub) start(t *testing.T, spec Spec) *GoModule {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.requestedPaths = append(s.requestedPaths, r.URL.Path)

		switch {
		case strings.HasSuffix(r.URL.Path, "/@v/list"):
			fmt.Fprint(w, strings.Join(s.publishedVersions, "\n"))

		case strings.HasSuffix(r.URL.Path, "/@latest"):
			if s.latestVersion.Version == "" {
				http.NotFound(w, r)
				return
			}
			assert.NoError(t, json.NewEncoder(w).Encode(s.latestVersion))

		case strings.HasSuffix(r.URL.Path, ".info"):
			requestedVersion := strings.TrimSuffix(path.Base(r.URL.Path), ".info")
			releaseDate, ok := s.releaseDates[requestedVersion]
			if !ok {
				http.NotFound(w, r)
				return
			}
			assert.NoError(t, json.NewEncoder(w).Encode(versionInfo{
				Version: requestedVersion,
				Time:    releaseDate,
			}))

		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	spec.Proxy = server.URL
	spec.Module = testModule

	got, err := New(spec)
	require.NoError(t, err)
	got.webClient = http.DefaultClient

	return got
}

// infoRequests returns the ".info" paths requested to the stub.
func (s *goProxyStub) infoRequests() []string {
	requests := []string{}
	for _, p := range s.requestedPaths {
		if strings.HasSuffix(p, ".info") {
			requests = append(requests, p)
		}
	}

	return requests
}

func TestVersions(t *testing.T) {
	tests := []struct {
		name string
		spec Spec
		stub goProxyStub
		// expectedVersion is the version returned by versions()
		expectedVersion string
		// expectedHeldBackByAge is true when the age filter is expected to discard every version
		expectedHeldBackByAge bool
		// expectedError is true when versions() is expected to fail for another reason than the age filter
		expectedError bool
		// expectedInfoRequests is the number of ".info" requests expected on the proxy
		expectedInfoRequests int
	}{
		{
			name: "module without any published version falls back to its latest commit",
			spec: Spec{
				VersionFilter: version.Filter{Kind: version.LATESTVERSIONKIND},
				Age:           age.Spec{Minimum: "7d"},
			},
			stub: goProxyStub{
				latestVersion: versionInfo{
					Version: "v0.0.0-20260101120000-abcdefabcdef",
					Time:    daysAgo(60),
				},
			},
			expectedVersion: "v0.0.0-20260101120000-abcdefabcdef",
			// An empty "@v/list" must not be turned into a request for "@v/.info"
			expectedInfoRequests: 0,
		},
		{
			name: "cooldown holds back the only commit of a module without any published version",
			spec: Spec{
				VersionFilter: version.Filter{Kind: version.LATESTVERSIONKIND},
				Age:           age.Spec{Minimum: "7d"},
			},
			stub: goProxyStub{
				latestVersion: versionInfo{
					Version: "v0.0.0-20260101120000-abcdefabcdef",
					Time:    daysAgo(1),
				},
			},
			expectedHeldBackByAge: true,
			expectedInfoRequests:  0,
		},
		{
			name: "latest endpoint answering a tagged version is accepted without any age filter",
			spec: Spec{
				VersionFilter: version.Filter{Kind: version.LATESTVERSIONKIND},
			},
			stub: goProxyStub{
				latestVersion: versionInfo{Version: "v1.2.3", Time: daysAgo(60)},
			},
			expectedVersion:      "v1.2.3",
			expectedInfoRequests: 0,
		},
		{
			name: "cooldown holds back every published version instead of using the latest commit",
			spec: Spec{
				VersionFilter: version.Filter{Kind: version.LATESTVERSIONKIND},
				Age:           age.Spec{Minimum: "7d"},
			},
			stub: goProxyStub{
				publishedVersions: []string{"v1.0.0", "v1.1.0"},
				releaseDates: map[string]string{
					"v1.0.0": daysAgo(2),
					"v1.1.0": daysAgo(1),
				},
				// A module publishing versions must never fall back to a pseudo version
				latestVersion: versionInfo{
					Version: "v0.0.0-20260101120000-abcdefabcdef",
					Time:    daysAgo(60),
				},
			},
			expectedHeldBackByAge: true,
			expectedInfoRequests:  2,
		},
		{
			name: "latest filter falls back to the most recently published version out of cooldown",
			spec: Spec{
				VersionFilter: version.Filter{Kind: version.LATESTVERSIONKIND},
				Age:           age.Spec{Minimum: "7d"},
			},
			stub: goProxyStub{
				publishedVersions: []string{"v1.9.0", "v1.10.0", "v1.11.0"},
				releaseDates: map[string]string{
					"v1.9.0":  daysAgo(50),
					"v1.10.0": daysAgo(10),
					"v1.11.0": daysAgo(1),
				},
			},
			// "v1.9.0" is the lexicographic maximum, "v1.10.0" the most recently published
			expectedVersion:      "v1.10.0",
			expectedInfoRequests: 3,
		},
		{
			name: "semver filter keeps ordering versions semantically",
			spec: Spec{
				VersionFilter: version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: "*"},
				Age:           age.Spec{Minimum: "7d"},
			},
			stub: goProxyStub{
				publishedVersions: []string{"v1.9.0", "v1.10.0", "v1.11.0"},
				releaseDates: map[string]string{
					"v1.9.0":  daysAgo(50),
					"v1.10.0": daysAgo(10),
					"v1.11.0": daysAgo(1),
				},
			},
			expectedVersion:      "v1.10.0",
			expectedInfoRequests: 3,
		},
		{
			// The autodiscovery bounds the filter at the version in go.mod, so while that
			// version is too recent, only older versions are left
			name: "cooldown holds back every version matching the filter while older ones pass",
			spec: Spec{
				VersionFilter: version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: ">=1.10.0"},
				Age:           age.Spec{Minimum: "7d"},
			},
			stub: goProxyStub{
				publishedVersions: []string{"v1.9.0", "v1.10.0", "v1.11.0"},
				releaseDates: map[string]string{
					"v1.9.0":  daysAgo(50),
					"v1.10.0": daysAgo(3),
					"v1.11.0": daysAgo(1),
				},
			},
			expectedHeldBackByAge: true,
			expectedInfoRequests:  3,
		},
		{
			name: "a version filter matching no published version still fails",
			spec: Spec{
				VersionFilter: version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: ">=2.0.0"},
				Age:           age.Spec{Minimum: "7d"},
			},
			stub: goProxyStub{
				publishedVersions: []string{"v1.9.0", "v1.10.0"},
				releaseDates: map[string]string{
					"v1.9.0":  daysAgo(50),
					"v1.10.0": daysAgo(1),
				},
			},
			expectedError:        true,
			expectedInfoRequests: 2,
		},
		{
			name: "a version filter matching no published version still fails while every version is in cooldown",
			spec: Spec{
				VersionFilter: version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: ">=2.0.0"},
				Age:           age.Spec{Minimum: "7d"},
			},
			stub: goProxyStub{
				publishedVersions: []string{"v1.9.0", "v1.10.0"},
				releaseDates: map[string]string{
					"v1.9.0":  daysAgo(2),
					"v1.10.0": daysAgo(1),
				},
			},
			expectedError:        true,
			expectedInfoRequests: 2,
		},
		{
			// The age filter can't be applied, which skips like a cooldown rather than fails
			name: "versions without any release date are held back",
			spec: Spec{
				VersionFilter: version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: "*"},
				Age:           age.Spec{Minimum: "7d"},
			},
			stub: goProxyStub{
				publishedVersions: []string{"v1.9.0", "v1.10.0"},
			},
			expectedHeldBackByAge: true,
			expectedInfoRequests:  2,
		},
		{
			// Like a cooldown, a maximum age discarding every match is an age hold, not a failure
			name: "a version filter matching only versions above the maximum age is held back",
			spec: Spec{
				VersionFilter: version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: "~1.9"},
				Age:           age.Spec{Minimum: "7d", Maximum: "1y"},
			},
			stub: goProxyStub{
				publishedVersions: []string{"v1.9.0", "v1.10.0"},
				releaseDates: map[string]string{
					"v1.9.0":  daysAgo(400),
					"v1.10.0": daysAgo(10),
				},
			},
			expectedHeldBackByAge: true,
			expectedInfoRequests:  2,
		},
		{
			// v1.9.0 is too recent and v1.10.0, the only version matching the filter, has no
			// release date, so no version passes the age filter, which skips
			name: "a version filter matching only a version without release date is held back while no version passes the age filter",
			spec: Spec{
				VersionFilter: version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: ">=1.10.0"},
				Age:           age.Spec{Minimum: "7d"},
			},
			stub: goProxyStub{
				publishedVersions: []string{"v1.9.0", "v1.10.0"},
				releaseDates: map[string]string{
					"v1.9.0": daysAgo(2),
				},
			},
			expectedHeldBackByAge: true,
			expectedInfoRequests:  2,
		},
		{
			// A version whose release date can't be retrieved isn't known to be cooling down
			name: "a version filter matching only a version without release date still fails",
			spec: Spec{
				VersionFilter: version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: ">=1.10.0"},
				Age:           age.Spec{Minimum: "7d"},
			},
			stub: goProxyStub{
				publishedVersions: []string{"v1.9.0", "v1.10.0"},
				releaseDates: map[string]string{
					"v1.9.0": daysAgo(50),
				},
			},
			expectedError:        true,
			expectedInfoRequests: 2,
		},
		{
			name: "no age filter means no release date lookup",
			spec: Spec{
				VersionFilter: version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: "*"},
			},
			stub: goProxyStub{
				publishedVersions: []string{"v1.9.0", "v1.10.0"},
			},
			expectedVersion:      "v1.10.0",
			expectedInfoRequests: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stub.start(t, tt.spec)

			gotVersion, err := got.versions(context.Background())

			switch {
			case tt.expectedHeldBackByAge:
				require.ErrorIs(t, err, age.ErrNoVersionMatchingAge)
				assert.Empty(t, gotVersion)
			case tt.expectedError:
				require.Error(t, err)
				assert.NotErrorIs(t, err, age.ErrNoVersionMatchingAge)
			default:
				require.NoError(t, err)
				assert.Equal(t, tt.expectedVersion, gotVersion)
			}

			assert.Len(t, tt.stub.infoRequests(), tt.expectedInfoRequests)
		})
	}
}

// TestVersionsTriesTheNextProxyWithoutReleaseDates checks that a proxy failing to date every
// version, such as when rate limited, doesn't prevent the next proxy from being queried.
func TestVersionsTriesTheNextProxyWithoutReleaseDates(t *testing.T) {
	spec := Spec{
		VersionFilter: version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: "*"},
		Age:           age.Spec{Minimum: "7d"},
	}

	undated := &goProxyStub{publishedVersions: []string{"v1.9.0", "v1.10.0"}}
	dated := &goProxyStub{
		publishedVersions: []string{"v1.9.0", "v1.10.0"},
		releaseDates: map[string]string{
			"v1.9.0":  daysAgo(50),
			"v1.10.0": daysAgo(10),
		},
	}

	first := undated.start(t, spec)
	got := dated.start(t, spec)
	got.Spec.Proxy = first.Spec.Proxy + "," + got.Spec.Proxy

	gotVersion, err := got.versions(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "v1.10.0", gotVersion)
	assert.Len(t, undated.infoRequests(), 2)

	// Without another proxy, the missing release dates skip the source, as a cooldown does
	gotVersion, err = first.versions(context.Background())
	require.ErrorIs(t, err, age.ErrNoVersionMatchingAge)
	assert.ErrorContains(t, err, "no release date")
	assert.Empty(t, gotVersion)
}

// TestVersionsKeepsAnAgeHoldFromAPreviousProxy checks that a proxy without any version matching
// the filter, such as a stale mirror, doesn't turn the age hold of a previous proxy into a failure.
func TestVersionsKeepsAnAgeHoldFromAPreviousProxy(t *testing.T) {
	spec := Spec{
		VersionFilter: version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: ">=1.10.0"},
		Age:           age.Spec{Minimum: "7d"},
	}

	recent := &goProxyStub{
		publishedVersions: []string{"v1.9.0", "v1.10.0"},
		releaseDates: map[string]string{
			"v1.9.0":  daysAgo(50),
			"v1.10.0": daysAgo(2),
		},
	}
	stale := &goProxyStub{
		publishedVersions: []string{"v1.9.0"},
		releaseDates: map[string]string{
			"v1.9.0": daysAgo(50),
		},
	}

	first := recent.start(t, spec)
	got := stale.start(t, spec)
	got.Spec.Proxy = first.Spec.Proxy + "," + got.Spec.Proxy

	gotVersion, err := got.versions(context.Background())
	require.ErrorIs(t, err, age.ErrNoVersionMatchingAge)
	assert.Empty(t, gotVersion)
	assert.NotEmpty(t, stale.requestedPaths)
}

func TestNewVersionFilter(t *testing.T) {
	tests := []struct {
		name             string
		versionFilter    version.Filter
		expectedFilter   version.Filter
		expectedNewErr   bool
		expectedIsLatest bool
	}{
		{
			name:             "no filter falls back to semantic versioning",
			expectedFilter:   version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: "*"},
			expectedIsLatest: true,
		},
		{
			// A "latest" kind without any pattern used to be searched as the literal
			// pattern "" and never matched anything.
			name:             "latest kind without pattern gets its default pattern",
			versionFilter:    version.Filter{Kind: version.LATESTVERSIONKIND},
			expectedFilter:   version.Filter{Kind: version.LATESTVERSIONKIND, Pattern: version.LATESTVERSIONKIND},
			expectedIsLatest: true,
		},
		{
			name:             "semver kind without pattern gets its default pattern",
			versionFilter:    version.Filter{Kind: version.SEMVERVERSIONKIND},
			expectedFilter:   version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: "*"},
			expectedIsLatest: true,
		},
		{
			name:           "explicit filter is left untouched",
			versionFilter:  version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: "1.0.x"},
			expectedFilter: version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: "1.0.x"},
		},
		{
			name:           "unsupported kind is rejected",
			versionFilter:  version.Filter{Kind: "notakind"},
			expectedNewErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(Spec{Module: testModule, VersionFilter: tt.versionFilter})
			if tt.expectedNewErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedFilter, got.versionFilter)
			assert.Equal(t, tt.expectedIsLatest, isLatestVersionFilter(got.versionFilter))
		})
	}
}
