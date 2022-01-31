package file

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/text"
)

// TODO: find a way to test when there are multiple files but only some are changed
func TestFile_Target(t *testing.T) {
	tests := []struct {
		spec                Spec
		name                string
		inputSourceValue    string
		mockReturnedContent string
		mockReturnedError   error
		mockFileExists      bool
		wantMockState       text.MockTextRetriever
		wantResult          bool
		wantErr             bool
		dryRun              bool
	}{
		{
			name:             "Replace content with matchPattern and ReplacePattern",
			inputSourceValue: "3.9.0",
			spec: Spec{
				File:           "foo.txt",
				MatchPattern:   "maven_(.*)=.*",
				ReplacePattern: "maven_$1= 3.9.0",
			},
			mockFileExists: true,
			mockReturnedContent: `maven_version = "3.8.2"
		git_version = "2.33.1"
		jdk11_version = "11.0.12+7"
		jdk17_version = "17+35"
		jdk8_version = "8u302-b08"
		maven_major_release = "3"
		git_lfs_version = "3.0.1"
		compose_version = "1.29.2"`,
			wantResult: true,
			wantMockState: text.MockTextRetriever{
				Location: "foo.txt",
				Content: `maven_version = 3.9.0
		git_version = "2.33.1"
		jdk11_version = "11.0.12+7"
		jdk17_version = "17+35"
		jdk8_version = "8u302-b08"
		maven_major_release = 3.9.0
		git_lfs_version = "3.0.1"
		compose_version = "1.29.2"`,
			},
		},
		{
			name: "Passing case with both input source and specified content but no line (specified content should be used)",
			spec: Spec{
				File:    "foo.txt",
				Content: "Hello World",
			},
			mockFileExists:   true,
			inputSourceValue: "current_version=1.2.3",
			wantResult:       true,
			wantMockState: text.MockTextRetriever{
				Location: "foo.txt",
				Content:  "Hello World",
			},
		},
		{
			name: "Passing case with an updated line from provided content",
			spec: Spec{
				File:    "foo.txt",
				Content: "Hello World",
				Line:    2,
			},
			mockFileExists:      true,
			mockReturnedContent: "Title\nGood Bye\nThe end",
			inputSourceValue:    "current_version=1.2.3",
			wantResult:          true,
			wantMockState: text.MockTextRetriever{
				Location: "foo.txt",
				Line:     2,
				Content:  "Hello World",
			},
		},
		{
			name: "Validation failure with an https:// URL instead of a file",
			spec: Spec{
				File: "https://github.com/foo.txt",
			},
			wantResult: false,
			wantErr:    true,
		},
		{
			name: "Validation failure with both line and forcecreate specified",
			spec: Spec{
				File:        "foo.txt",
				ForceCreate: true,
				Line:        2,
			},
			wantResult: false,
			wantErr:    true,
		},
		{
			name: "Validation Failure with invalid regexp for MatchPattern",
			spec: Spec{
				MatchPattern: "(d+:1",
				File:         "/bar.txt",
			},
			mockFileExists:      true,
			mockReturnedContent: `maven_version = "3.8.2"`,
			wantResult:          false,
			wantErr:             true,
		},
		{
			name: "Error with file not existing (with line)",
			spec: Spec{
				File: "not_existing.txt",
				Line: 3,
			},
			wantResult: false,
			wantErr:    true,
		},
		{
			name: "Error with file not existing (with content)",
			spec: Spec{
				File:    "not_existing.txt",
				Content: "Hello World",
			},
			wantResult: false,
			wantErr:    true,
		},
		{
			name: "Error while reading the line in file",
			spec: Spec{
				File: "not_existing.txt",
				Line: 3,
			},
			mockFileExists:    true,
			mockReturnedError: fmt.Errorf("I/O error: file system too slow"),
			wantResult:        false,
			wantErr:           true,
		},
		{
			name: "Error while reading a full file",
			spec: Spec{
				File:    "not_existing.txt",
				Content: "Hello World",
			},
			mockFileExists:    true,
			mockReturnedError: fmt.Errorf("I/O error: file system too slow"),
			wantResult:        false,
			wantErr:           true,
		},
		{
			name: "Line in file not updated",
			spec: Spec{
				File: "foo.txt",
				Line: 3,
			},
			mockFileExists:      true,
			mockReturnedContent: "current_version=1.2.3",
			inputSourceValue:    "current_version=1.2.3",
			wantResult:          false,
			wantMockState: text.MockTextRetriever{
				Location: "foo.txt",
				Line:     3,
				Content:  "current_version=1.2.3",
			},
		},
		{
			name: "File not updated (input source, no specified line)",
			spec: Spec{
				File: "foo.txt",
			},
			mockFileExists:      true,
			mockReturnedContent: "current_version=1.2.3",
			inputSourceValue:    "current_version=1.2.3",
			wantResult:          false,
			wantMockState: text.MockTextRetriever{
				Location: "foo.txt",
				Content:  "current_version=1.2.3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockText := text.MockTextRetriever{
				Content: tt.mockReturnedContent,
				Err:     tt.mockReturnedError,
				Exists:  tt.mockFileExists,
			}
			f := &File{
				spec:             tt.spec,
				contentRetriever: &mockText,
			}

			f.files = make(map[string]string)
			// File as unique element of f.files
			if len(f.spec.File) > 0 {
				f.files[strings.TrimPrefix(f.spec.File, "file://")] = ""
			}
			// Files
			for _, file := range f.spec.Files {
				f.files[strings.TrimPrefix(file, "file://")] = ""
			}

			gotResult, gotErr := f.Target(tt.inputSourceValue, tt.dryRun)
			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantResult, gotResult)
			assert.Equal(t, tt.wantMockState.Location, mockText.Location)
			assert.Equal(t, tt.wantMockState.Line, mockText.Line)
			assert.Equal(t, tt.wantMockState.Content, mockText.Content)
		})
	}
}

