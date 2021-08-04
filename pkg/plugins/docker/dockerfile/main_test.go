package dockerfile

import (
	"testing"

	"github.com/updatecli/updatecli/pkg/plugins/docker/dockerfile/mobyparser"
	"github.com/updatecli/updatecli/pkg/plugins/docker/dockerfile/simpletextparser"
	"github.com/updatecli/updatecli/pkg/plugins/docker/dockerfile/simpletextparser/keywords"
	"github.com/updatecli/updatecli/pkg/plugins/docker/dockerfile/types"
	"github.com/stretchr/testify/assert"
)

func TestDockerfile_SetParser(t *testing.T) {
	tests := []struct {
		name           string
		instruction    types.Instruction
		value          string
		wantParser     types.DockerfileParser
		wantErrMessage string
	}{
		{
			name: "Simple Text Parser with strong typing",
			instruction: map[string]string{
				"keyword": "ARG",
				"matcher": "foo",
			},
			wantParser: simpletextparser.SimpleTextDockerfileParser{
				Keyword:      "ARG",
				Matcher:      "foo",
				KeywordLogic: keywords.Arg{},
			},
		},
		{
			name: "Simple Text Parser with weak typing",
			instruction: map[string]interface{}{
				"keyword": "FROM",
				"matcher": "foo",
			},
			wantParser: simpletextparser.SimpleTextDockerfileParser{
				Keyword:      "FROM",
				Matcher:      "foo",
				KeywordLogic: keywords.From{},
			},
		},
		{
			name:        "Moby Parser",
			instruction: "ARG[0][1]",
			value:       "HELM_VERSION",
			wantParser: mobyparser.MobyParser{
				Instruction: "ARG[0][1]",
				Value:       "HELM_VERSION",
			},
		},
		{
			name:           "Cannot determine parser",
			instruction:    []string{"ARG", "TERRAFORM_VERSION"},
			wantParser:     nil,
			wantErrMessage: "Parsing Error: cannot determine instruction: [ARG TERRAFORM_VERSION].",
		},
		{
			name: "Simple Text Parser with weak typing and mixed types",
			instruction: map[string]interface{}{
				"keyword": "ARG",
				"matcher": 123,
			},
			wantErrMessage: "Parsing Error: cannot determine instruction: map[keyword:ARG matcher:123].",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dockerfile{
				Instruction: tt.instruction,
				Value:       tt.value,
			}

			gotErr := d.SetParser()

			if tt.wantErrMessage != "" {
				assert.NotNil(t, gotErr)
				assert.Equal(t, tt.wantErrMessage, gotErr.Error())
			}

			assert.Equal(t, tt.wantParser, d.parser)

		})
	}
}
