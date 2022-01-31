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
		name    string
		spec    Spec
		files   map[string]string
		wantErr bool
	}{
		{
			name: "Normal case with 'File'",
			spec: Spec{
				File: "/tmp/foo.txt",
				Line: 12,
			},
			wantErr: false,
		},
		{
			name: "Normal case with 'Files' containing one element and a 'Line' is specified",
			spec: Spec{
				Files: []string{
					"/tmp/foo.txt",
				},
				Line: 12,
			},
			wantErr: false,
		},
		{
			name: "Normal case with 'Files' containing more than one element and no 'Line' is specified",
			spec: Spec{
				Files: []string{
					"/tmp/foo.txt",
					"/tmp/bar.txt",
				},
			},
			wantErr: false,
		},
		{
			name: "raises and error when 'Files' containing duplicated values",
			spec: Spec{
				Files: []string{
					"/tmp/foo.txt",
					"/tmp/foo.txt",
				},
			},
			wantErr: true,
		},
		{
			name: "raises an error when 'File' and 'Files' are empty",
			spec: Spec{
				File: "",
			},
			wantErr: true,
		},
		{
			name: "raises an error when 'File' and 'Files' are not empty",
			spec: Spec{
				File: "/tmp/foo.txt",
				Files: []string{
					"/tmp/bar.txt",
				},
			},
			wantErr: true,
		},
		{
			name: "raises an error when 'Files' contains more than one element and 'Line' is specified",
			spec: Spec{
				Files: []string{
					"/tmp/foo.txt",
					"/tmp/bar.txt",
				},
				Line: 12,
			},
			wantErr: true,
		},
		{
			name: "raises an error when 'Line' is negative",
			spec: Spec{
				File: "/tmp/foo.txt",
				Line: -1,
			},
			wantErr: true,
		},
		{
			name: "raises an error when both 'Line' and `ForceCreate=true` are specified",
			spec: Spec{
				File:        "/tmp/foo.txt",
				ForceCreate: true,
				Line:        12,
			},
			wantErr: true,
		},
		{
			name: "raises an error when both 'Line' and `MatchPattern` are specified",
			spec: Spec{
				File:         "/tmp/foo.txt",
				MatchPattern: "pattern=.*",
				Line:         12,
			},
			wantErr: true,
		},
		{
			name: "raises an error when both 'Line' and `ReplacePattern` are specified",
			spec: Spec{
				File:           "/tmp/foo.txt",
				ReplacePattern: "pattern=.*",
				Line:           13,
			},
			wantErr: true,
		},
		{
			name: "raises an error when both 'Content' and `ReplacePattern` are specified",
			spec: Spec{
				File:           "/tmp/foo.txt",
				ReplacePattern: "pattern=.*",
				Content:        "Hello World",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := tt.spec
			gotErr := spec.Validate()
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
		spec                Spec
		files               map[string]string
		mockReturnedContent string
		mockReturnedError   error
		mockFileExist       bool
		wantContent         string
		wantMockState       text.MockTextRetriever
		wantErr             bool
	}{
		{
			name: "Normal case with a line",
			spec: Spec{
				Line: 3,
				File: "/foo.txt",
			},
			files: map[string]string{
				"/foo.txt": "",
			},
			mockReturnedContent: "Hello World",
			mockFileExist:       true,
			wantContent:         "Hello World",
			wantMockState: text.MockTextRetriever{
				Line:     3,
				Location: "/foo.txt",
			},
		},
		{
			name: "Normal case without a line",
			spec: Spec{
				File: "/bar.txt",
			},
			files: map[string]string{
				"/bar.txt": "",
			},
			mockReturnedContent: "Hello World",
			mockFileExist:       true,
			wantContent:         "Hello World",
			wantMockState: text.MockTextRetriever{
				Location: "/bar.txt",
			},
		},
		{
			name: "File does not exist with a line",
			spec: Spec{
				File: "/not_existing.txt",
				Line: 15,
			},
			files: map[string]string{
				"/not_existing.txt": "",
			},
			mockReturnedError: fmt.Errorf("no such file or directory"),
			mockFileExist:     false,
			wantErr:           true,
		},
		{
			name: "File does not exist without line",
			spec: Spec{
				File: "/not_existing.txt",
			},
			files: map[string]string{
				"/not_existing.txt": "",
			},
			mockReturnedError: fmt.Errorf("no such file or directory"),
			mockFileExist:     false,
			wantErr:           true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockText := text.MockTextRetriever{
				Content: tt.mockReturnedContent,
				Err:     tt.mockReturnedError,
				Exists:  tt.mockFileExist,
			}
			f := &File{
				spec:             tt.spec,
				contentRetriever: &mockText,
				files:            tt.files,
			}
			gotErr := f.Read()
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)

			assert.Equal(t, tt.wantContent, f.files[f.spec.File])
			assert.Equal(t, tt.wantMockState.Line, mockText.Line)
			assert.Equal(t, tt.wantMockState.Location, mockText.Location)
		})
	}
}
