package pypi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		spec          Spec
		expectedError bool
		expectedURL   string
	}{
		{
			name: "Empty name returns validation error",
			spec: Spec{
				Name: "",
			},
			expectedError: true,
		},
		{
			name: "Valid spec returns no error",
			spec: Spec{
				Name: "requests",
			},
			expectedError: false,
			expectedURL:   pypiDefaultURL,
		},
		{
			name: "URL without trailing slash gets one added",
			spec: Spec{
				Name: "requests",
				URL:  "https://private.registry.example.com",
			},
			expectedError: false,
			expectedURL:   "https://private.registry.example.com/",
		},
		{
			name: "URL with trailing slash is unchanged",
			spec: Spec{
				Name: "requests",
				URL:  "https://private.registry.example.com/",
			},
			expectedError: false,
			expectedURL:   "https://private.registry.example.com/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedURL, got.spec.URL)
		})
	}
}
