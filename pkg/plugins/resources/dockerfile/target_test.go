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

func TestDockerfile_Target(t *testing.T) {
	tests := []struct {
		name             string
		inputSourceValue string
		dryRun           bool
		spec             Spec
		mockFile         text.MockTextRetriever
		wantChanged      bool
		wantErr          error
		wantMockState    text.MockTextRetriever
		scm              scm.ScmHandler
		files            []string
	}{
		{
			name:             "FROM with text parser and dryrun",
			inputSourceValue: "1.16",
			dryRun:           true,
			spec: Spec{
				Instruction: map[string]interface{}{
					"keyword": "FROM",
					"matcher": "golang",
				},
			},
			mockFile: text.MockTextRetriever{
				Contents: map[string]string{
					"FROM.Dockerfile": dockerfileFixture,
				},
			},
			files:       []string{"FROM.Dockerfile"},
			wantChanged: true,
			wantMockState: text.MockTextRetriever{
				Contents: map[string]string{
					"FROM.Dockerfile": dockerfileFixture,
				},
			},
		},
		{
			name:             "FROM with text parser and dryrun and scm",
			inputSourceValue: "1.16",
			dryRun:           true,
			spec: Spec{
				Instruction: map[string]interface{}{
					"keyword": "FROM",
					"matcher": "golang",
				},
			},
			files: []string{"FROM.Dockerfile"},
			mockFile: text.MockTextRetriever{
				Contents: map[string]string{
					"/tmp/FROM.Dockerfile": dockerfileFixture,
				},
			},
			wantChanged: true,
			wantMockState: text.MockTextRetriever{
				Contents: map[string]string{
					"/tmp/FROM.Dockerfile": dockerfileFixture,
				},
			},
			scm: &scm.MockScm{
				WorkingDir: "/tmp",
			},
		},
		{
			name:             "FROM with text parser",
			inputSourceValue: "1.16",
			spec: Spec{
				Instruction: map[string]interface{}{
					"keyword": "FROM",
					"matcher": "golang",
				},
			},
			mockFile: text.MockTextRetriever{
				Contents: map[string]string{
					"FROM.Dockerfile": dockerfileFixture,
				},
			},
			files:       []string{"FROM.Dockerfile"},
			wantChanged: true,
			wantMockState: text.MockTextRetriever{
				Contents: map[string]string{
					"FROM.Dockerfile": `FROM golang:1.16 AS builder
ARG golang=3.0.0
LABEL org.opencontainers.image.version=1.0.0
COPY ./golang .
RUN go get -d -v ./... && echo golang
FROM golang:1.16
WORKDIR /go/src/app
FROM ubuntu:20.04 AS golang
RUN apt-get update
FROM python AS python-runtime
FROM python-runtime AS final
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
			dryRun:           true,
			spec: Spec{
				Instruction: "FROM[12][1]",
			},
			files: []string{"FROM.Dockerfile"},
			mockFile: text.MockTextRetriever{
				Contents: map[string]string{
					"FROM.Dockerfile": dockerfileFixture,
				},
			},
			wantChanged: false,
			wantMockState: text.MockTextRetriever{
				Contents: map[string]string{
					"FROM.Dockerfile": dockerfileFixture,
				},
			},
			wantErr: fmt.Errorf("%s cannot find instruction \"FROM[12][1]\"", result.FAILURE),
		},
		{
			name:             "FROM with text parser, dryrun and no changed lines",
			inputSourceValue: "20.04",
			dryRun:           true,
			spec: Spec{
				Instruction: map[string]interface{}{
					"keyword": "FROM",
					"matcher": "ubuntu",
				},
			},
			files: []string{"FROM.Dockerfile"},
			mockFile: text.MockTextRetriever{
				Contents: map[string]string{
					"FROM.Dockerfile": dockerfileFixture,
				},
			},
			wantChanged: false,
			wantMockState: text.MockTextRetriever{
				// dryRun is true: no change
				Contents: map[string]string{
					"FROM.Dockerfile": dockerfileFixture,
				},
			},
		},
		{
			name:             "FROM with text parser and stage selection",
			inputSourceValue: "1.16",
			spec: Spec{
				Stage: "builder",
				Instruction: map[string]interface{}{
					"keyword": "FROM",
					"matcher": "golang",
				},
			},
			mockFile: text.MockTextRetriever{
				Contents: map[string]string{
					"FROM.Dockerfile": dockerfileFixture,
				},
			},
			files:       []string{"FROM.Dockerfile"},
			wantChanged: true,
			wantMockState: text.MockTextRetriever{
				Contents: map[string]string{
					"FROM.Dockerfile": `FROM golang:1.16 AS builder
ARG golang=3.0.0
LABEL org.opencontainers.image.version=1.0.0
COPY ./golang .
RUN go get -d -v ./... && echo golang
FROM golang
WORKDIR /go/src/app
FROM ubuntu:20.04 AS golang
RUN apt-get update
FROM python AS python-runtime
FROM python-runtime AS final
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
			name:             "FROM with stage name matching base image",
			inputSourceValue: "3.25",
			spec: Spec{
				Instruction: map[string]interface{}{
					"keyword": "FROM",
					"matcher": "python",
				},
			},
			files: []string{"FROM.Dockerfile"},
			mockFile: text.MockTextRetriever{
				Contents: map[string]string{
					"FROM.Dockerfile": dockerfileFixture,
				},
			},
			wantChanged: true,
			wantMockState: text.MockTextRetriever{
				// dryRun is true: no change
				Contents: map[string]string{
					"FROM.Dockerfile": `FROM golang:1.15 AS builder
ARG golang=3.0.0
LABEL org.opencontainers.image.version=1.0.0
COPY ./golang .
RUN go get -d -v ./... && echo golang
FROM golang
WORKDIR /go/src/app
FROM ubuntu:20.04 AS golang
RUN apt-get update
FROM python:3.25 AS python-runtime
FROM python-runtime AS final
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newParser, err := getParser(tt.spec)
			require.NoError(t, err)

			mockFile := tt.mockFile

			gotResult := result.Target{}

			d := &Dockerfile{
				spec:             tt.spec,
				contentRetriever: &mockFile,
				parser:           newParser,
				files:            tt.files,
			}
			gotErr := d.Target(tt.inputSourceValue, tt.scm, tt.dryRun, &gotResult)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantChanged, gotResult.Changed)
			assert.Equal(t, len(tt.files), len(gotResult.Files))
			for _, file := range tt.files {
				assert.Equal(t, tt.wantMockState.Contents[file], mockFile.Contents[file])
			}
		})
	}
}
