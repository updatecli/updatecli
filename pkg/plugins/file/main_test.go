package file

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShell_New(t *testing.T) {
	tests := []struct {
		name     string
		spec     FileSpec
		wantErr  bool
		wantFile *File
	}{
		{
			name: "Normal case",
			spec: FileSpec{
				File:    "a_file.txt",
				Content: "Hello World\nInFile",
			},
			wantErr: false,
			wantFile: &File{
				spec: FileSpec{
					File:    "a_file.txt",
					Content: "Hello World\nInFile",
				},
			},
		},
		{
			name:    "Fail if no file path specified",
			spec:    FileSpec{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFile, gotErr := New(tt.spec)

			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)

			assert.Equal(t, tt.wantFile, gotFile)
		})
	}
}
