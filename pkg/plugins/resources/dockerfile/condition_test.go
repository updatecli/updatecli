package dockerfile

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func TestDockerfile_Condition(t *testing.T) {
	tests := []struct {
		name             string
		inputSourceValue string
		spec             Spec
		wantChanged      bool
		wantErr          error
		mockTest         text.MockTextRetriever
	}{
		{
			name:             "Found FROM with text parser",
			inputSourceValue: "1.16",
			spec: Spec{
				File: "FROM.Dockerfile",
				Instruction: map[string]string{
					"keyword": "FROM",
					"matcher": "golang",
				},
			},
			mockTest: text.MockTextRetriever{
				Contents: map[string]string{
					"FROM.Dockerfile": dockerfileFixture,
				},
			},
			wantChanged: true,
		},
		{
			name:             "Not Found ARG with moby parser",
			inputSourceValue: "golang:1.15",
			spec: Spec{
				File:        "FROM.Dockerfile",
				Instruction: "ARG[1][0]",
			},
			mockTest: text.MockTextRetriever{
				Contents: map[string]string{
					"FROM.Dockerfile": dockerfileFixture,
				},
			},
			wantChanged: false,
		},
		{
			name:             "Non existent Dockerfile",
			inputSourceValue: "1.16",
			spec: Spec{
				File: "NOTEXISTING.Dockerfile",
				Instruction: map[string]string{
					"keyword": "FROM",
					"matcher": "golang",
				},
			},
			wantErr: fmt.Errorf("The file NOTEXISTING.Dockerfile does not exist."),
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
			}

			gotChanged, gotErr := d.Condition(tt.inputSourceValue)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, gotErr)
				return
			}

			require.NoError(t, gotErr)

			assert.Equal(t, tt.wantChanged, gotChanged)
		})
	}
}

func TestDockerfile_ConditionFromSCM(t *testing.T) {
	tests := []struct {
		name             string
		inputSourceValue string
		spec             Spec
		wantChanged      bool
		wantErr          error
		scm              scm.ScmHandler
		mockTest         text.MockTextRetriever
	}{
		{
			name:             "Found FROM with text parser",
			inputSourceValue: "1.16",
			spec: Spec{
				File: "FROM.Dockerfile",
				Instruction: map[string]string{
					"keyword": "FROM",
					"matcher": "golang",
				},
			},
			mockTest: text.MockTextRetriever{
				Contents: map[string]string{
					"/tmp/FROM.Dockerfile": dockerfileFixture,
				},
			},
			wantChanged: true,
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
		},
		{
			name:             "Non existent Dockerfile",
			inputSourceValue: "1.16",
			spec: Spec{
				File: "NOTEXISTING.Dockerfile",
				Instruction: map[string]string{
					"keyword": "FROM",
					"matcher": "golang",
				},
			},
			scm: &scm.MockScm{
				WorkingDir: "/foo",
			},
			wantErr: fmt.Errorf("The file /foo/NOTEXISTING.Dockerfile does not exist."),
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
			}

			gotChanged, gotErr := d.ConditionFromSCM(tt.inputSourceValue, tt.scm)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, gotErr)
				return
			}

			require.NoError(t, gotErr)

			assert.Equal(t, tt.wantChanged, gotChanged)
		})
	}
}
