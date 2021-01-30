package dockerfile

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

type dataSet struct {
	dockerfile      string
	spec            Dockerfile
	expectedFound   bool
	expectedReplace bool
}

type positionKeyDataSet struct {
	key                         string
	expectedValue               bool
	expectedInstruction         string
	expectedInstructionPosition int
	expectedElementPosition     int
}

type positionKeyDataSets []positionKeyDataSet
type dataSets []dataSet

var (
	rawDockerfile string = `FROM ubuntu:20.04

#Simple labels
LABEL version="0.1"
LABEL maintainer="John Smith "
LABEL release-date="2020-04-05"
LABEL promoted="true"

#One line labels
LABEL com.example.version="0.0.1-beta" com.example.release-date="2015-02-12"

#Multi line label
LABEL vendor=ACME\ Incorporated \
      com.example.is-beta= \
      com.example.is-production="" \
      com.example.version="0.0.1-beta" \
      com.example.release-date="2015-02-12"


RUN echo "Hello World"
RUN \
	ls
RUN \
	echo true && \
	echo false && \
	echo true
`
	multiStageDockerfile string = `FROM golang:1.15 as builder

WORKDIR /go/src/app

COPY . .

RUN go get -d -v ./...

FROM ubuntu

LABEL maintainer="Olblak <me@olblak.com>"

VOLUME /tmp

RUN useradd -d /home/updatecli -U -u 1000 -m updatecli

RUN \
  apt-get update && \
  apt-get install -y ca-certificates && \
  apt-get clean && \
  find /var/lib/apt/lists -type f -delete

USER updatecli

WORKDIR /home/updatecli

COPY --from=builder --chown=updatecli:updatecli /go/src/app/bin/updateCli /usr/bin/updatecli

ENTRYPOINT [ "/usr/bin/updatecli" ]

CMD ["--help"]
`

	datas dataSets = []dataSet{
		{
			dockerfile: rawDockerfile,
			spec: Dockerfile{
				File:        "Dockerfile",
				Instruction: "From",
				Value:       "ubuntu:20.04",
				DryRun:      false,
			},
			expectedFound:   true,
			expectedReplace: false,
		},
		{
			dockerfile: rawDockerfile,
			spec: Dockerfile{
				File:        "Dockerfile",
				Instruction: "From[0][0]",
				Value:       "ubuntu:20.04",
				DryRun:      false,
			},
			expectedFound:   true,
			expectedReplace: false,
		},
		{
			dockerfile: rawDockerfile,
			spec: Dockerfile{
				File:        "Dockerfile",
				Instruction: "FROM",
				Value:       "ubuntu:20.04",
				DryRun:      false,
			},
			expectedFound:   true,
			expectedReplace: false,
		},
		{
			dockerfile: rawDockerfile,
			spec: Dockerfile{
				File:        "Dockerfile",
				Instruction: "FROM",
				Value:       "UBUNTU:20.04",
				DryRun:      false,
			},
			expectedFound:   true,
			expectedReplace: true,
		},
		{
			dockerfile: rawDockerfile,
			spec: Dockerfile{
				File:        "Dockerfile",
				Instruction: "From",
				Value:       "ubuntu:18.04",
				DryRun:      false,
			},
			expectedFound:   true,
			expectedReplace: true,
		},
		{
			dockerfile: rawDockerfile,
			spec: Dockerfile{
				File:        "Dockerfile",
				Instruction: "label[4][2]",
				Value:       "com.example.release-date",
				DryRun:      false,
			},
			expectedFound:   true,
			expectedReplace: false,
		},
		{
			dockerfile: multiStageDockerfile,
			spec: Dockerfile{
				File:        "Dockerfile",
				Instruction: "From",
				Value:       "golang:1.15",
				DryRun:      false,
			},
			expectedFound:   true,
			expectedReplace: false,
		},
		{
			dockerfile: multiStageDockerfile,
			spec: Dockerfile{
				File:        "Dockerfile",
				Instruction: "From[1][0]",
				Value:       "ubuntu",
				DryRun:      false,
			},
			expectedFound:   true,
			expectedReplace: false,
		},
		{
			dockerfile: multiStageDockerfile,
			spec: Dockerfile{
				File:        "Dockerfile",
				Instruction: "From[3][0]",
				Value:       "ubuntu",
				DryRun:      false,
			},
			expectedFound:   false,
			expectedReplace: false,
		},
		{
			dockerfile: multiStageDockerfile,
			spec: Dockerfile{
				File:        "Dockerfile",
				Instruction: "---",
				Value:       "",
				DryRun:      false,
			},
			expectedFound:   false,
			expectedReplace: false,
		},
		{
			dockerfile: multiStageDockerfile,
			spec: Dockerfile{
				File:        "Dockerfile",
				Instruction: "FROM[a][b]",
				Value:       "",
				DryRun:      false,
			},
			expectedFound:   false,
			expectedReplace: false,
		},
	}

	positionKeysdata positionKeyDataSets = positionKeyDataSets{
		{
			key:                         "LABEL[0][1]",
			expectedValue:               true,
			expectedInstruction:         "LABEL",
			expectedInstructionPosition: 0,
			expectedElementPosition:     1,
		},
		{
			key:                         "LABEL[0][1][2]",
			expectedValue:               true, // True at the key end with two [0][0]
			expectedInstruction:         "LABEL[0]",
			expectedInstructionPosition: 1,
			expectedElementPosition:     2,
		},
		{
			key:                         "LABEL[0]",
			expectedValue:               false,
			expectedInstruction:         "LABEL[0]",
			expectedInstructionPosition: 0,
			expectedElementPosition:     0,
		},
		{
			key:                         "LABEL[0][0]x",
			expectedValue:               false,
			expectedInstruction:         "LABEL[0][0]x",
			expectedInstructionPosition: 0,
			expectedElementPosition:     0,
		},
		{
			key:                         "LABEL[x][0]",
			expectedValue:               false,
			expectedInstruction:         "LABEL[x][0]",
			expectedInstructionPosition: 0,
			expectedElementPosition:     0,
		},
		{
			key:                         "LABEL",
			expectedValue:               false,
			expectedInstruction:         "LABEL",
			expectedInstructionPosition: 0,
			expectedElementPosition:     0,
		},
	}
)

