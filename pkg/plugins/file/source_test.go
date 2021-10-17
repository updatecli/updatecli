package file

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func TestFile_Source(t *testing.T) {
	tests := []struct {
		name                string
		file                FileSpec
		wantSource          string
		wantErr             bool
		mockReturnedContent string
		mockReturnedError   error
		wantMockState       text.MockTextRetriever
	}{
		{
			name: "Normal Case",
			file: FileSpec{
				File: "/home/ucli/foo.txt",
			},
			mockReturnedContent: "current_version=1.2.3",
			wantSource:          "current_version=1.2.3",
			wantMockState: text.MockTextRetriever{
				Location: "/home/ucli/foo.txt",
			},
		},
		{
			name:                "File does not exists",
			mockReturnedError:   fmt.Errorf("no such file or directory"),
			mockReturnedContent: "",
			wantErr:             true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockText := text.MockTextRetriever{
				Content: tt.mockReturnedContent,
				Err:     tt.mockReturnedError,
			}
			f := &File{
				spec:             tt.file,
				contentRetriever: &mockText,
			}
			source, gotErr := f.Source(tt.file.File)
			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantSource, source)
			assert.Equal(t, tt.wantMockState.Location, mockText.Location)
			assert.Equal(t, tt.wantMockState.Line, mockText.Line)
		})
	}
}
