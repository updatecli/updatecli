package toolversions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPut(t *testing.T) {
	tests := []struct {
		name     string
		initial  []Entry
		key      string
		value    string
		expected []Entry
	}{
		{
			name: "Add new entry",
			initial: []Entry{
				{Key: "node", Value: "20.12.0"},
			},
			key:   "go",
			value: "1.20",
			expected: []Entry{
				{Key: "node", Value: "20.12.0"},
				{Key: "go", Value: "1.20"},
			},
		},
		{
			name: "Update existing entry",
			initial: []Entry{
				{Key: "node", Value: "20.12.0"},
				{Key: "go", Value: "1.20"},
			},
			key:   "go",
			value: "1.21",
			expected: []Entry{
				{Key: "node", Value: "20.12.0"},
				{Key: "go", Value: "1.21"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FileContent{
				Entries: tt.initial,
			}
			err := f.Put(tt.key, tt.value)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, f.Entries)
		})
	}
}
