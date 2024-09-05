package mavenmetadata

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
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
			<version>1.6.1</version>
			<version>1.6.0</version>
    </versions>
    <lastUpdated>20130619211401</lastUpdated>
  </versioning>
</metadata>
`

const gradleLombokPluginMetadata = `<?xml version='1.0' encoding='US-ASCII'?>
<metadata>
  <groupId>io.freefair.gradle</groupId>
  <artifactId>lombok-plugin</artifactId>
  <version>8.6</version>
  <versioning>
    <latest>8.6</latest>
    <release>8.6</release>
    <versions>
      <version>8.0.1</version>
      <version>8.1.0</version>
      <version>8.2.0</version>
      <version>8.2.1</version>
      <version>8.2.2</version>
      <version>8.3</version>
      <version>8.4</version>
      <version>8.6</version>
    </versions>
    <lastUpdated>20240215231139</lastUpdated>
  </versioning>
</metadata>`

const invalidXMLEncoding = `<?xml version='1.0' encoding='SOMETHING'?>`

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
		versionFilter version.Filter
		want          *DefaultHandler
	}{
		{
			name:          "Normal case",
			repositoryURL: "https://somewhere",
			want: &DefaultHandler{
				metadataURL: "https://somewhere",
				versionFilter: version.Filter{
					Kind: "latest",
				},
				contentRetriever: &text.Text{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, New(tt.repositoryURL, tt.versionFilter))
		})
	}
}

func TestDefaultHandler_GetLatestVersion(t *testing.T) {
	tests := []struct {
		name                 string
		versionFilter        version.Filter
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
			name:                 "Normal case, with US-ASCII encoding io.freefair.gradle:lombok-plugin on plugins.gradle.org/m2",
			metadataURL:          "https://plugins.gradle.org/m2/io/freefair/gradle/lombok-plugin/maven-metadata.xml",
			mockedHTTPStatusCode: 200,
			mockedHttpBody:       gradleLombokPluginMetadata,
			want:                 "8.6",
		},
		{
			name: "Normal case with org.eclipse.mylyn.wikitext.wikitext.core on repo.jenkins-ci.org/releases using semver filter",
			versionFilter: version.Filter{
				Kind:    "semver",
				Pattern: "1.6.x",
			},
			metadataURL:          "https://repo.jenkins-ci.org/releases/org/eclipse/mylyn/wikitext/wikitext.core/maven-metadata.xml",
			mockedHTTPStatusCode: 200,
			mockedHttpBody:       wikitextCoreMavenMetadata,
			want:                 "1.6.1",
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
		{
			name:                 "Case with invalid XML encoding",
			metadataURL:          "https://repo.jenkins-ci.org/releases/org/eclipse/mylyn/wikitext/wikitext.core/maven-metadata.xml",
			mockedHTTPStatusCode: 200,
			mockedHttpBody:       invalidXMLEncoding,
			wantErr:              true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := New(tt.metadataURL, tt.versionFilter)

			sut.contentRetriever.SetHttpClient(&httpclient.MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					body := tt.mockedHttpBody
					statusCode := tt.mockedHTTPStatusCode
					return &http.Response{
						StatusCode: statusCode,
						Body:       io.NopCloser(strings.NewReader(body)),
					}, tt.mockedHttpError
				},
			})
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
			want:                 []string{"1.7.4.v20130429", "1.7.3", "1.7.2", "1.7.1", "1.7.0", "1.6.1", "1.6.0"},
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
			sut := New(tt.metadataURL, version.Filter{})
			sut.contentRetriever.SetHttpClient(&httpclient.MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					body := tt.mockedHttpBody
					statusCode := tt.mockedHTTPStatusCode
					return &http.Response{
						StatusCode: statusCode,
						Body:       io.NopCloser(strings.NewReader(body)),
					}, tt.mockedHttpError
				},
			})

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
