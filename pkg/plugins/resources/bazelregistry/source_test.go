package bazelregistry

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestSource(t *testing.T) {
	tests := []struct {
		name           string
		spec           Spec
		mockResponse   string
		mockStatusCode int
		mockError      error
		wantErr        bool
		wantVersion    string
		wantResult     string
	}{
		{
			name: "Success - latest version (no filter)",
			spec: Spec{
				Module: "rules_go",
			},
			mockResponse:   testMetadataJSON,
			mockStatusCode: http.StatusOK,
			wantErr:        false,
			wantVersion:    "0.52.0", // Latest version, excluding yanked
			wantResult:     result.SUCCESS,
		},
		{
			name: "Success - semver filter",
			spec: Spec{
				Module: "rules_go",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "~0.51",
				},
			},
			mockResponse:   testMetadataJSON,
			mockStatusCode: http.StatusOK,
			wantErr:        false,
			wantVersion:    "0.51.0", // Latest matching version
			wantResult:     result.SUCCESS,
		},
		{
			name: "Success - regex filter",
			spec: Spec{
				Module: "rules_go",
				VersionFilter: version.Filter{
					Kind:    "regex",
					Pattern: "^0\\.5[01]",
				},
			},
			mockResponse:   testMetadataJSON,
			mockStatusCode: http.StatusOK,
			wantErr:        false,
			wantVersion:    "0.51.0", // Latest matching version
			wantResult:     result.SUCCESS,
		},
		{
			name: "Error - module not found",
			spec: Spec{
				Module: "nonexistent",
			},
			mockResponse:   "",
			mockStatusCode: http.StatusNotFound,
			wantErr:        true,
		},
		{
			name: "Error - all versions yanked",
			spec: Spec{
				Module: "rules_go",
			},
			mockResponse: `{
				"versions": ["1.0.0"],
				"yanked_versions": {
					"1.0.0": "All versions yanked"
				}
			}`,
			mockStatusCode: http.StatusOK,
			wantErr:        true,
		},
		{
			name: "Error - network error",
			spec: Spec{
				Module: "rules_go",
			},
			mockResponse:   "",
			mockStatusCode: 0,
			mockError:      http.ErrServerClosed,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := New(tt.spec)
			require.NoError(t, err)

			b.webClient = &httpclient.MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return &http.Response{
						StatusCode: tt.mockStatusCode,
						Body:       io.NopCloser(strings.NewReader(tt.mockResponse)),
					}, nil
				},
			}

			resultSource := &result.Source{}
			err = b.Source("", resultSource)

			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, result.FAILURE, resultSource.Result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, resultSource.Result)
				assert.Equal(t, tt.wantVersion, resultSource.Information)
				assert.Contains(t, resultSource.Description, tt.wantVersion)
			}
		})
	}
}

func TestSource_YankedVersions(t *testing.T) {
	// Test that yanked versions are excluded
	b, err := New(Spec{
		Module: "rules_go",
	})
	require.NoError(t, err)

	b.webClient = &httpclient.MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(testMetadataJSON)),
			}, nil
		},
	}

	resultSource := &result.Source{}
	err = b.Source("", resultSource)
	require.NoError(t, err)

	// Should return 0.52.0 (latest), not 0.50.0 (yanked) or 0.51.0
	assert.Equal(t, "0.52.0", resultSource.Information)
	assert.NotEqual(t, "0.50.0", resultSource.Information) // Should not be yanked version
}

func TestSource_SemanticVersionSorting(t *testing.T) {
	// Test that semantic versions are sorted correctly (not lexicographically)
	// This tests the bug fix where "0.9.0" vs "0.10.0" would be sorted incorrectly
	// with lexicographic sorting ("0.10.0" < "0.9.0" lexicographically)
	testCases := []struct {
		name           string
		versions       []string
		expectedLatest string
	}{
		{
			name:           "Semantic versions with double digits",
			versions:       []string{"0.9.0", "0.10.0", "0.11.0"},
			expectedLatest: "0.11.0",
		},
		{
			name:           "Semantic versions with single and double digits",
			versions:       []string{"1.0.0", "1.9.0", "1.10.0", "2.0.0"},
			expectedLatest: "2.0.0",
		},
		{
			name:           "Mixed semantic versions",
			versions:       []string{"0.1.0", "0.2.0", "0.9.0", "0.10.0"},
			expectedLatest: "0.10.0",
		},
		{
			name:           "Pre-release versions",
			versions:       []string{"1.0.0", "1.0.1", "1.1.0-rc1", "1.1.0"},
			expectedLatest: "1.1.0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create metadata JSON with the test versions
			versionsJSON := `["` + strings.Join(tc.versions, `", "`) + `"]`
			mockResponse := `{
				"versions": ` + versionsJSON + `,
				"yanked_versions": {}
			}`

			b, err := New(Spec{
				Module: "test_module",
			})
			require.NoError(t, err)

			b.webClient = &httpclient.MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(mockResponse)),
					}, nil
				},
			}

			resultSource := &result.Source{}
			err = b.Source("", resultSource)
			require.NoError(t, err)

			assert.Equal(t, result.SUCCESS, resultSource.Result)
			assert.Equal(t, tc.expectedLatest, resultSource.Information,
				"Expected latest version %q but got %q. Versions were: %v",
				tc.expectedLatest, resultSource.Information, tc.versions)
		})
	}
}
