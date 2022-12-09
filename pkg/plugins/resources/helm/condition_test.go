package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
