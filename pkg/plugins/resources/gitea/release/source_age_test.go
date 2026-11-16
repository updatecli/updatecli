package release

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// giteaReleasesStub serves the releases endpoint of the Gitea API, so the age filter can be
// tested against release dates that don't change over time.
func giteaReleasesStub(t *testing.T, releaseAges map[string]time.Duration) string {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/repos/owner/repository/releases") {
			http.NotFound(w, r)
			return
		}

		releases := []map[string]interface{}{}
		for tag, releaseAge := range releaseAges {
			releases = append(releases, map[string]interface{}{
				"tag_name":     tag,
				"published_at": time.Now().Add(-releaseAge).Format(time.RFC3339),
			})
		}
		assert.NoError(t, json.NewEncoder(w).Encode(releases))
	}))
	t.Cleanup(server.Close)

	return server.URL
}

func TestSourceAgeHold(t *testing.T) {
	day := 24 * time.Hour

	tests := []struct {
		name        string
		pattern     string
		maximum     string
		releaseAges map[string]time.Duration
		wantSkipped bool
		wantErr     bool
	}{
		{
			// The autodiscovery bounds the filter at the version in use, so while that
			// version is too recent, only older releases are left
			name:        "skipped while every release matching the version filter is in cooldown",
			pattern:     ">=2.0.0",
			wantSkipped: true,
		},
		{
			name:    "a version filter matching no release still fails",
			pattern: ">=3.0.0",
			wantErr: true,
		},
		{
			name:    "a version filter matching no release still fails while every release is in cooldown",
			pattern: ">=3.0.0",
			releaseAges: map[string]time.Duration{
				"v1.0.0": 2 * day,
				"v2.0.0": 1 * day,
			},
			wantErr: true,
		},
		{
			name:    "skipped while every release is in cooldown",
			pattern: "*",
			releaseAges: map[string]time.Duration{
				"v1.0.0": 2 * day,
				"v2.0.0": 1 * day,
			},
			wantSkipped: true,
		},
		{
			// Like a cooldown, a maximum age discarding every match is an age hold, not a failure
			name:    "skipped while every release matching the version filter is above the maximum age",
			pattern: "<2.0.0",
			maximum: "20d",
			releaseAges: map[string]time.Duration{
				"v1.0.0": 30 * day,
				"v2.0.0": 10 * day,
			},
			wantSkipped: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			releaseAges := tt.releaseAges
			if releaseAges == nil {
				releaseAges = map[string]time.Duration{
					"v1.0.0": 30 * day,
					"v2.0.0": 2 * day,
					"v2.1.0": 1 * day,
				}
			}
			url := giteaReleasesStub(t, releaseAges)

			g, err := New(struct {
				URL           string
				Owner         string
				Repository    string
				VersionFilter version.Filter
				Age           age.Spec
			}{
				URL:           url,
				Owner:         "owner",
				Repository:    "repository",
				VersionFilter: version.Filter{Kind: "semver", Pattern: tt.pattern},
				Age:           age.Spec{Minimum: "7d", Maximum: tt.maximum},
			})
			require.NoError(t, err)

			gotResult := result.Source{}
			err = g.Source(context.Background(), pathresolver.Resolver{}, &gotResult)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantSkipped, gotResult.Result == result.SKIPPED)
		})
	}
}