func TestIsPositionKeys(t *testing.T) {
	for _, d := range positionKeysdata {
		got := isPositionKeys(d.key)
		if got != d.expectedValue {
			t.Errorf("Expected key:' %s' to be '%v', got %v",
				d.key,
				d.expectedValue,
				got)
		}

	}
}

func TestGetPositionKeys(t *testing.T) {
	for _, d := range positionKeysdata {
		gotKey, gotInstPos, gotElemPos, err := getPositionKeys(d.key)

		if err != nil {
			logrus.Errorf("err - %s", err)
		}

		if gotInstPos != d.expectedInstructionPosition {
			t.Errorf("Expected instruction position:' %s' to be '%v', got %v",
				d.key,
				d.expectedInstructionPosition,
				gotInstPos)
		}
		if gotElemPos != d.expectedElementPosition {
			t.Errorf("Expected element position:' %s' to be '%v', got %v",
				d.key,
				d.expectedElementPosition,
				gotElemPos)
		}
		if gotKey != d.expectedInstruction {
			t.Errorf("Expected key:' %s' to be '%v', got %v",
				d.key,
				d.expectedInstruction,
				gotKey)
		}

	}
}

func TestReplaceNode(t *testing.T) {
	for i, data := range datas {
		d, err := parser.Parse(bytes.NewReader([]byte(data.dockerfile)))

		if err != nil {
			logrus.Errorf("err - %s", err)
		}

		found, val, err := data.spec.replace(d.AST)

		if err != nil {
			logrus.Errorf("err - %s", err)
		}

		if found != data.expectedFound {
			t.Errorf("%d: Expected %s %s to be found, got %v",
				i,
				data.spec.Instruction,
				data.spec.Value,
				found)
		}
		if data.expectedReplace && val == data.spec.Value {
			t.Errorf("%d: Expected %s %s to be replace, got %v",
				i,
				data.spec.Instruction,
				data.spec.Value,
				found)
		}

	}

}
