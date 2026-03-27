package temurin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestApiGetReleaseNames_FallbackOnFeature verifies that when the most recent
// feature release (26) has no GA builds indexed, the plugin falls back to the
// next available release (25) and returns its release names.
func TestApiGetReleaseNames_FallbackOnFeature(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc(availableReleasesEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(apiInfoReleases{
			MostRecentFeatureRelease: 26,
			MostRecentLTS:            25,
			AvailableReleases:        []int{8, 11, 17, 21, 25, 26},
			AvailableLTSReleases:     []int{8, 11, 17, 21, 25},
		})
	})

	mux.HandleFunc(releaseNamesEndpoint, func(w http.ResponseWriter, r *http.Request) {
		// The version range for major N is "(N.0.0, N+1.0.0]", so we match on
		// the range start to avoid false positives (e.g. the range for 25
		// contains "26" as the upper bound).
		versionParam := r.URL.Query().Get("version")
		switch {
		case strings.HasPrefix(versionParam, "(26."):
			// Feature release 26 has no GA builds yet.
			http.Error(w, "not found", http.StatusNotFound)
		case strings.HasPrefix(versionParam, "(25."):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(releaseInformation{
				Releases: []string{"jdk-25.0.2+10"},
			})
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	})

	mux.HandleFunc(architecturesEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]string{"x64"})
	})

	mux.HandleFunc(osEndpoints, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]string{"linux"})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	sut := &Temurin{
		spec: Spec{
			ReleaseLine:     "feature",
			ReleaseType:     "ga",
			Architecture:    "x64",
			OperatingSystem: "linux",
			ImageType:       "jdk",
			Project:         "jdk",
			Result:          "version",
		},
		apiURL: server.URL,
	}
	sut.apiWebClient = server.Client()
	sut.apiWebRedirectionClient = server.Client()

	releases, err := sut.apiGetReleaseNames(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []string{"jdk-25.0.2+10"}, releases)
}

// TestApiGetReleaseNames_FallbackOnLTS verifies that when the most recent LTS
// release (25) has no GA builds indexed, the plugin falls back to the next
// available LTS (21) and returns its release names.
func TestApiGetReleaseNames_FallbackOnLTS(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc(availableReleasesEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(apiInfoReleases{
			MostRecentLTS:            25,
			MostRecentFeatureRelease: 26,
			AvailableReleases:        []int{8, 11, 17, 21, 25, 26},
			AvailableLTSReleases:     []int{8, 11, 17, 21, 25},
		})
	})

	mux.HandleFunc(releaseNamesEndpoint, func(w http.ResponseWriter, r *http.Request) {
		versionParam := r.URL.Query().Get("version")
		switch {
		case strings.HasPrefix(versionParam, "(25."):
			http.Error(w, "not found", http.StatusNotFound)
		case strings.HasPrefix(versionParam, "(21."):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(releaseInformation{
				Releases: []string{"jdk-21.0.6+7"},
			})
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	})

	mux.HandleFunc(architecturesEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]string{"x64"})
	})

	mux.HandleFunc(osEndpoints, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]string{"linux"})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	sut := &Temurin{
		spec: Spec{
			ReleaseLine:     "",
			ReleaseType:     "ga",
			Architecture:    "x64",
			OperatingSystem: "linux",
			ImageType:       "jdk",
			Project:         "jdk",
			Result:          "version",
		},
		apiURL: server.URL,
	}
	sut.apiWebClient = server.Client()
	sut.apiWebRedirectionClient = server.Client()

	releases, err := sut.apiGetReleaseNames(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []string{"jdk-21.0.6+7"}, releases)
}

// TestApiGetReleaseNames_AllCandidatesFail verifies that when every candidate
// version returns an API error the first error is propagated to the caller.
func TestApiGetReleaseNames_AllCandidatesFail(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc(availableReleasesEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(apiInfoReleases{
			MostRecentFeatureRelease: 26,
			MostRecentLTS:            25,
			AvailableReleases:        []int{8, 11, 17, 21, 25, 26},
			AvailableLTSReleases:     []int{8, 11, 17, 21, 25},
		})
	})

	mux.HandleFunc(releaseNamesEndpoint, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	})

	mux.HandleFunc(architecturesEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]string{"x64"})
	})

	mux.HandleFunc(osEndpoints, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]string{"linux"})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	sut := &Temurin{
		spec: Spec{
			ReleaseLine:     "feature",
			ReleaseType:     "ga",
			Architecture:    "x64",
			OperatingSystem: "linux",
			ImageType:       "jdk",
			Project:         "jdk",
			Result:          "version",
		},
		apiURL: server.URL,
	}
	sut.apiWebClient = server.Client()
	sut.apiWebRedirectionClient = server.Client()

	releases, err := sut.apiGetReleaseNames(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Empty(t, releases)
}

