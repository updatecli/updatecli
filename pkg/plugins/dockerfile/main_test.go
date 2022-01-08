package dockerfile

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/dockerfile/mobyparser"
	"github.com/updatecli/updatecli/pkg/plugins/dockerfile/simpletextparser"
	"github.com/updatecli/updatecli/pkg/plugins/dockerfile/simpletextparser/keywords"
	"github.com/updatecli/updatecli/pkg/plugins/dockerfile/types"
)

func TestDockerfile_New(t *testing.T) {
	tests := []struct {
		name       string
		spec       Spec
		wantParser types.DockerfileParser
		wantErr    error
	}{
		{
			name: "Simple Text Parser with strong typing",
			spec: Spec{
				Instruction: map[string]string{
					"keyword": "ARG",
					"matcher": "foo",
				},
			},
			wantParser: simpletextparser.SimpleTextDockerfileParser{
				Keyword:      "ARG",
				Matcher:      "foo",
				KeywordLogic: keywords.Arg{},
			},
		},
		{
			name: "Simple Text Parser with weak typing",
			spec: Spec{
				Instruction: map[string]interface{}{
					"keyword": "FROM",
					"matcher": "foo",
				},
			},
			wantParser: simpletextparser.SimpleTextDockerfileParser{
				Keyword:      "FROM",
				Matcher:      "foo",
				KeywordLogic: keywords.From{},
			},
		},
		{
			name: "Moby Parser",
			spec: Spec{
				Instruction: "ARG[0][1]",
				Value:       "HELM_VERSION",
			},
			wantParser: mobyparser.MobyParser{
				Instruction: "ARG[0][1]",
				Value:       "HELM_VERSION",
			},
		},
		{
			name: "Cannot determine parser",
			spec: Spec{
				Instruction: []string{"ARG", "TERRAFORM_VERSION"},
			},
			wantErr: fmt.Errorf("Parsing Error: cannot determine instruction: [ARG TERRAFORM_VERSION]."),
		},
		{
			name: "Simple Text Parser with weak typing and mixed types",
			spec: Spec{
				Instruction: map[string]interface{}{
					"keyword": "ARG",
					"matcher": 123,
				},
			},
			wantErr: fmt.Errorf("Parsing Error: cannot determine instruction: map[keyword:ARG matcher:123]."),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResource, gotErr := New(tt.spec)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantParser, gotResource.parser)

		})
	}
}

// {
// 	name:   "Not Found ARG with text parser",
// 	source: "1.16",
// 	file:   "FROM.Dockerfile",
// 	spec: Spec{
// 		File: "./test_fixtures/FROM.Dockerfile",
// 		Instruction: map[string]string{
// 			"keyword": "ARG",
// 			"matcher": "TERRAFORM_VERSION",
// 		},
// 	},
// 	wantChanged: false,
// },
// {
// 	name:   "Failed to init Parser",
// 	source: "1.16",
// 	file:   "FROM.Dockerfile",
// 	instruction: map[string]int{
// 		"ARG": 0,
// 	},
// 	wantChanged:    false,
// 	wantErrMessage: "Parsing Error: cannot determine instruction: map[ARG:0].",
// },

// func TestDockerfile_SetParser(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		instruction    types.Instruction
// 		value          string
// 		wantParser     types.DockerfileParser
// 		wantErrMessage string
// 	}{
// 		{
// 			name: "Simple Text Parser with strong typing",
// 			instruction: map[string]string{
// 				"keyword": "ARG",
// 				"matcher": "foo",
// 			},
// 			wantParser: simpletextparser.SimpleTextDockerfileParser{
// 				Keyword:      "ARG",
// 				Matcher:      "foo",
// 				KeywordLogic: keywords.Arg{},
// 			},
// 		},
// 		{
// 			name: "Simple Text Parser with weak typing",
// 			instruction: map[string]interface{}{
// 				"keyword": "FROM",
// 				"matcher": "foo",
// 			},
// 			wantParser: simpletextparser.SimpleTextDockerfileParser{
// 				Keyword:      "FROM",
// 				Matcher:      "foo",
// 				KeywordLogic: keywords.From{},
// 			},
// 		},
// 		{
// 			name:        "Moby Parser",
// 			instruction: "ARG[0][1]",
// 			value:       "HELM_VERSION",
// 			wantParser: mobyparser.MobyParser{
// 				Instruction: "ARG[0][1]",
// 				Value:       "HELM_VERSION",
// 			},
// 		},
// 		{
// 			name:           "Cannot determine parser",
// 			instruction:    []string{"ARG", "TERRAFORM_VERSION"},
// 			wantParser:     nil,
// 			wantErrMessage: "Parsing Error: cannot determine instruction: [ARG TERRAFORM_VERSION].",
// 		},
// 		{
// 			name: "Simple Text Parser with weak typing and mixed types",
// 			instruction: map[string]interface{}{
// 				"keyword": "ARG",
// 				"matcher": 123,
// 			},
// 			wantErrMessage: "Parsing Error: cannot determine instruction: map[keyword:ARG matcher:123].",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			d := &Dockerfile{
// 				spec: Spec{
// 					Instruction: tt.instruction,
// 					Value:       tt.value,
// 				},
// 			}

// 			gotErr := d.SetParser()

// 			if tt.wantErrMessage != "" {
// 				assert.NotNil(t, gotErr)
// 				assert.Equal(t, tt.wantErrMessage, gotErr.Error())
// 			}

// 			assert.Equal(t, tt.wantParser, d.parser)

// 		})
// 	}
// }
