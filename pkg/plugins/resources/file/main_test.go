package file

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func Test_Validate(t *testing.T) {
	tests := []struct {
		name          string
		spec          Spec
		mockFileExist bool
		wantErr       bool
	}{
		{
			name: "Normal case",
			spec: Spec{
				File: "/tmp/foo.txt",
				Line: 12,
			},
			mockFileExist: true,
			wantErr:       false,
		},
		{
			name: "raises an error when 'File' is empty",
			spec: Spec{
				File: "",
			},
			mockFileExist: true,
			wantErr:       true,
		},
		{
			name: "raises an error when 'Line' is negative",
			spec: Spec{
				File: "/tmp/foo.txt",
				Line: -1,
			},
			mockFileExist: true,
			wantErr:       true,
		},
		{
			name: "raises an error when both 'Line' and `ForceCreate=true` are specified",
			spec: Spec{
				File:        "/tmp/foo.txt",
				ForceCreate: true,
				Line:        12,
			},
			mockFileExist: true,
			wantErr:       true,
		},
		{
			name: "raises an error when both 'Line' and `MatchPattern` are specified",
			spec: Spec{
				File:         "/tmp/foo.txt",
				MatchPattern: "pattern=.*",
				Line:         12,
			},
			mockFileExist: true,
			wantErr:       true,
		},
		{
			name: "raises an error when both 'Line' and `ReplacePattern` are specified",
			spec: Spec{
				File:           "/tmp/foo.txt",
				ReplacePattern: "pattern=.*",
				Line:           13,
			},
			mockFileExist: true,
			wantErr:       true,
		},
		{
			name: "raises an error when both 'Content' and `ReplacePattern` are specified",
			spec: Spec{
				File:           "/tmp/foo.txt",
				ReplacePattern: "pattern=.*",
				Content:        "Hello World",
			},
			mockFileExist: true,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := File{
				spec: tt.spec,
				contentRetriever: &text.MockTextRetriever{
					Exists: tt.mockFileExist,
				},
			}
			gotErr := file.Validate()
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

func TestFile_Read(t *testing.T) {
	tests := []struct {
		name                string
		wantErr             bool
		mockReturnedContent string
		mockReturnedError   error
		spec                Spec
		wantContent         string
		wantMockState       text.MockTextRetriever
	}{
		{
			name:                "Normal case with a line",
			mockReturnedContent: "Hello World",
			spec: Spec{
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
			spec: Spec{
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
			spec: Spec{
				File: "/not_existing.txt",
				Line: 15,
			},
			wantErr: true,
		},
		{
			name:              "File does not exist without line",
			mockReturnedError: fmt.Errorf("no such file or directory"),
			spec: Spec{
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
