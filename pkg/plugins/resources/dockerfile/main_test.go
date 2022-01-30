package dockerfile

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile/mobyparser"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile/simpletextparser"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile/simpletextparser/keywords"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile/types"
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

const dockerfileFixture = `FROM golang:1.15 AS builder
ARG golang=3.0.0
COPY ./golang .
RUN go get -d -v ./... && echo golang
FROM golang
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
`
