package helm

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestSource(t *testing.T) {

	tests := []struct {
		name                 string
		chart                Spec
		expected             string
		expectedError        bool
		expectedErrorMessage error
	}{
		{
			name: "Successful result",
			chart: Spec{
				URL:  "https://stenic.github.io/helm-charts",
				Name: "proxy",
			},
			expected: "1.0.3",
		},
		{
			name: "Chart not found",
			chart: Spec{
				URL:  "https://charts.jetstack.io",
				Name: "tor-prox",
			},
			expectedError:        true,
			expectedErrorMessage: errors.New("helm chart \"tor-prox\" not found from Helm Chart repository \"https://example.com/index.yaml\""),
		},
		{
			name: "Registry not found",
			chart: Spec{
				URL:     "https://example.com",
				Name:    "jenkins",
				Version: "999",
			},
			expected:             "",
			expectedError:        true,
			expectedErrorMessage: errors.New("something went wrong while contacting \"https://example.com/index.yaml\""),
		},
		{
			name: "Successful OCI result",
			chart: Spec{
				URL:  "oci://ghcr.io/olblak/charts/",
				Name: "upgrade-responder",
				// Following credentials are needed by Github Action workflow to run the tests
				// If GITHUB_ACTOR and GITHUB_TOKEN are not set then we fallback to
				// the default docker credential file
				InlineKeyChain: docker.InlineKeyChain{
					Username: os.Getenv("GITHUB_ACTOR"),
					Token:    os.Getenv("GITHUB_TOKEN"),
				},
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "v0.1.5",
				},
			},
			expected: "v0.1.5",
		},
		{
			name: "Successful OCI result using semver version filter",
			chart: Spec{
				URL:  "oci://ghcr.io/olblak/charts",
				Name: "upgrade-responder",
				// Following credentials are needed by Github Action workflow to run the tests
				// If GITHUB_ACTOR and GITHUB_TOKEN are not set then we fallback to
				// the default docker credential file
				InlineKeyChain: docker.InlineKeyChain{
					Username: os.Getenv("GITHUB_ACTOR"),
					Token:    os.Getenv("GITHUB_TOKEN"),
				},
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "v0.1.5",
				},
			},
			expected: "v0.1.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.chart)
			require.NoError(t, err)

			gotVersion, err := got.Source("")

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
