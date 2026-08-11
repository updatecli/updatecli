package file

import "testing"

func TestIsBinaryContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "text content",
			content:  "Hello, World!",
			expected: false,
		},
		{
			name:     "binary content with null bytes",
			content:  "\x00\x01\x02\x03binary content",
			expected: true,
		},
		{
			name:     "empty content",
			content:  "",
			expected: false,
		},
		{
			name:     "JSON content",
			content:  `{"key": "value", "number": 123}`,
			expected: false,
		},
		{
			name:     "YAML content",
			content:  "name: test\nversion: 1.0.0\n",
			expected: false,
		},
		{
			name:     "multi-line text",
			content:  "Line 1\nLine 2\nLine 3\n",
			expected: false,
		},
		{
			name:     "text with special characters",
			content:  "Hello! @#$%^&*() 世界",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBinaryContent(tt.content)
			if result != tt.expected {
				t.Errorf("isBinaryContent() = %v, want %v", result, tt.expected)
			}
		})
	}
}
