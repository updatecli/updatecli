package dockerfile

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func TestDockerfile_Target(t *testing.T) {
	tests := []struct {
		name             string
		inputSourceValue string
		workingDir       string
		dryRun           bool
		spec             Spec
		mockFile         text.MockTextRetriever
		wantChanged      bool
		wantErr          error
		wantMockState    text.MockTextRetriever
		wantFiles        []string
		wantMessage      string
	}{
		{
			name:             "FROM with text parser and dryrun",
			inputSourceValue: "1.16",
			workingDir:       "/tmp",
			dryRun:           true,
			spec: Spec{
				File: "FROM.Dockerfile",
				Instruction: map[string]interface{}{
					"keyword": "FROM",
					"matcher": "golang",
				},
			},
			mockFile: text.MockTextRetriever{
				Contents: map[string]string{
					"/tmp/FROM.Dockerfile": dockerfileFixture,
				},
			},
			wantChanged: true,
			wantFiles:   []string{"FROM.Dockerfile"},
			wantMessage: "changed lines [1 5] of file \"FROM.Dockerfile\"",
			wantMockState: text.MockTextRetriever{
				// dryRun is true: no change
				Contents: map[string]string{
					"/tmp/FROM.Dockerfile": dockerfileFixture,
				},
			},
		},
		{
			name:             "FROM with text parser and no dry run",
			inputSourceValue: "1.16",
			workingDir:       "/",
			spec: Spec{
				File: "foo/FROM.Dockerfile",
				Instruction: map[string]interface{}{
					"keyword": "FROM",
					"matcher": "golang",
				},
			},
			mockFile: text.MockTextRetriever{
				Contents: map[string]string{
					"/foo/FROM.Dockerfile": dockerfileFixture,
				},
			},
			wantChanged: true,
			wantFiles:   []string{"foo/FROM.Dockerfile"},
			wantMessage: "changed lines [1 5] of file \"foo/FROM.Dockerfile\"",
			wantMockState: text.MockTextRetriever{
				Contents: map[string]string{
					"/foo/FROM.Dockerfile": `FROM golang:1.16 AS builder
		ARG golang=3.0.0
		COPY ./golang .
		RUN go get -d -v ./... && echo golang
		FROM golang:1.16
		WORKDIR /go/src/app
		FROM ubuntu:20.04 AS golang
		RUN apt-get update
		FROM ubuntu:20.04
		RUN apt-get update
		LABEL golang="${GOLANG_VERSION}"
		VOLUME /tmp / golang
		USER golang
		WORKDIR /home/updatecli
		COPY --from=golang --chown=updatecli:golang /go/src/app/dist/updatecli /usr/bin/golang
		ENTRYPOINT [ "/usr/bin/golang" ]
		CMD ["--help:golang"]
		`,
				},
			},
		},
		{
			name:             "FROM with moby parser, dryrun, instruction not found",
			inputSourceValue: "1.16",
			workingDir:       "/tmp",
			dryRun:           true,
			spec: Spec{
				File:        "FROM.Dockerfile",
				Instruction: "FROM[12][1]",
			},
			mockFile: text.MockTextRetriever{
				Contents: map[string]string{
					"/tmp/FROM.Dockerfile": dockerfileFixture,
				},
			},
			wantChanged: false,
			wantMockState: text.MockTextRetriever{
				Contents: map[string]string{
					"/tmp/FROM.Dockerfile": dockerfileFixture,
				},
			},
			wantErr: fmt.Errorf("%s cannot find instruction \"FROM[12][1]\"", result.FAILURE),
		},
		{
			name:             "FROM with text parser, dryrun and no changed lines",
			inputSourceValue: "20.04",
			workingDir:       "/foo",
			dryRun:           true,
			spec: Spec{
				File: "FROM.Dockerfile",
				Instruction: map[string]interface{}{
					"keyword": "FROM",
					"matcher": "ubuntu",
				},
			},
			mockFile: text.MockTextRetriever{
				Contents: map[string]string{
					"/foo/FROM.Dockerfile": dockerfileFixture,
				},
			},
			wantChanged: false,
			// wantFiles:   []string{"FROM.Dockerfile"},
			// wantMessage: "changed lines [1 5] of file \"FROM.Dockerfile\"",
			wantMockState: text.MockTextRetriever{
				// dryRun is true: no change
				Contents: map[string]string{
					"/foo/FROM.Dockerfile": dockerfileFixture,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newParser, err := getParser(tt.spec)
			require.NoError(t, err)

			mockFile := &tt.mockFile

			d := &Dockerfile{
				spec:             tt.spec,
				contentRetriever: mockFile,
				parser:           newParser,
			}
			gotChanged, gotFiles, gotMessage, gotErr := d.Target(tt.inputSourceValue, tt.workingDir, tt.dryRun)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantChanged, gotChanged)
			assert.Equal(t, tt.wantMockState.Contents[tt.spec.File], mockFile.Contents[tt.spec.File])
			assert.Equal(t, tt.wantFiles, gotFiles)
			assert.Equal(t, tt.wantMessage, gotMessage)
		})
	}
}
