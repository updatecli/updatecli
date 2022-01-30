package github

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

var now = time.Now()

func TestChangelog(t *testing.T) {
	tests := []struct {
		name          string
		spec          Spec
		version       version.Version
		mockedQuery   *changelogQuery
		mockedError   error
		wantChangelog string
		wantErr       bool
	}{
		{
			name: "Normal case with changelog returned",
			spec: Spec{
				Owner:      "updatecli",
				Repository: "updatecli",
				Username:   "joe",
				Token:      "SuperSecretToken",
			},
			version: version.Version{
				OriginalVersion: "v0.17.0",
				ParsedVersion:   "0.17.0",
			},
			mockedQuery: &changelogQuery{
				Repository: struct {
					Release queriedRelease "graphql:\"release(tagName: $tagName)\""
				}{
					Release: queriedRelease{
						Url:         "https://github.com/updatecli/updatecli.git",
						Description: "Super Awesome Description 🤩\n\nFor Unit Tests!",
						PublishedAt: now,
					},
				},
			},
			wantChangelog: fmt.Sprintf("\nRelease published on the %v at the url %v\n\n%v",
				now,
				"https://github.com/updatecli/updatecli.git",
				"Super Awesome Description 🤩\n\nFor Unit Tests!",
			),
		},
		{
			name: "Case with no release returned (empty or tag only)",
			spec: Spec{
				Owner:      "updatecli",
				Repository: "updatecli",
				Username:   "joe",
				Token:      "SuperSecretToken",
			},
			version: version.Version{
				OriginalVersion: "v0.17.0",
				ParsedVersion:   "0.17.0",
			},
			mockedQuery: &changelogQuery{
				Repository: struct {
					Release queriedRelease "graphql:\"release(tagName: $tagName)\""
				}{
					Release: queriedRelease{},
				},
			},
			wantChangelog: "No Github Release found for v0.17.0 on https://github.com/updatecli/updatecli",
		},
		{
			name: "Case with error returned from query",
			spec: Spec{
				Owner:      "updatecli",
				Repository: "updatecli",
				Username:   "joe",
				Token:      "SuperSecretToken",
			},
			version: version.Version{
				OriginalVersion: "v0.17.0",
				ParsedVersion:   "0.17.0",
			},
			mockedQuery: &changelogQuery{},
			mockedError: fmt.Errorf("Dummy error from github.com."),
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.mockedQuery)

			sut := Github{
				Spec: tt.spec,
				client: &MockGitHubClient{
					mockedQuery: tt.mockedQuery,
					mockedErr:   tt.mockedError,
				},
			}
			got, err := sut.Changelog(tt.version)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantChangelog, got)
		})
	}
}
