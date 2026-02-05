package bazelregistry

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

const testMetadataJSON = `{
  "homepage": "https://github.com/bazelbuild/rules_go",
  "maintainers": [
    {
      "name": "Test Maintainer",
      "email": "test@example.com"
    }
  ],
  "repository": [
    "github:bazel-contrib/rules_go"
  ],
  "versions": [
    "0.50.0",
    "0.51.0",
    "0.52.0"
  ],
  "yanked_versions": {
    "0.50.0": "Test yank reason"
  }
}`

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		spec       interface{}
		wantErr    bool
		wantModule string
	}{
		{
			name: "Valid spec with module",
			spec: map[string]interface{}{
				"module": "rules_go",
			},
			wantErr:    false,
			wantModule: "rules_go",
		},
		{
			name: "Valid spec with module and version filter",
			spec: map[string]interface{}{
				"module": "rules_go",
				"versionfilter": map[string]interface{}{
					"kind":    "semver",
					"pattern": "~0.50",
				},
			},
			wantErr:    false,
			wantModule: "rules_go",
		},
		{
			name: "Valid spec with custom URL",
			spec: map[string]interface{}{
				"module": "rules_go",
				"url":    "https://example.com/modules/{module}/metadata.json",
			},
			wantErr:    false,
			wantModule: "rules_go",
		},
		{
			name: "Invalid spec - missing module",
			spec: map[string]interface{}{
				"url": "https://example.com/modules/{module}/metadata.json",
			},
			wantErr: true,
		},
		{
			name: "Invalid spec - empty module",
			spec: map[string]interface{}{
				"module": "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tt.wantModule, got.spec.Module)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		spec    Spec
		wantErr bool
	}{
		{
			name: "Valid spec",
			spec: Spec{
				Module: "rules_go",
			},
			wantErr: false,
		},
		{
			name: "Invalid spec - missing module",
			spec: Spec{
				Module: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestReportConfig(t *testing.T) {
	b := &Bazelregistry{
		spec: Spec{
			Module: "rules_go",
			VersionFilter: version.Filter{
				Kind:    "semver",
				Pattern: "~0.50",
			},
			URL: "https://example.com/modules/{module}/metadata.json",
		},
	}

	config := b.ReportConfig()
	spec, ok := config.(Spec)
	require.True(t, ok)
	assert.Equal(t, "rules_go", spec.Module)
	assert.Equal(t, "semver", spec.VersionFilter.Kind)
	assert.Equal(t, "~0.50", spec.VersionFilter.Pattern)
	// URL should be redacted
	assert.NotEqual(t, "https://example.com/modules/{module}/metadata.json", spec.URL)
}

func TestFetchModuleMetadata(t *testing.T) {
	tests := []struct {
		name           string
		module         string
		mockResponse   string
		mockStatusCode int
		mockError      error
		wantErr        bool
		wantVersions   []string
	}{
		{
			name:           "Success - valid metadata",
			module:         "rules_go",
			mockResponse:   testMetadataJSON,
			mockStatusCode: http.StatusOK,
			wantErr:        false,
			wantVersions:   []string{"0.50.0", "0.51.0", "0.52.0"},
		},
		{
			name:           "Error - module not found",
			module:         "nonexistent",
			mockResponse:   "",
			mockStatusCode: http.StatusNotFound,
			wantErr:        true,
		},
		{
			name:           "Error - server error",
			module:         "rules_go",
			mockResponse:   "Internal Server Error",
			mockStatusCode: http.StatusInternalServerError,
			wantErr:        true,
		},
		{
			name:           "Error - invalid JSON",
			module:         "rules_go",
			mockResponse:   "{ invalid json }",
			mockStatusCode: http.StatusOK,
			wantErr:        true,
		},
		{
			name:           "Error - empty versions",
			module:         "rules_go",
			mockResponse:   `{"versions": []}`,
			mockStatusCode: http.StatusOK,
			wantErr:        true,
		},
		{
			name:           "Error - network error",
			module:         "rules_go",
			mockResponse:   "",
			mockStatusCode: 0,
			mockError:      http.ErrServerClosed,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bazelregistry{
				baseURL: "https://example.com/modules/{module}/metadata.json",
				webClient: &httpclient.MockClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						if tt.mockError != nil {
							return nil, tt.mockError
						}
						return &http.Response{
							StatusCode: tt.mockStatusCode,
							Body:       io.NopCloser(strings.NewReader(tt.mockResponse)),
						}, nil
					},
				},
			}

			metadata, err := b.fetchModuleMetadata(tt.module)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, metadata)
			} else {
				require.NoError(t, err)
				require.NotNil(t, metadata)
				assert.Equal(t, tt.wantVersions, metadata.Versions)
				if len(tt.wantVersions) > 0 {
					assert.Equal(t, "Test yank reason", metadata.YankedVersions["0.50.0"])
				}
			}
		})
	}
}

func TestParseMetadata(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		wantErr   bool
		wantCount int
	}{
		{
			name:      "Valid metadata",
			data:      testMetadataJSON,
			wantErr:   false,
			wantCount: 3,
		},
		{
			name:    "Invalid JSON",
			data:    "{ invalid }",
			wantErr: true,
		},
		{
			name:    "Empty versions",
			data:    `{"versions": []}`,
			wantErr: true,
		},
		{
			name:      "Missing optional fields",
			data:      `{"versions": ["1.0.0"]}`,
			wantErr:   false,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := parseMetadata([]byte(tt.data))
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, metadata)
			} else {
				require.NoError(t, err)
				require.NotNil(t, metadata)
				assert.Equal(t, tt.wantCount, len(metadata.Versions))
			}
		})
	}
}