// TestApiGetReleaseNames_ExplicitFeatureVersion verifies that when FeatureVersion
// is explicitly set in Spec the available-releases endpoint is never consulted and
// only that single version is queried.
func TestApiGetReleaseNames_ExplicitFeatureVersion(t *testing.T) {
	mux := http.NewServeMux()

	// If this endpoint is hit the test should still pass, but we register a 500
	// so that any accidental call surfaces as a test failure.
	mux.HandleFunc(availableReleasesEndpoint, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "should not be called", http.StatusInternalServerError)
	})

	mux.HandleFunc(releaseNamesEndpoint, func(w http.ResponseWriter, r *http.Request) {
		versionParam := r.URL.Query().Get("version")
		switch {
		case strings.HasPrefix(versionParam, "(21."):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(releaseInformation{
				Releases: []string{"jdk-21.0.6+7"},
			})
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	})

	mux.HandleFunc(architecturesEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]string{"x64"})
	})

	mux.HandleFunc(osEndpoints, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]string{"linux"})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	sut := &Temurin{
		spec: Spec{
			FeatureVersion:  21,
			ReleaseLine:     "feature",
			ReleaseType:     "ga",
			Architecture:    "x64",
			OperatingSystem: "linux",
			ImageType:       "jdk",
			Project:         "jdk",
			Result:          "version",
		},
		apiURL: server.URL,
	}
	sut.apiWebClient = server.Client()
	sut.apiWebRedirectionClient = server.Client()

	releases, err := sut.apiGetReleaseNames(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []string{"jdk-21.0.6+7"}, releases)
}

// TestApiGetReleaseNames_EmptyReleasesTriggersNextCandidate verifies that an
// HTTP 200 response carrying an empty releases list is treated as "no results"
// and causes the plugin to fall back to the next candidate version.
func TestApiGetReleaseNames_EmptyReleasesTriggersNextCandidate(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc(availableReleasesEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(apiInfoReleases{
			MostRecentFeatureRelease: 26,
			MostRecentLTS:            25,
			AvailableReleases:        []int{8, 11, 17, 21, 25, 26},
			AvailableLTSReleases:     []int{8, 11, 17, 21, 25},
		})
	})

	mux.HandleFunc(releaseNamesEndpoint, func(w http.ResponseWriter, r *http.Request) {
		versionParam := r.URL.Query().Get("version")
		switch {
		case strings.HasPrefix(versionParam, "(26."):
			// 200 with empty releases — should trigger fallback.
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(releaseInformation{Releases: []string{}})
		case strings.HasPrefix(versionParam, "(25."):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(releaseInformation{
				Releases: []string{"jdk-25.0.2+10"},
			})
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	})

	mux.HandleFunc(architecturesEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]string{"x64"})
	})

	mux.HandleFunc(osEndpoints, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]string{"linux"})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	sut := &Temurin{
		spec: Spec{
			ReleaseLine:     "feature",
			ReleaseType:     "ga",
			Architecture:    "x64",
			OperatingSystem: "linux",
			ImageType:       "jdk",
			Project:         "jdk",
			Result:          "version",
		},
		apiURL: server.URL,
	}
	sut.apiWebClient = server.Client()
	sut.apiWebRedirectionClient = server.Client()

	releases, err := sut.apiGetReleaseNames(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []string{"jdk-25.0.2+10"}, releases)
}

// TestApiGetReleaseNames_DescendingAvailableReleases verifies that the fallback
// ordering is correct even when AvailableReleases is returned in descending order
// by the API. The plugin should still attempt the highest version after the
// primary and not start from the lowest.
func TestApiGetReleaseNames_DescendingAvailableReleases(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc(availableReleasesEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Deliberately descending — the sort fix must normalise this.
		_ = json.NewEncoder(w).Encode(apiInfoReleases{
			MostRecentFeatureRelease: 26,
			MostRecentLTS:            25,
			AvailableReleases:        []int{26, 25, 21, 17, 11, 8},
			AvailableLTSReleases:     []int{25, 21, 17, 11, 8},
		})
	})

	mux.HandleFunc(releaseNamesEndpoint, func(w http.ResponseWriter, r *http.Request) {
		versionParam := r.URL.Query().Get("version")
		switch {
		case strings.HasPrefix(versionParam, "(26."):
			http.Error(w, "not found", http.StatusNotFound)
		case strings.HasPrefix(versionParam, "(25."):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(releaseInformation{
				Releases: []string{"jdk-25.0.2+10"},
			})
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	})

	mux.HandleFunc(architecturesEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]string{"x64"})
	})

	mux.HandleFunc(osEndpoints, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]string{"linux"})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	sut := &Temurin{
		spec: Spec{
			ReleaseLine:     "feature",
			ReleaseType:     "ga",
			Architecture:    "x64",
			OperatingSystem: "linux",
			ImageType:       "jdk",
			Project:         "jdk",
			Result:          "version",
		},
		apiURL: server.URL,
	}
	sut.apiWebClient = server.Client()
	sut.apiWebRedirectionClient = server.Client()

	releases, err := sut.apiGetReleaseNames(context.Background())

	// The fallback must have tried 25 (next highest after 26), not 8.
	require.NoError(t, err)
	assert.Equal(t, []string{"jdk-25.0.2+10"}, releases)
}
