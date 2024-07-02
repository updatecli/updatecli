package toolversions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	tests := []struct {
		name          string
		fileContent   FileContent
		key           string
		expectedValue string
		expectError   bool
	}{
		{
			name: "Key exists",
			fileContent: FileContent{
				Entries: []Entry{
					{Key: "node", Value: "20.12.0"},
				},
			},
			key:           "node",
			expectedValue: "20.12.0",
			expectError:   false,
		},
		{
			name: "Key does not exist",
			fileContent: FileContent{
				Entries: []Entry{
					{Key: "node", Value: "20.12.0"},
				},
			},
			key:         "python",
			expectError: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.fileContent.Get(tt.key)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, value)
			}
		})
	}
}