func TestFile_TargetFromSCM(t *testing.T) {
	tests := []struct {
		spec                Spec
		wantFiles           []string
		name                string
		inputSourceValue    string
		mockReturnedContent string
		mockReturnedError   error
		scm                 scm.ScmHandler
		wantResult          bool
		wantErr             bool
		mockFileExists      bool
		dryRun              bool
		wantMockState       text.MockTextRetriever
	}{
		{
			name: "Passing case with relative path",
			spec: Spec{
				File: "foo.txt",
				Line: 3,
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			mockFileExists:   true,
			inputSourceValue: "current_version=1.2.3",
			wantResult:       true,
			wantFiles:        []string{"/tmp/foo.txt"},
			wantMockState: text.MockTextRetriever{
				Location: "/tmp/foo.txt",
				Content:  "current_version=1.2.3",
				Line:     3,
			},
		},
		{
			name: "Passing case with file created",
			spec: Spec{
				File:        "foo.txt",
				ForceCreate: true,
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			mockFileExists:   false,
			inputSourceValue: "current_version=1.2.3",
			wantResult:       true,
			wantFiles:        []string{"/tmp/foo.txt"},
			wantMockState: text.MockTextRetriever{
				Location: "/tmp/foo.txt",
				Content:  "current_version=1.2.3",
			},
		},
		{
			name:             "No line matched with matchPattern and ReplacePattern defined",
			inputSourceValue: "3.9.0",
			spec: Spec{
				File:           "foo.txt",
				MatchPattern:   "notmatching=*",
				ReplacePattern: "maven_version= 3.9.0",
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			mockFileExists: true,
			mockReturnedContent: `maven_version = "3.8.2"
		git_version = "2.33.1"
		jdk11_version = "11.0.12+7"
		jdk17_version = "17+35"
		jdk8_version = "8u302-b08"
		maven_major_release = "3"
		git_lfs_version = "3.0.1"
		compose_version = "1.29.2"`,
			wantResult: false,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockText := text.MockTextRetriever{
				Err:    tt.mockReturnedError,
				Exists: tt.mockFileExists,
			}
			f := &File{
				spec:             tt.spec,
				contentRetriever: &mockText,
			}

			f.files = make(map[string]string)
			// File as unique element of f.files
			if len(f.spec.File) > 0 {
				f.files[strings.TrimPrefix(f.spec.File, "file://")] = ""
			}
			// Files
			for _, file := range f.spec.Files {
				f.files[strings.TrimPrefix(file, "file://")] = ""
			}

			gotResult, gotFiles, _, gotErr := f.TargetFromSCM(tt.inputSourceValue, tt.scm, tt.dryRun)

			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)

			assert.Equal(t, tt.wantResult, gotResult)
			assert.Equal(t, tt.wantFiles, gotFiles)
			assert.Equal(t, tt.wantMockState.Location, mockText.Location)
			assert.Equal(t, tt.wantMockState.Line, mockText.Line)
			assert.Equal(t, tt.wantMockState.Content, mockText.Content)
		})
	}
}
