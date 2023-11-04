package dockerfile

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func TestDockerfile_Condition(t *testing.T) {
	tests := []struct {
		name             string
		inputSourceValue string
		spec             Spec
		wantChanged      bool
		wantErr          error
		scm              scm.ScmHandler
		files            []string
		mockTest         text.MockTextRetriever
	}{
		{
			name:             "Found FROM with text parser",
			inputSourceValue: "1.16",
			spec: Spec{
				Instruction: map[string]string{
					"keyword": "FROM",
					"matcher": "golang",
				},
			},
			files: []string{"FROM.Dockerfile"},
			mockTest: text.MockTextRetriever{
				Contents: map[string]string{
					"FROM.Dockerfile": dockerfileFixture,
				},
			},
			wantChanged: true,
		},
		{
			name:             "Found FROM with text parser and absolute path",
			inputSourceValue: "1.16",
			spec: Spec{
				Instruction: map[string]string{
					"keyword": "FROM",
					"matcher": "golang",
				},
			},
			files: []string{"FROM.Dockerfile"},
			mockTest: text.MockTextRetriever{
				Contents: map[string]string{
					"/tmp/FROM.Dockerfile": dockerfileFixture,
				},
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
			wantChanged: true,
		},
		{
			name:             "Not Found ARG with moby parser",
			inputSourceValue: "golang:1.15",
			spec: Spec{
				Instruction: "ARG[1][0]",
			},
			mockTest: text.MockTextRetriever{
				Contents: map[string]string{
					"FROM.Dockerfile": dockerfileFixture,
				},
			},
			files:       []string{"FROM.Dockerfile"},
			wantChanged: false,
		},
		{
			name:             "Nonexistent Dockerfile",
			inputSourceValue: "1.16",
			spec: Spec{
				Instruction: map[string]string{
					"keyword": "FROM",
					"matcher": "golang",
				},
			},
			files:   []string{"NONEXISTENT.Dockerfile"},
			wantErr: fmt.Errorf("the file NONEXISTENT.Dockerfile does not exist"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newParser, err := getParser(tt.spec)
			require.NoError(t, err)

			mockText := tt.mockTest

			d := &Dockerfile{
				spec:             tt.spec,
				parser:           newParser,
				contentRetriever: &mockText,
				files:            tt.files,
			}

			gotResult := result.Condition{}
			gotErr := d.Condition(tt.inputSourceValue, tt.scm, &gotResult)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, gotErr)
				return
			}

			require.NoError(t, gotErr)

			assert.Equal(t, tt.wantChanged, gotResult.Pass)
		})
	}
}
