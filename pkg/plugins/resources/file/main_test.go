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
		name           string
		spec           Spec
		files          map[string]fileMetadata
		mockedContents map[string]string
		mockedError    error
		wantedContents map[string]string
		wantedResult   bool
		wantedErr      bool
	}{
		{
			name: "Passing case",
			spec: Spec{
				File: "/bar.txt",
			},
			files: map[string]fileMetadata{
				"/bar.txt": {
					originalPath: "/bar.txt",
					path:         "/bar.txt",
				},
			},
			mockedContents: map[string]string{
				"/bar.txt": "Hello World",
			},
			wantedContents: map[string]string{
				"/bar.txt": "Hello World",
			},
			wantedResult: true,
		},
		{
			name: "Passing case with 'Line'",
			spec: Spec{
				Line: 3,
				File: "/foo.txt",
			},
			files: map[string]fileMetadata{
				"/foo.txt": {
					originalPath: "/foo.txt",
					path:         "/foo.txt",
				},
			},
			mockedContents: map[string]string{
				"/foo.txt": "Title\r\nGood Bye\r\nThe End",
			},
			wantedContents: map[string]string{
				"/foo.txt": "Title\r\nGood Bye\r\nThe End",
			},
			wantedResult: true,
		},
		{
			name: "Failing case with nonexistent 'Line'",
			spec: Spec{
				Line: 5,
				File: "/foo.txt",
			},
			files: map[string]fileMetadata{
				"/foo.txt": {
					originalPath: "/foo.txt",
					path:         "/foo.txt",
				},
			},
			mockedContents: map[string]string{
				"/foo.txt": "Title\r\nGood Bye\r\nThe End",
			},
			wantedContents: map[string]string{
				"/foo.txt": "Title\r\nGood Bye\r\nThe End",
			},
			wantedErr: true,
		},
		{
			name: "Failing case with nonexistent 'File'",
			spec: Spec{
				File: "/not_existing.txt",
			},
			files: map[string]fileMetadata{
				"/not_existing.txt": {
					originalPath: "/not_existing.txt",
					path:         "/not_existing.txt",
				},
			},
			mockedError: fmt.Errorf("no such file or directory"),
			wantedErr:   true,
		},
		{
			name: "Failing case with nonexistent 'File' and a specified 'Line'",
			spec: Spec{
				File: "/not_existing.txt",
				Line: 15,
			},
			files: map[string]fileMetadata{
				"/not_existing.txt": {
					originalPath: "/not_existing.txt",
					path:         "/not_existing.txt",
				},
			},
			mockedError: fmt.Errorf("no such file or directory"),
			wantedErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockText := text.MockTextRetriever{
				Contents: tt.mockedContents,
				Err:      tt.mockedError,
			}
			f := &File{
				spec:             tt.spec,
				contentRetriever: &mockText,
				files:            tt.files,
			}

			gotErr := f.Read()

			if tt.wantedErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)

			for filePath := range tt.files {
				assert.Equal(t, tt.wantedContents[filePath], mockText.Contents[filePath])
			}
		})
	}
}
