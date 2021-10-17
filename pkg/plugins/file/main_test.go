package file

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func TestFile_New(t *testing.T) {
	tests := []struct {
		name     string
		spec     FileSpec
		wantErr  bool
		wantFile *File
	}{
		{
			name: "Normal case",
			spec: FileSpec{
				File: "/tmp/foo.txt",
				Line: 12,
			},
			wantErr: false,
			wantFile: &File{
				contentRetriever: &text.Text{},
				spec: FileSpec{
					File: "/tmp/foo.txt",
					Line: 12,
				},
			},
		},
		{
			name: "Normal case with a 'file://' prefix",
			spec: FileSpec{
				File: "file:///tmp/bar.txt",
			},
			wantErr: false,
			wantFile: &File{
				contentRetriever: &text.Text{},
				spec: FileSpec{
					File: "/tmp/bar.txt",
				},
			},
		},
		{
			name: "raises an error when 'File' is empty",
			spec: FileSpec{
				File: "",
			},
			wantErr: true,
		},
		{
			name: "raises an error when 'Line' is negative",
			spec: FileSpec{
				File: "/tmp/foo.txt",
				Line: -1,
			},
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

func TestFile_Read(t *testing.T) {
	tests := []struct {
		name                string
		wantErr             bool
		mockReturnedContent string
		mockReturnedError   error
		spec                FileSpec
		wantContent         string
		wantMockState       text.MockTextRetriever
	}{
		{
			name:                "Normal case with a line",
			mockReturnedContent: "Hello World",
			spec: FileSpec{
				Line: 3,
				File: "/foo.txt",
			},
			wantContent: "Hello World",
			wantMockState: text.MockTextRetriever{
				Line:     3,
				Location: "/foo.txt",
			},
		},
		{
			name:                "Normal case without a line",
			mockReturnedContent: "Hello World",
			spec: FileSpec{
				File: "/bar.txt",
			},
			wantContent: "Hello World",
			wantMockState: text.MockTextRetriever{
				Location: "/bar.txt",
			},
		},
		{
			name:              "File does not exist with a line",
			mockReturnedError: fmt.Errorf("no such file or directory"),
			spec: FileSpec{
				File: "/not_existing.txt",
				Line: 15,
			},
			wantErr: true,
		},
		{
			name:              "File does not exist without line",
			mockReturnedError: fmt.Errorf("no such file or directory"),
			spec: FileSpec{
				File: "/not_existing.txt",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockText := text.MockTextRetriever{
				Content: tt.mockReturnedContent,
				Err:     tt.mockReturnedError,
			}
			f := &File{
				spec:             tt.spec,
				contentRetriever: &mockText,
			}
			gotErr := f.Read()

			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)

			assert.Equal(t, tt.wantContent, f.CurrentContent)
			assert.Equal(t, tt.wantMockState.Line, mockText.Line)
			assert.Equal(t, tt.wantMockState.Location, mockText.Location)
		})
	}
}
