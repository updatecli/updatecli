package file

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func TestFile_Target(t *testing.T) {
	tests := []struct {
		spec                  FileSpec
		wantMockState         text.MockTextRetriever
		name                  string
		inputSourceValue      string
		mockReturnedContent   string
		mockReturnedError     error
		mockReturnsFileExists bool
		wantResult            bool
		wantErr               bool
		dryRun                bool
	}{
		{
			name:             "Replace content with matchPattern and ReplacePattern",
			inputSourceValue: "3.9.0",
			spec: FileSpec{
				File:           "foo.txt",
				MatchPattern:   "maven_(.*)=.*",
				ReplacePattern: "maven_$1= 3.9.0",
			},
			mockReturnsFileExists: true,
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
			name: "Passing Case with both input source and specified content but no line (specified content should be used)",
			spec: FileSpec{
				File:    "foo.txt",
				Content: "Hello World",
			},
			mockReturnsFileExists: true,
			inputSourceValue:      "current_version=1.2.3",
			wantResult:            true,
			wantMockState: text.MockTextRetriever{
				Location: "foo.txt",
				Content:  "Hello World",
			},
		},
		{
			name: "Passing Case with an updated line from provided content",
			spec: FileSpec{
				File:    "foo.txt",
				Content: "Hello World",
				Line:    2,
			},
			mockReturnsFileExists: true,
			mockReturnedContent:   "Title\nGood Bye\nThe end",
			inputSourceValue:      "current_version=1.2.3",
			wantResult:            true,
			wantMockState: text.MockTextRetriever{
				Location: "foo.txt",
				Line:     2,
				Content:  "Hello World",
			},
		},
		{
			name: "Validation failure with an https:// URL instead of a file",
			spec: FileSpec{
				File: "https://github.com/foo.txt",
			},
			wantResult: false,
			wantErr:    true,
		},
		{
			name: "Validation failure with both line and forcecreate specified",
			spec: FileSpec{
				File:        "foo.txt",
				ForceCreate: true,
				Line:        2,
			},
			wantResult: false,
			wantErr:    true,
		},
		{
			name: "Validation Failure with invalid regexp for MatchPattern",
			spec: FileSpec{
				MatchPattern: "(d+:1",
				File:         "/bar.txt",
			},
			mockReturnsFileExists: true,
			mockReturnedContent:   `maven_version = "3.8.2"`,
			wantResult:            false,
			wantErr:               true,
		},
		{
			name: "Error with file not existing (with line)",
			spec: FileSpec{
				File: "not_existing.txt",
				Line: 3,
			},
			wantResult: false,
			wantErr:    true,
		},
		{
			name: "Error with file not existing (with content)",
			spec: FileSpec{
				File:    "not_existing.txt",
				Content: "Hello World",
			},
			wantResult: false,
			wantErr:    true,
		},
		{
			name: "Error while reading the line in file",
			spec: FileSpec{
				File: "not_existing.txt",
				Line: 3,
			},
			mockReturnsFileExists: true,
			mockReturnedError:     fmt.Errorf("I/O error: file system too slow"),
			wantResult:            false,
			wantErr:               true,
		},
		{
			name: "Error while reading a full file",
			spec: FileSpec{
				File:    "not_existing.txt",
				Content: "Hello World",
			},
			mockReturnsFileExists: true,
			mockReturnedError:     fmt.Errorf("I/O error: file system too slow"),
			wantResult:            false,
			wantErr:               true,
		},
		{
			name: "Line in file not updated",
			spec: FileSpec{
				File: "foo.txt",
				Line: 3,
			},
			mockReturnsFileExists: true,
			mockReturnedContent:   "current_version=1.2.3",
			inputSourceValue:      "current_version=1.2.3",
			wantResult:            false,
			wantMockState: text.MockTextRetriever{
				Location: "foo.txt",
				Line:     3,
				Content:  "current_version=1.2.3",
			},
		},
		{
			name: "File not updated (input source, no specified line)",
			spec: FileSpec{
				File: "foo.txt",
			},
			mockReturnsFileExists: true,
			mockReturnedContent:   "current_version=1.2.3",
			inputSourceValue:      "current_version=1.2.3",
			wantResult:            false,
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
				Exists:  tt.mockReturnsFileExists,
			}
			f := &File{
				spec:             tt.spec,
				contentRetriever: &mockText,
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
		spec                  FileSpec
		wantFiles             []string
		name                  string
		inputSourceValue      string
		mockReturnedContent   string
		mockReturnedError     error
		scm                   scm.Scm
		wantResult            bool
		wantErr               bool
		mockReturnsFileExists bool
		dryRun                bool
		wantMockState         text.MockTextRetriever
	}{
		{
			name: "Passing Case with relative path",
			spec: FileSpec{
				File: "foo.txt",
				Line: 3,
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			mockReturnsFileExists: true,
			inputSourceValue:      "current_version=1.2.3",
			wantResult:            true,
			wantFiles:             []string{"/tmp/foo.txt"},
			wantMockState: text.MockTextRetriever{
				Location: "/tmp/foo.txt",
				Content:  "current_version=1.2.3",
				Line:     3,
			},
		},
		{
			name: "Passing Case with file created",
			spec: FileSpec{
				File:        "foo.txt",
				ForceCreate: true,
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			mockReturnsFileExists: false,
			inputSourceValue:      "current_version=1.2.3",
			wantResult:            true,
			wantFiles:             []string{"/tmp/foo.txt"},
			wantMockState: text.MockTextRetriever{
				Location: "/tmp/foo.txt",
				Content:  "current_version=1.2.3",
			},
		},
		{
			name:             "No line matched with matchPattern and ReplacePattern defined",
			inputSourceValue: "3.9.0",
			spec: FileSpec{
				File:           "foo.txt",
				MatchPattern:   "notmatching=*",
				ReplacePattern: "maven_version= 3.9.0",
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			mockReturnsFileExists: true,
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
				Exists: tt.mockReturnsFileExists,
			}
			f := &File{
				spec:             tt.spec,
				contentRetriever: &mockText,
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
