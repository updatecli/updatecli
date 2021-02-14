package dockerfile

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/olblak/updateCli/pkg/core/helpers"
	"github.com/olblak/updateCli/pkg/plugins/docker/dockerfile/types"
	"github.com/stretchr/testify/assert"
)

func TestDockerfile_Target(t *testing.T) {
	tests := []struct {
		name           string
		file           string
		instruction    types.Instruction
		value          string
		source         string
		dryRun         bool
		wantChanged    bool
		wantErrMessage string
		wantDiff       types.ChangedLines
	}{
		{
			name:   "FROM with text parser and dryrun",
			source: "1.16",
			file:   "FROM.Dockerfile",
			dryRun: true,
			instruction: map[string]interface{}{
				"keyword": "FROM",
				"matcher": "golang",
			},
			wantChanged: true,
		},
		{
			name:   "FROM with text parser",
			source: "1.16.0-alpine",
			file:   "FROM.Dockerfile",
			dryRun: false,
			instruction: map[string]interface{}{
				"keyword": "FROM",
				"matcher": "golang",
			},
			wantChanged: true,
			wantDiff: types.ChangedLines{
				1: types.LineDiff{
					Original: "FROM golang:1.15 AS builder",
					New:      "FROM golang:1.16.0-alpine AS builder",
				},
				11: types.LineDiff{
					Original: "FROM golang:1.15 as tester",
					New:      "FROM golang:1.16.0-alpine as tester",
				},
				15: types.LineDiff{
					Original: "FROM golang AS reporter",
					New:      "FROM golang:1.16.0-alpine AS reporter",
				},
				19: types.LineDiff{
					Original: "FROM golang",
					New:      "FROM golang:1.16.0-alpine",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// create a new empty temporary file in TMPDIR, which name starts with "tt.file"
			tmpfile, err := ioutil.TempFile("", tt.file)
			if err != nil {
				t.Fatal(err)
			}
			// clean up the temp file at the end of the test
			defer os.Remove(tmpfile.Name())

			// copy fixture file's content to temp file's content
			fixtureContent, err := ioutil.ReadFile(filepath.Join(".", "test_fixtures", tt.file))
			if err != nil {
				t.Fatal(err)
			}
			_, err = tmpfile.Write(fixtureContent)
			if err != nil {
				t.Fatal(err)
			}

			d := &Dockerfile{
				File:        tmpfile.Name(),
				Instruction: tt.instruction,
				Value:       tt.value,
				DryRun:      tt.dryRun,
			}
			gotChanged, gotErr := d.Target(tt.source, tt.dryRun)

			if tt.wantErrMessage != "" {
				assert.NotNil(t, gotErr)
				assert.Equal(t, tt.wantErrMessage, gotErr.Error())
			}

			assert.Equal(t, tt.wantChanged, gotChanged)

			// Verify current file's content
			currentDockerfileContent, err := helpers.ReadFile(tmpfile.Name())
			if err != nil {
				t.Fatal(err)
			}
			// Perform validation between the 2 byte slices when file is written
			if tt.dryRun {
				// Actual temp file must not have changed (dry run enabled)
				assert.Equal(t, fixtureContent, currentDockerfileContent)
			} else {
				fixtureLines := strings.Split(string(fixtureContent), "\n")
				currentDockerfileLines := strings.Split(string(currentDockerfileContent), "\n")

				// Both content should have the same number of line as only replacements are expected
				assert.Equal(t, len(fixtureLines), len(currentDockerfileLines))

				// The diff between fixture and resulting tmpfile must match the tt.wantedChanges
				for index, currentLine := range fixtureLines {
					lineDiff, exists := tt.wantDiff[index+1]
					if exists {
						// current's line number found in wantDiff: currentLine should have been changed
						assert.Equal(t, lineDiff.Original, currentLine)
						assert.Equal(t, lineDiff.New, currentDockerfileLines[index])
					} else {
						// current's line number NOT found in wantDiff lines should be the same
						assert.Equal(t, currentLine, currentDockerfileLines[index])
					}
				}
			}

		})
	}
}
