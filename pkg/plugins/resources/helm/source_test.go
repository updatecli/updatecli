package helm

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
