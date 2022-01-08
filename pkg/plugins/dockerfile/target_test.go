package dockerfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/dockerfile/types"
)

func TestDockerfile_Target(t *testing.T) {
	tests := []struct {
		name                  string
		source                string
		spec                  Spec
		mockReturnedContent   string
		mockReturnedError     error
		mockReturnsFileExists bool
		dryRun                bool
		wantChanged           bool
		wantErr               error
		wantDiff              types.ChangedLines
		wantMockState         text.MockTextRetriever
	}{
		{
			name:   "FROM with text parser and dryrun",
			source: "1.16",
			dryRun: true,
			spec: Spec{
				File: "FROM.Dockerfile",
				Instruction: map[string]interface{}{
					"keyword": "FROM",
					"matcher": "golang",
				},
			},
			mockReturnedContent:   fromDockerfileFixture,
			mockReturnsFileExists: true,
			wantChanged:           true,
		},
		// {
		// 	name:   "FROM with text parser",
		// 	source: "1.16.0-alpine",
		// 	file:   "FROM.Dockerfile",
		// 	dryRun: false,
		// 	instruction: map[string]interface{}{
		// 		"keyword": "FROM",
		// 		"matcher": "golang",
		// 	},
		// 	wantChanged: true,
		// 	wantDiff: types.ChangedLines{
		// 		1: types.LineDiff{
		// 			Original: "FROM golang:1.15 AS builder",
		// 			New:      "FROM golang:1.16.0-alpine AS builder",
		// 		},
		// 		11: types.LineDiff{
		// 			Original: "FROM golang:1.15 as tester",
		// 			New:      "FROM golang:1.16.0-alpine as tester",
		// 		},
		// 		15: types.LineDiff{
		// 			Original: "FROM golang AS reporter",
		// 			New:      "FROM golang:1.16.0-alpine AS reporter",
		// 		},
		// 		19: types.LineDiff{
		// 			Original: "FROM golang",
		// 			New:      "FROM golang:1.16.0-alpine",
		// 		},
		// 	},
		// },
		// {
		// 	name:        "FROM with moby parser and dryrun",
		// 	source:      "golang:1.16-alpine",
		// 	file:        "FROM.Dockerfile",
		// 	dryRun:      true,
		// 	instruction: "FROM[1][0]",
		// 	wantChanged: true,
		// },
		// {
		// 	name:        "FROM with moby parser",
		// 	source:      "golang:1.16-alpine",
		// 	file:        "FROM.Dockerfile",
		// 	dryRun:      true,
		// 	instruction: "FROM[1][0]",
		// 	wantChanged: true,
		// 	wantDiff: types.ChangedLines{
		// 		11: types.LineDiff{
		// 			Original: "FROM golang:1.15 as tester",
		// 			New:      "FROM golang:1.16-alpine as tester",
		// 		},
		// 	},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newParser, err := getParser(tt.spec)
			require.NoError(t, err)

			mockText := text.MockTextRetriever{
				Content: tt.mockReturnedContent,
				Err:     tt.mockReturnedError,
				Exists:  tt.mockReturnsFileExists,
			}

			d := &Dockerfile{
				spec:             tt.spec,
				contentRetriever: &mockText,
				parser:           newParser,
			}
			gotChanged, gotErr := d.Target(tt.source, tt.dryRun)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantResult, gotChanged)
			assert.Equal(t, tt.wantMockState.Location, mockText.Location)
			assert.Equal(t, tt.wantMockState.Line, mockText.Line)
			assert.Equal(t, tt.wantMockState.Content, mockText.Content)

			// if tt.wantErrMessage != nil {
			// 	assert.NotNil(t, gotErr)
			// 	assert.Equal(t, tt.wantErrMessage, gotErr.Error())
			// }

			// assert.Equal(t, tt.wantChanged, gotChanged)

			// Verify current file's content
			// currentDockerfileContent, err := helpers.ReadFile(tmpfile.Name())
			// if err != nil {
			// 	t.Fatal(err)
			// }
			// // Perform validation between the 2 byte slices when file is written
			// if tt.dryRun {
			// 	// Actual temp file must not have changed (dry run enabled)
			// 	assert.Equal(t, fixtureContent, currentDockerfileContent)
			// } else {
			// 	fixtureLines := strings.Split(string(fixtureContent), "\n")
			// 	currentDockerfileLines := strings.Split(string(currentDockerfileContent), "\n")

			// 	// Both content should have the same number of line as only replacements are expected
			// 	assert.Equal(t, len(fixtureLines), len(currentDockerfileLines))

			// 	// The diff between fixture and resulting tmpfile must match the tt.wantedChanges
			// 	for index, currentLine := range fixtureLines {
			// 		lineDiff, exists := tt.wantDiff[index+1]
			// 		if exists {
			// 			// current's line number found in wantDiff: currentLine should have been changed
			// 			assert.Equal(t, lineDiff.Original, currentLine)
			// 			assert.Equal(t, lineDiff.New, currentDockerfileLines[index])
			// 		} else {
			// 			// current's line number NOT found in wantDiff lines should be the same
			// 			assert.Equal(t, currentLine, currentDockerfileLines[index])
			// 		}
			// 	}
			// }

		})
	}
}
