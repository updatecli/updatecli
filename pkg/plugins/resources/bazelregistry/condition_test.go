package bazelregistry

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
)

func TestCondition(t *testing.T) {
	tests := []struct {
		name           string
		spec           Spec
		source         string
		mockResponse   string
		mockStatusCode int
		mockError      error
		wantErr        bool
		wantPass       bool
		wantMessage    string
	}{
		{
			name: "Success - version exists",
			spec: Spec{
				Module: "rules_go",
			},
			source:         "0.51.0",
			mockResponse:   testMetadataJSON,
			mockStatusCode: http.StatusOK,
			wantErr:        false,
			wantPass:       true,
			wantMessage:    "version \"0.51.0\" is available",
		},
		{
			name: "Failure - version does not exist",
			spec: Spec{
				Module: "rules_go",
			},
			source:         "99.99.99",
			mockResponse:   testMetadataJSON,
			mockStatusCode: http.StatusOK,
			wantErr:        false,
			wantPass:       false,
			wantMessage:    "version \"99.99.99\" not found",
		},
		{
			name: "Failure - version is yanked",
			spec: Spec{
				Module: "rules_go",
			},
			source:         "0.50.0",
			mockResponse:   testMetadataJSON,
			mockStatusCode: http.StatusOK,
			wantErr:        false,
			wantPass:       false,
			wantMessage:    "version \"0.50.0\" exists but is yanked",
		},
		{
			name: "Error - empty source version",
			spec: Spec{
				Module: "rules_go",
			},
			source:         "",
			mockResponse:   testMetadataJSON,
			mockStatusCode: http.StatusOK,
			wantErr:        true,
		},
		{
			name: "Error - module not found",
			spec: Spec{
				Module: "nonexistent",
			},
			source:         "1.0.0",
			mockResponse:   "",
			mockStatusCode: http.StatusNotFound,
			wantErr:        true,
		},
		{
			name: "Error - network error",
			spec: Spec{
				Module: "rules_go",
			},
			source:         "1.0.0",
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

			pass, message, err := b.Condition(tt.source, nil)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantPass, pass)
				assert.Contains(t, message, tt.wantMessage)
			}
		})
	}
}
