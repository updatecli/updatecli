package pypi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizePEP440(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "2.31.0", expected: "2.31.0"},
		{input: "0.51b0", expected: "0.51.0-beta.0"},
		{input: "1.0a1", expected: "1.0.0-alpha.1"},
		{input: "2.0rc1", expected: "2.0.0-rc.1"},
		{input: "1.0.post1", expected: "1.0.0"},
		{input: "1.0.dev3", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, normalizePEP440(tt.input))
		})
	}
}

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
