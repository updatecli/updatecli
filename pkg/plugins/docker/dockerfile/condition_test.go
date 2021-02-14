package dockerfile

import (
	"testing"

	"github.com/olblak/updateCli/pkg/plugins/docker/dockerfile/types"
	"github.com/stretchr/testify/assert"
)

func TestDockerfile_Condition(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		file           string
		instruction    types.Instruction
		wantChanged    bool
		wantErrMessage string
	}{
		{
			name:   "Found FROM with text parser",
			source: "1.16",
			file:   "FROM.Dockerfile",
			instruction: map[string]interface{}{
				"keyword": "FROM",
				"matcher": "golang",
			},
			wantChanged: true,
		},
		{
			name:        "Found FROM with moby parser",
			source:      "golang:1.15",
			file:        "FROM.Dockerfile",
			instruction: "FROM[1][0]",
			wantChanged: true,
		},
		{
			name:   "Not Found ARG with text parser",
			source: "1.16",
			file:   "FROM.Dockerfile",
			instruction: map[string]interface{}{
				"keyword": "ARG",
				"matcher": "TERRAFORM_VERSION",
			},
			wantChanged: false,
		},
		{
			name:   "Failed to init Parser",
			source: "1.16",
			file:   "FROM.Dockerfile",
			instruction: map[string]int{
				"ARG": 0,
			},
			wantChanged:    false,
			wantErrMessage: "Parsing Error: cannot determine instruction: map[ARG:0].",
		},
		{
			name:   "Non existent Dockerfile",
			source: "1.16",
			file:   "NOTEXISTING.Dockerfile",
			instruction: map[string]interface{}{
				"keyword": "FROM",
				"matcher": "golang",
			},
			wantChanged:    false,
			wantErrMessage: "open ./test_fixtures/NOTEXISTING.Dockerfile: no such file or directory",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dockerfile{
				File:        "./test_fixtures/" + tt.file,
				Instruction: tt.instruction,
				Value:       tt.source, // Value and source are the same field in a condition
			}

			gotChanged, gotErr := d.Condition(tt.source)
			if tt.wantErrMessage != "" {
				assert.NotNil(t, gotErr)
				assert.Equal(t, tt.wantErrMessage, gotErr.Error())
			}

			assert.Equal(t, tt.wantChanged, gotChanged)
		})
	}
}
