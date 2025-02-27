package npm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestChangelog(t *testing.T) {
	tests := []struct {
		name           string
		from           string
		to             string
		spec           Spec
		expectedResult *result.Changelogs
	}{
		{
			name: "Test getting changelog from github",
			from: "1.0.0",
			to:   "1.0.0",
			spec: Spec{
				Name: "axios",
			},
			expectedResult: &result.Changelogs{
				{
					Title:       "v1.0.0",
					PublishedAt: "2022-10-04 19:21:51 +0000 UTC",
					URL:         "https://github.com/axios/axios/releases/tag/v1.0.0",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := New(tt.spec)
			require.NoError(t, err)
			gotChangelogs := n.Changelog(tt.from, tt.to)

			assert.Equal(t, len(*tt.expectedResult), len(*gotChangelogs))
			if len(*tt.expectedResult) == len(*gotChangelogs) {
				for i := range *tt.expectedResult {
					assert.Equal(t, (*tt.expectedResult)[i].Title, (*gotChangelogs)[i].Title)
					assert.Equal(t, (*tt.expectedResult)[i].PublishedAt, (*gotChangelogs)[i].PublishedAt)
					assert.Equal(t, (*tt.expectedResult)[i].URL, (*gotChangelogs)[i].URL)
				}
			}
		})
	}
}

func TestFilteredRelease(t *testing.T) {
	tests := []struct {
		input          []string
		name           string
		from           string
		to             string
		expectedResult []string
	}{
		{
			name:           "Test same from to",
			from:           "1.0.0",
			to:             "1.0.0",
			input:          []string{"0.9.0", "", "1.0.0", "1.1.0"},
			expectedResult: []string{"1.0.0"},
		},
		{
			name:           "Test diff of from to",
			from:           "1.0.0",
			to:             "1.1.0",
			input:          []string{"0.9.0", "", "1.0.0", "1.1.0"},
			expectedResult: []string{"1.0.0", "1.1.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFilteredVersion := filterVersions(tt.input, tt.from, tt.to)
			assert.Equal(t, tt.expectedResult, gotFilteredVersion)
		})
	}

}
