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
			name: "Passing case",
			spec: Spec{
				File: "/tmp/foo.txt",
			},
			wantErr: false,
		},
		{
			name: "Passing case with 'Line'",
			spec: Spec{
				File: "/tmp/foo.txt",
				Line: 12,
			},
			wantErr: false,
		},
		{
			name: "Passing case with 'Files' containing more than one element",
			spec: Spec{
				Files: []string{
					"/tmp/foo.txt",
					"/tmp/bar.txt",
				},
			},
			wantErr: false,
		},
		{
			name: "Passing case with 'Files' containing one element and a 'Line' is specified",
			spec: Spec{
				Files: []string{
					"/tmp/foo.txt",
				},
				Line: 12,
			},
			wantErr: false,
		},
		{
			name: "Validation failure with 'Files' containing duplicated values",
			spec: Spec{
				Files: []string{
					"/tmp/foo.txt",
					"/tmp/foo.txt",
				},
			},
			wantErr: true,
		},
		{
			name: "Validation failure with both 'File' and 'Files' empty",
			spec: Spec{
				File: "",
			},
			wantErr: true,
		},
		{
			name: "Validation failure with both 'File' and 'Files' not empty",
			spec: Spec{
				File: "/tmp/foo.txt",
				Files: []string{
					"/tmp/bar.txt",
				},
			},
			wantErr: true,
		},
		{
			name: "Validation failure with 'Files' containing more than one element and 'Line' specified",
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
			name: "Validation failure with 'Line' negative",
			spec: Spec{
				File: "/tmp/foo.txt",
				Line: -1,
			},
			wantErr: true,
		},
		{
			name: "Validation failure with both 'Line' and 'ForceCreate=true' specified",
			spec: Spec{
				File:        "/tmp/foo.txt",
				ForceCreate: true,
				Line:        12,
			},
			wantErr: true,
		},
		{
			name: "Validation failure with both 'Line' and 'MatchPattern' specified",
			spec: Spec{
				File:         "/tmp/foo.txt",
				MatchPattern: "pattern=.*",
				Line:         12,
			},
			wantErr: true,
		},
		{
			name: "Validation failure with both 'Line' and 'ReplacePattern' specified",
			spec: Spec{
				File:           "/tmp/foo.txt",
				ReplacePattern: "pattern=.*",
				Line:           13,
			},
			wantErr: true,
		},
		{
			name: "Validation failure with both 'Content' and 'ReplacePattern' specified",
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
			name: "Passing case",
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
			name: "Passing case with 'Line'",
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
			name: "Failing case with non existant 'File'",
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
		{
			name: "Failing case with non existant 'File' and a specified 'Line'",
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
