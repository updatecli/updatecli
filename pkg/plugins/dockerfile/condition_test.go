package dockerfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func TestDockerfile_Condition(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		spec           Spec
		wantChanged    bool
		wantErrMessage string
	}{
		{
			name:   "Found FROM with text parser",
			source: "1.16",
			spec: Spec{
				File: "./test_fixtures/FROM.Dockerfile",
				Instruction: map[string]string{
					"keyword": "FROM",
					"matcher": "golang",
				},
			},
			wantChanged: true,
		},
		{
			name:   "Not Found ARG with moby parser",
			source: "golang:1.15",
			spec: Spec{
				File:        "./test_fixtures/FROM.Dockerfile",
				Instruction: "ARG[1][0]",
			},
			wantChanged: false,
		},
		{
			name:   "Non existent Dockerfile",
			source: "1.16",
			spec: Spec{
				File: "NOTEXISTING.Dockerfile",
				Instruction: map[string]string{
					"keyword": "FROM",
					"matcher": "golang",
				},
			},
			wantErrMessage: "open NOTEXISTING.Dockerfile: no such file or directory",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newParser, err := getParser(tt.spec)
			require.NoError(t, err)

			d := &Dockerfile{
				spec:   tt.spec,
				parser: newParser,
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

func TestDockerfile_ConditionFromSCM(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		file           string
		spec           Spec
		wantChanged    bool
		wantErrMessage string
		scm            scm.ScmHandler
	}{
		{
			name:   "Found FROM with text parser",
			source: "1.16",
			spec: Spec{
				File: "./test_fixtures/FROM.Dockerfile",
				Instruction: map[string]string{
					"keyword": "FROM",
					"matcher": "golang",
				},
			},
			// Should be true. TODO: mock with a content retriever
			wantChanged: false,
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newParser, err := getParser(tt.spec)
			require.NoError(t, err)

			d := &Dockerfile{
				spec:   tt.spec,
				parser: newParser,
			}

			gotChanged, gotErr := d.ConditionFromSCM(tt.source, tt.scm)
			if tt.wantErrMessage != "" {
				assert.NotNil(t, gotErr)
				assert.Equal(t, tt.wantErrMessage, gotErr.Error())
			}

			assert.Equal(t, tt.wantChanged, gotChanged)
		})
	}
}
