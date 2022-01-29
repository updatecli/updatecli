package mavenmetadata

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
)

const wikitextCoreMavenMetadata = `<?xml version="1.0" encoding="UTF-8"?>
<metadata>
  <groupId>org.eclipse.mylyn.wikitext</groupId>
  <artifactId>wikitext.core</artifactId>
  <version>1.7.4.v20130429</version>
  <versioning>
    <latest>1.7.4.v20130429</latest>
    <release>1.7.4.v20130429</release>
    <versions>
      <version>1.7.4.v20130429</version>
			<version>1.7.3</version>
			<version>1.7.2</version>
			<version>1.7.1</version>
			<version>1.7.0</version>
    </versions>
    <lastUpdated>20130619211401</lastUpdated>
  </versioning>
</metadata>
`

const invalidXML = `<?xml version="1.0" encoding="UTF-8"?>
<metadata>
`

const noLatestVersionMavenMetadata = `<?xml version="1.0" encoding="UTF-8"?>
<metadata>
  <groupId>org.eclipse.mylyn.wikitext</groupId>
  <artifactId>wikitext.core</artifactId>
  <version></version>
  <versioning>
  </versioning>
</metadata>
`

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		repositoryURL string
		want          *DefaultHandler
	}{
		{
			name:          "Normal case",
			repositoryURL: "https://somewhere",
			want: &DefaultHandler{
				metadataURL: "https://somewhere",
				webClient:   http.DefaultClient,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, New(tt.repositoryURL))
		})
	}
}

func TestDefaultHandler_GetLatestVersion(t *testing.T) {
	tests := []struct {
		name                 string
		metadataURL          string
		mockedHTTPStatusCode int
		mockedHttpError      error
		mockedHttpBody       string
		want                 string
		wantErr              bool
	}{
		{
			name:                 "Normal case with org.eclipse.mylyn.wikitext.wikitext.core on repo.jenkins-ci.org/releases",
			metadataURL:          "https://repo.jenkins-ci.org/releases/org/eclipse/mylyn/wikitext/wikitext.core/maven-metadata.xml",
			mockedHTTPStatusCode: 200,
			mockedHttpBody:       wikitextCoreMavenMetadata,
			want:                 "1.7.4.v20130429",
		},
		{
			name:                 "Case with HTTP/500 error",
			metadataURL:          "https://repo.jenkins-ci.org/releases/org/eclipse/mylyn/wikitext/wikitext.core/maven-metadata.xml",
			mockedHTTPStatusCode: 500,
			mockedHttpBody:       wikitextCoreMavenMetadata,
			wantErr:              true,
		},
		{
			name:                 "Case with TCP error",
			metadataURL:          "https://repo.jenkins-ci.org/releases/org/eclipse/mylyn/wikitext/wikitext.core/maven-metadata.xml",
			mockedHTTPStatusCode: 200,
			mockedHttpBody:       wikitextCoreMavenMetadata,
			mockedHttpError:      fmt.Errorf("TCP I/O connection timeout"),
			wantErr:              true,
		},
		{
			name:                 "Case with invalid XML error",
			metadataURL:          "https://repo.jenkins-ci.org/releases/org/eclipse/mylyn/wikitext/wikitext.core/maven-metadata.xml",
			mockedHTTPStatusCode: 200,
			mockedHttpBody:       invalidXML,
			wantErr:              true,
		},
		{
			name:                 "Case with no latest artifact version",
			metadataURL:          "https://repo.jenkins-ci.org/releases/org/eclipse/mylyn/wikitext/wikitext.core/maven-metadata.xml",
			mockedHTTPStatusCode: 200,
			mockedHttpBody:       noLatestVersionMavenMetadata,
			wantErr:              true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := DefaultHandler{
				metadataURL: tt.metadataURL,
				webClient: &httpclient.MockClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						body := tt.mockedHttpBody
						statusCode := tt.mockedHTTPStatusCode
						return &http.Response{
							StatusCode: statusCode,
							Body:       ioutil.NopCloser(strings.NewReader(body)),
						}, tt.mockedHttpError
					},
				},
			}

			got, err := sut.GetLatestVersion()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDefaultHandler_GetVersions(t *testing.T) {
	tests := []struct {
		name                 string
		metadataURL          string
		mockedHTTPStatusCode int
		mockedHttpError      error
		mockedHttpBody       string
		want                 []string
		wantErr              bool
	}{
		{
			name:                 "Normal case with org.eclipse.mylyn.wikitext.wikitext.core on repo.jenkins-ci.org/releases",
			metadataURL:          "https://repo.jenkins-ci.org/releases/org/eclipse/mylyn/wikitext/wikitext.core/maven-metadata.xml",
			mockedHTTPStatusCode: 200,
			mockedHttpBody:       wikitextCoreMavenMetadata,
			want:                 []string{"1.7.4.v20130429", "1.7.3", "1.7.2", "1.7.1", "1.7.0"},
		},
		{
			name:                 "Error case with an error returned from the handler",
			metadataURL:          "https://repo.jenkins-ci.org/releases/org/eclipse/mylyn/wikitext/wikitext.core/maven-metadata.xml",
			mockedHTTPStatusCode: 500,
			mockedHttpBody:       wikitextCoreMavenMetadata,
			wantErr:              true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := DefaultHandler{
				metadataURL: tt.metadataURL,
				webClient: &httpclient.MockClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						body := tt.mockedHttpBody
						statusCode := tt.mockedHTTPStatusCode
						return &http.Response{
							StatusCode: statusCode,
							Body:       ioutil.NopCloser(strings.NewReader(body)),
						}, tt.mockedHttpError
					},
				},
			}

			got, err := sut.GetVersions()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
