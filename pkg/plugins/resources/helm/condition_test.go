package helm

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
)

func TestCondition(t *testing.T) {
	tests := []struct {
		name                 string
		chart                Spec
		expected             bool
		expectedError        bool
		expectedErrorMessage error
	}{
		{
			chart: Spec{
				URL:     "https://charts.jenkins.io",
				Name:    "jenkins",
				Version: "2.19.0",
			},
			expected: true,
		},
		{
			chart: Spec{
				URL:     "https://kubernetes-charts.storage.googleapis.com",
				Name:    "jenkins",
				Version: "999",
			},
			expected: false,
		},
		{
			chart: Spec{
				URL:     "https://example.com",
				Name:    "jenkins",
				Version: "999",
			},
			expected: false,
		},
		{
			name: "Successful OCI result v1",
			chart: Spec{
				URL:     "oci://ghcr.io/olblak/charts/",
				Name:    "upgrade-responder",
				Version: "v0.1.5",
				// Following credentials are needed by Github Action workflow to run the tests
				// If GITHUB_ACTOR and GITHUB_TOKEN are not set then we fallback to
				// the default docker credential file
				InlineKeyChain: docker.InlineKeyChain{
					Username: os.Getenv("GITHUB_ACTOR"),
					Token:    os.Getenv("GITHUB_TOKEN"),
				},
			},
			expected: true,
		},
		{
			name: "Not found OCI result",
			chart: Spec{
				URL:     "oci://ghcr.io/olblak/charts/",
				Name:    "upgrade-responder",
				Version: "v9.9.9",
				// Following credentials are needed by Github Action workflow to run the tests
				// If GITHUB_ACTOR and GITHUB_TOKEN are not set then we fallback to
				// the default docker credential file
				InlineKeyChain: docker.InlineKeyChain{
					Username: os.Getenv("GITHUB_ACTOR"),
					Token:    os.Getenv("GITHUB_TOKEN"),
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.chart)
			require.NoError(t, err)

			gotVersion, err := got.Condition("")

			switch tt.expectedError {
			case true:
				assert.Error(t, err)
			case false:
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expected, gotVersion)
		})
	}
}
