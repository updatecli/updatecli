package redact

import "testing"

func TestURL(t *testing.T) {
	testdata := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "invalid URL",
			input:    "https://user:",
			expected: "https://user:",
		},
		{
			name:     "redact user and password from a valid URL",
			input:    "https://user:password@example.com",
			expected: "https://****:****@example.com",
		},
		{
			name:     "nothing to redact from a valid URL",
			input:    "https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "nothing to redact from a valid URL",
			input:    "example.com",
			expected: "example.com",
		},
		{
			name:     "nothing to redact from a valid URL",
			input:    "tfr://example.com",
			expected: "tfr://example.com",
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			got := URL(tt.input)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
