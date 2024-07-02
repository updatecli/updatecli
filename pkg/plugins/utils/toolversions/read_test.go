package toolversions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadToolVersions(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []Entry
		wantErr bool
	}{
		{
			name:    "single entry",
			content: "nodejs 20.12.0",
			want: []Entry{
				{Key: "nodejs", Value: "20.12.0"},
			},
		},
		{
			name: "multiple entries",
			content: `nodejs 20.12.0
go 1.20`,
			want: []Entry{
				{Key: "nodejs", Value: "20.12.0"},
				{Key: "go", Value: "1.20"},
			},
		},
		{
			name: "ignore comments and empty lines",
			content: `# This is a comment
nodejs 20.12.0

# Another comment
go 1.20`,
			want: []Entry{
				{Key: "nodejs", Value: "20.12.0"},
				{Key: "go", Value: "1.20"},
			},
		},
		{
			name:    "empty content",
			content: "",
			want:    []Entry{},
		},
		{
			name:    "invalid format",
			content: "invalid",
			want:    []Entry{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := readToolVersions(tt.content)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, gotResult)
		})
	}
}
