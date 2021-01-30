package dockerfile

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

type MarshalDataSet struct {
	dockerfile         string
	expectedResult     bool
	expectedError      bool
	expectedDockerfile string
}

type MarshalDataSets []MarshalDataSet

var (
	expectedRawDockerfile = `FROM ubuntu:20.04

LABEL version="0.1"

LABEL maintainer="John Smith "

LABEL release-date="2020-04-05"

LABEL promoted="true"

LABEL com.example.version="0.0.1-beta"\
      com.example.release-date="2015-02-12"

LABEL vendor=ACME\ Incorporated\
      com.example.is-beta=\
      com.example.is-production=""\
      com.example.version="0.0.1-beta"\
      com.example.release-date="2015-02-12"

RUN echo "Hello World"

RUN ls

RUN echo true && \
	echo false && \
	echo true
`
	marshalData = MarshalDataSets{
		{
			dockerfile:         "#Comment\nFROM golang:1.15 as builder\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "FROM golang:1.15 as builder\n",
		},
		{
			dockerfile:         "# Comment should be remove\n",
			expectedResult:     false,
			expectedError:      true,
			expectedDockerfile: "",
		},
		{
			dockerfile:         "MAINTAINER olblak\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "MAINTAINER olblak\n",
		},
		{
			dockerfile:         "EXPOSE 80/tcp\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "EXPOSE 80/tcp\n",
		},
		{
			dockerfile:         "ENV key=value\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "ENV key=value\n",
		},
		{
			dockerfile:         "ENV key=value key=value\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "ENV key=value\\\n    key=value\n",
		},
		{
			dockerfile:         "ENV key value\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "ENV key=value\n",
		},
		{
			dockerfile:         "ADD hom* /mydir/\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "ADD hom* /mydir/\n",
		},
		{
			dockerfile:         "ADD arr[[]0].txt /mydir/\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "ADD arr[[]0].txt /mydir/\n",
		},
		{
			dockerfile:         "ADD --chown=55:mygroup files* /somedir/\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "ADD --chown=55:mygroup files* /somedir/\n",
		},
		{
			dockerfile:         "ONBUILD RUN /usr/local/bin/python-build --dir /app/src\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "ONBUILD RUN /usr/local/bin/python-build --dir /app/src\n",
		},
		{
			dockerfile:         "ONBUILD ADD . /app/src\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "ONBUILD ADD . /app/src\n",
		},
		{
			dockerfile:         "STOPSIGNAL SIGKILL\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "STOPSIGNAL SIGKILL\n",
		},
		{
			dockerfile:         "STOPSIGNAL 9\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "STOPSIGNAL 9\n",
		},
		{
			dockerfile:         "HEALTHCHECK --interval=5m --timeout=3s CMD curl -f http://localhost/ || exit 1\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "HEALTHCHECK --interval=5m --timeout=3s CMD curl -f http://localhost/ || exit 1\n",
		},
		{
			dockerfile:         "ARG user1=someuser\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "ARG user1=someuser\n",
		},
		{
			dockerfile:         "ARG user1\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "ARG user1\n",
		},
		{
			dockerfile:         "STOPSIGNAL 9\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "STOPSIGNAL 9\n",
		},
		{
			dockerfile:         "COPY hom* /mydir/\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "COPY hom* /mydir/\n",
		},
		{
			dockerfile:         "COPY arr[[]0].txt /mydir/\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "COPY arr[[]0].txt /mydir/\n",
		},
		{
			dockerfile:         "COPY --chown=55:mygroup files* /somedir/\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "COPY --chown=55:mygroup files* /somedir/\n",
		},
		{
			dockerfile:         "RUN echo true\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "RUN echo true\n",
		},
		{
			dockerfile:         "RUN [\"/bin/bash\", \"-c\", \"echo hello\"]\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "RUN /bin/bash -c echo hello\n",
		},
		{
			dockerfile:         "CMD echo true\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "CMD [ \"echo true\" ]\n",
		},
		{
			dockerfile:         "CMD [\"/bin/bash\", \"-c\", \"echo hello\"]\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "CMD [ \"/bin/bash\",\"-c\",\"echo hello\" ]\n",
		},
		{
			dockerfile:         "ENTRYPOINT echo true\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "ENTRYPOINT [ \"echo true\" ]\n",
		},
		{
			dockerfile:         "ENTRYPOINT \"/usr/local/bin/helmfile\"\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "ENTRYPOINT [ \"/usr/local/bin/helmfile\" ]\n",
		},
		{
			dockerfile:         "ENTRYPOINT [\"/bin/bash\", \"-c\", \"echo hello\"]\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "ENTRYPOINT [ \"/bin/bash\",\"-c\",\"echo hello\" ]\n",
		},
		{
			dockerfile:         "SHELL [\"powershell\", \"-command\"]\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "SHELL [ \"powershell\",\"-command\" ]\n",
		},
		{
			dockerfile:         "VOLUME [\"/data\"]\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "VOLUME /data\n",
		},
		{
			dockerfile:         "VOLUME /data\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "VOLUME /data\n",
		},
		{
			dockerfile:         "USER patrick\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "USER patrick\n",
		},
		{
			dockerfile:         "USER 1000:1000\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "USER 1000:1000\n",
		},
		{
			dockerfile:         "WORKDIR $DIRPATH/$DIRNAME\n",
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: "WORKDIR $DIRPATH/$DIRNAME\n",
		},
		{
			dockerfile:         rawDockerfile,
			expectedResult:     true,
			expectedError:      false,
			expectedDockerfile: expectedRawDockerfile,
		},
	}
)

func TestMarshal(t *testing.T) {
	for _, data := range marshalData {

		d, err := parser.Parse(bytes.NewReader([]byte(data.dockerfile)))

		if err != nil && !data.expectedError {
			t.Errorf("%v", err)
		}
		if err == nil && !data.expectedError {

			gotDockerfile := ""

			err = Marshal(d, &gotDockerfile)
			if err != nil {
				logrus.Errorf("err - %s", err)
			}

			if (gotDockerfile != data.expectedDockerfile) == data.expectedResult {
				//t.Errorf("Raw new Dockerfile\nGot:\n%q\n%s\nExpected:\n%q\n%s\n",
				t.Errorf("Raw new Dockerfile\nGot:\n%v\n%s\nExpected:\n%v\n%s\n",
					gotDockerfile,
					strings.Repeat("=", 10),
					data.expectedDockerfile,
					strings.Repeat("=", 10))
			}
		}
	}
}
