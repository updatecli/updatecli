package text

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFile_ReadFromUrl(t *testing.T) {
	tests := []struct {
		name         string
		location     string
		mockRetry    int
		mockBody     string
		mockPath     string
		mockStatus   *int
		expectedBody string
		expectedErr  error
	}{
		{
			name:     "Success",
			location: "https://raw.githubusercontent.com/updatecli/updatecli/v0.81.0/.dockerignore",
			expectedBody: `bin/
*zip
*gz
*.swp
`,
		},
		{
			name:         "Mocked Success",
			mockBody:     "test",
			expectedBody: "test",
		},
		{
			name:        "404",
			location:    "http://updatecli.io/notfound",
			expectedErr: fmt.Errorf("URL \"http://updatecli.io/notfound\" not found or in error"),
		},
		{
			name:         "Success with retry",
			mockBody:     "test",
			mockRetry:    1,
			expectedBody: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location := tt.location
			if tt.mockBody != "" {
				currentRetry := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if currentRetry < tt.mockRetry {
						currentRetry += 1
						w.WriteHeader(http.StatusGatewayTimeout)
						return
					}
					if tt.mockPath != "" && r.URL.Path != tt.mockPath {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					status := http.StatusOK
					if tt.mockStatus != nil {
						status = *tt.mockStatus
					}
					w.WriteHeader(status)
					_, _ = w.Write([]byte(tt.mockBody))
				}))
				defer server.Close()
				location = fmt.Sprintf("%s/%s", server.URL, tt.location)
			}

			tr := Text{}
			content, err := tr.readFromURL(location, 0)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedBody, content)
		})
	}
}
