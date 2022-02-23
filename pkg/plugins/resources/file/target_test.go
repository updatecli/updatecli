package file

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/text"
)

// TODO: find a way to test when there are multiple files but only some are changed
func TestFile_Target(t *testing.T) {
	tests := []struct {
		name                string
		spec                Spec
		files               map[string]string
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
			files: map[string]string{
				"foo.txt": "",
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
			wantResult: true,
		},
		{
			name: "Passing case with both input source and specified content but no line (specified content should be used)",
			spec: Spec{
				File:    "foo.txt",
				Content: "Hello World",
			},
			files: map[string]string{
				"foo.txt": "",
			},
			mockFileExists:   true,
			inputSourceValue: "current_version=1.2.3",
			wantMockState: text.MockTextRetriever{
				Location: "foo.txt",
				Content:  "Hello World",
			},
			wantResult: true,
		},
		// {
		// 	name: "Passing case with multiple 'Files' and both input source and specified content but no line (specified content should be used)",
		// 	spec: Spec{
		// 		Files: []string{
		// 			"foo.txt",
		// 			"bar.txt",
		// 		},
		// 		Content: "Hello World",
		// 	},
		// 	files: map[string]string{
		// 		"foo.txt": "",
		// 		"bar.txt": "",
		// 	},
		// 	mockFileExists:   true,
		// 	inputSourceValue: "current_version=1.2.3",
		// 	// TODO: check multiple locations
		// 	// wantMockState: text.MockTextRetriever{
		// 	// 	Location: "foo.txt",
		// 	// 	Content:  "Hello World",
		// 	// },
		// 	wantResult: true,
		// },
		{
			name: "Passing case with an updated line from provided content",
			spec: Spec{
				File:    "foo.txt",
				Content: "Hello World",
				Line:    2,
			},
			files: map[string]string{
				"foo.txt": "",
			},
			mockFileExists:      true,
			mockReturnedContent: "Title\nGood Bye\nThe end",
			inputSourceValue:    "current_version=1.2.3",
			wantMockState: text.MockTextRetriever{
				Location: "foo.txt",
				Line:     2,
				Content:  "Hello World",
			},
			wantResult: true,
		},
		{
			name: "Validation failure with an https:// URL instead of a file",
			spec: Spec{
				File: "https://github.com/foo.txt",
			},
			files: map[string]string{
				"https://github.com/foo.txt": "",
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
			files: map[string]string{
				"foo.txt": "",
			},
			wantResult: false,
			wantErr:    true,
		},
		{
			name: "Validation failure with invalid regexp for MatchPattern",
			spec: Spec{
				MatchPattern: "(d+:1",
				File:         "foo.txt",
			},
			files: map[string]string{
				"foo.txt": "",
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
			files: map[string]string{
				"not_existing.txt": "",
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
			files: map[string]string{
				"not_existing.txt": "",
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
			files: map[string]string{
				"not_existing.txt": "",
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
			files: map[string]string{
				"not_existing.txt": "",
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
			files: map[string]string{
				"foo.txt": "",
			},
			inputSourceValue:    "current_version=1.2.3",
			mockFileExists:      true,
			mockReturnedContent: "current_version=1.2.3",
			wantMockState: text.MockTextRetriever{
				Location: "foo.txt",
				Line:     3,
				Content:  "current_version=1.2.3",
			},
			wantResult: false,
		},
		{
			name: "File not updated (input source, no specified line)",
			spec: Spec{
				File: "foo.txt",
			},
			files: map[string]string{
				"foo.txt": "",
			},
			inputSourceValue:    "current_version=1.2.3",
			mockFileExists:      true,
			mockReturnedContent: "current_version=1.2.3",
			wantMockState: text.MockTextRetriever{
				Location: "foo.txt",
				Content:  "current_version=1.2.3",
			},
			wantResult: false,
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
				files:            tt.files,
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

// TODO: find a way to test when there are multiple files but only some are changed
func TestFile_TargetMultiples(t *testing.T) {
	tests := []struct {
		name                string
		spec                Spec
		files               map[string]string
		inputSourceValue    string
		mockReturnedContent map[string]string
		mockReturnedError   map[string]error
		mockFileExists      map[string]bool
		wantMockState       map[string]text.MockTextRetriever
		wantResult          map[string]bool
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
			files: map[string]string{
				"foo.txt": "",
			},
			mockFileExists: map[string]bool{
				"foo.txt": true,
			},
			mockReturnedContent: map[string]string{
				"foo.txt": `maven_version = "3.8.2"
				git_version = "2.33.1"
				jdk11_version = "11.0.12+7"
				jdk17_version = "17+35"
				jdk8_version = "8u302-b08"
				maven_major_release = "3"
				git_lfs_version = "3.0.1"
				compose_version = "1.29.2"`,
			},
			wantMockState: map[string]text.MockTextRetriever{
				"foo.txt": {
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
			wantResult: map[string]bool{
				"foo.txt": true,
			},
		},
		// {
		// 	name: "Passing case with multiple 'Files' and both input source and specified content but no line (specified content should be used)",
		// 	spec: Spec{
		// 		Files: []string{
		// 			"foo.txt",
		// 			"bar.txt",
		// 		},
		// 		Content: "Hello World",
		// 	},
		// 	files: map[string]string{
		// 		"foo.txt": "",
		// 		"bar.txt": "",
		// 	},
		// 	mockFileExists:   true,
		// 	inputSourceValue: "current_version=1.2.3",
		// 	// TODO: check multiple locations
		// 	// wantMockState: text.MockTextRetriever{
		// 	// 	Location: "foo.txt",
		// 	// 	Content:  "Hello World",
		// 	// },
		// 	wantResult: true,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTexts := make(map[string]text.MockTextRetriever)
			for filePath := range tt.files {
				mockTexts[filePath] = text.MockTextRetriever{
					Content: tt.mockReturnedContent[filePath],
					Err:     tt.mockReturnedError[filePath],
					Exists:  tt.mockFileExists[filePath],
				}
			}
			aMockText := mockTexts["foo.txt"]
			f := &File{
				spec:             tt.spec,
				contentRetriever: &aMockText,
				files:            tt.files,
			}

			gotResult, gotErr := f.Target(tt.inputSourceValue, tt.dryRun)
			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			for filePath := range tt.files {
				assert.Equal(t, tt.wantResult[filePath], gotResult)
				assert.Equal(t, tt.wantMockState[filePath].Location, mockTexts[filePath].Location)
				assert.Equal(t, tt.wantMockState[filePath].Line, mockTexts[filePath].Line)
				assert.Equal(t, tt.wantMockState[filePath].Content, mockTexts[filePath].Content)
			}
		})
	}
}

/*
func TestFile_TargetFromSCM(t *testing.T) {
	tests := []struct {
		name                string
		spec                Spec
		files               map[string]string
		scm                 scm.ScmHandler
		inputSourceValue    string
		mockFileExists      bool
		mockReturnedContent string
		mockReturnedError   error
		wantFiles           []string
		wantMockState       map[string]text.MockTextRetriever
		wantResult          bool
		wantErr             bool
		dryRun              bool
	}{
		// TODO: test with multiples files, same blocking 'location' issue as above, will probably need to keep track of all files in the mock
		{
			name: "Passing case with 'Line' specified",
			spec: Spec{
				File: "foo.txt",
				Line: 3,
			},
			files: map[string]string{
				"foo.txt": "",
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			inputSourceValue: "current_version=1.2.3",
			mockFileExists:   true,
			wantFiles:        []string{"/tmp/foo.txt"},
			wantMockState: map[string]text.MockTextRetriever{
				"/tmp/foo.txt": {
					Location: "/tmp/foo.txt",
					Content:  "current_version=1.2.3",
					Line:     3,
				},
			},
			wantResult: true,
		},
		{
			name: "Passing case with 'ForceCreate' specified",
			spec: Spec{
				File:        "foo.txt",
				ForceCreate: true,
			},
			files: map[string]string{
				"foo.txt": "",
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			inputSourceValue: "current_version=1.2.3",
			mockFileExists:   false,
			wantFiles:        []string{"/tmp/foo.txt"},
			wantMockState: map[string]text.MockTextRetriever{
				"/tmp/foo.txt": {
					Location: "/tmp/foo.txt",
					Content:  "current_version=1.2.3",
				},
			},
			wantResult: true,
		},
		{
			name: "No line matched with matchPattern and ReplacePattern defined",
			spec: Spec{
				File:           "foo.txt",
				MatchPattern:   "notmatching=*",
				ReplacePattern: "maven_version= 3.9.0",
			},
			files: map[string]string{
				"foo.txt": "",
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			inputSourceValue: "3.9.0",
			mockFileExists:   true,
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
			// mockTexts := make(map[string]text.TextRetriever)
			// for filePath, _ := range tt.files {
			// 	mockTexts[filePath] = tt.
			// }
			mockText := text.MockTextRetriever{
				Err:    tt.mockReturnedError,
				Exists: tt.mockFileExists,
			}
			f := &File{
				spec:             tt.spec,
				contentRetriever: &mockText,
				files:            tt.files,
			}

			gotResult, gotFiles, _, gotErr := f.TargetFromSCM(tt.inputSourceValue, tt.scm, tt.dryRun)

			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)

			assert.Equal(t, tt.wantResult, gotResult)
			assert.Equal(t, tt.wantFiles, gotFiles)
			for filePath := range f.files {
				assert.Equal(t, tt.wantMockState[filePath].Location, mockText.Location[filePath])
				assert.Equal(t, tt.wantMockState[filePath].Line, mockText[filePath].Line)
				assert.Equal(t, tt.wantMockState[filePath].Content, mockText[filePath].Content)
			}
		})
	}
}
*/
