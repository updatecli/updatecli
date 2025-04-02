package dockerfile

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func TestDockerfile_Source(t *testing.T) {
	tests := []struct {
		name           string
		spec           Spec
		expectedResult string
		mockFile       text.MockTextRetriever
		wantErr        error
		parserErr      bool
		files          []string
	}{
		{
			name: "LABEL Uppercase",
			spec: Spec{
				Instruction: map[string]interface{}{
					"keyword": "LABEL",
					"matcher": "org.opencontainers.image.version",
				},
			},
			expectedResult: "1.0.0",
			mockFile: text.MockTextRetriever{
				Contents: map[string]string{
					"SOURCE.Dockerfile": dockerfileFixture,
				},
			},
			files: []string{"SOURCE.Dockerfile"},
		},
		{
			name: "FROM Uppercase",
			spec: Spec{
				Stage: "builder",
				Instruction: map[string]interface{}{
					"keyword": "FROM",
					"matcher": "golang",
				},
			},
			expectedResult: "golang:1.15",
			mockFile: text.MockTextRetriever{
				Contents: map[string]string{
					"SOURCE.Dockerfile": dockerfileFixture,
				},
			},
			files: []string{"SOURCE.Dockerfile"},
		},
		{
			name: "LABEL lowercase",
			spec: Spec{
				Instruction: map[string]interface{}{
					"keyword": "label",
					"matcher": "org.opencontainers.image.version",
				},
			},
			expectedResult: "1.0.0",
			mockFile: text.MockTextRetriever{
				Contents: map[string]string{
					"SOURCE.Dockerfile": dockerfileFixture,
				},
			},
			files: []string{"SOURCE.Dockerfile"},
		},
		{
			name: "Unsupported instruction",
			spec: Spec{
				Instruction: map[string]interface{}{
					"keyword": "quack",
					"matcher": "org.opencontainers.image.version",
				},
			},
			parserErr: true,
			wantErr:   fmt.Errorf("✗ Unknown keyword \"quack\" provided for Dockerfile's instruction"),
			mockFile: text.MockTextRetriever{
				Contents: map[string]string{
					"SOURCE.Dockerfile": dockerfileFixture,
				},
			},
			files: []string{"SOURCE.Dockerfile"},
		},
		{
			name: "Missing instruction",
			spec: Spec{
				Instruction: map[string]interface{}{
					"keyword": "LABEL",
					"matcher": "org.opencontainers.image.source",
				},
			},
			expectedResult: "",
			mockFile: text.MockTextRetriever{
				Contents: map[string]string{
					"SOURCE.Dockerfile": dockerfileFixture,
				},
			},
			files: []string{"SOURCE.Dockerfile"},
		},
		{
			name: "More than one files",
			spec: Spec{
				Instruction: map[string]interface{}{
					"keyword": "LABEL",
					"matcher": "org.opencontainers.image.source",
				},
			},
			wantErr: fmt.Errorf("validation error in sources of type 'dockerfile': the attributes `spec.files` can't contain more than one element for sources"),
			mockFile: text.MockTextRetriever{
				Contents: map[string]string{
					"SOURCE.Dockerfile": dockerfileFixture,
					"LABEL2.Dockerfile": dockerfileFixture,
				},
			},
			files: []string{"SOURCE.Dockerfile", "LABEL2.Dockerfile"},
		}, {
			name: "No files specified",
			spec: Spec{
				Instruction: map[string]interface{}{
					"keyword": "LABEL",
					"matcher": "org.opencontainers.image.source",
				},
			},
			wantErr: fmt.Errorf("no dockerfile specified"),
			mockFile: text.MockTextRetriever{
				Contents: map[string]string{
					"SOURCE.Dockerfile": dockerfileFixture,
				},
			},
			files: []string{},
		}, {
			name: "Bad Path",
			spec: Spec{
				Instruction: map[string]interface{}{
					"keyword": "LABEL",
					"matcher": "org.opencontainers.image.source",
				},
			},
			wantErr: fmt.Errorf("the file LABEL2.Dockerfile does not exist"),
			mockFile: text.MockTextRetriever{
				Contents: map[string]string{
					"SOURCE.Dockerfile": dockerfileFixture,
				},
			},
			files: []string{"LABEL2.Dockerfile"},
		},
		{
			name: "MultiStage",
			spec: Spec{
				Stage: "builder",
				Instruction: map[string]interface{}{
					"keyword": "LABEL",
					"matcher": "org.opencontainers.image.version",
				},
			},
			expectedResult: "0.1.0",
			mockFile: text.MockTextRetriever{
				Contents: map[string]string{
					"SOURCE.Dockerfile": `FROM golang:1.15 AS builder
LABEL org.opencontainers.image.version=0.1.0
FROM golang
LABEL org.opencontainers.image.version=1.0.0
`,
				},
			},
			files: []string{"SOURCE.Dockerfile"},
		},
		{
			name: "Non existing stage",
			spec: Spec{
				Stage: "tester",
				Instruction: map[string]interface{}{
					"keyword": "LABEL",
					"matcher": "org.opencontainers.image.version",
				},
			},
			expectedResult: "",
			mockFile: text.MockTextRetriever{
				Contents: map[string]string{
					"SOURCE.Dockerfile": `FROM golang:1.15 AS builder
LABEL org.opencontainers.image.version=0.1.0
FROM golang
LABEL org.opencontainers.image.version=1.0.0
`,
				},
			},
			files: []string{"SOURCE.Dockerfile"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFile := tt.mockFile
			newParser, gotErr := getParser(tt.spec)
			if tt.parserErr {
				assert.Error(t, gotErr)
				assert.Equal(t, tt.wantErr, gotErr)
				return
			} else {
				require.NoError(t, gotErr)
			}
			d := &Dockerfile{
				spec:             tt.spec,
				contentRetriever: &mockFile,
				parser:           newParser,
				files:            tt.files,
			}
			gotResult := result.Source{}
			gotErr = d.Source("", &gotResult)

			if tt.wantErr != nil {
				assert.Error(t, gotErr)
				assert.Equal(t, tt.wantErr, gotErr)
			} else {
				require.NoError(t, gotErr)
			}

			assert.Equal(t, tt.expectedResult, gotResult.Information)
		})
	}
}
