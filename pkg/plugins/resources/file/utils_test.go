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
			name:     "binary content with null byte at start",
			content:  "\x00Hello World",
			expected: true,
		},
		{
			name:     "binary content with null byte in middle",
			content:  "Hello\x00World",
			expected: true,
		},
		{
			name:     "binary content with null byte at end",
			content:  "Hello World\x00",
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

func TestTruncateBinaryContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "binary with null byte",
			content:  "binary\x00data",
			expected: "[binary content, 11 bytes]",
		},
		{
			name:     "empty content",
			content:  "",
			expected: "[binary content, 0 bytes]",
		},
		{
			name:     "large binary content",
			content:  string(make([]byte, 10240)),
			expected: "[binary content, 10240 bytes]",
		},
		{
			name:     "text content",
			content:  "This is text",
			expected: "[binary content, 12 bytes]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateBinaryContent(tt.content)
			if result != tt.expected {
				t.Errorf("truncateBinaryContent() = %v, want %v", result, tt.expected)
			}
		})
	}
}
