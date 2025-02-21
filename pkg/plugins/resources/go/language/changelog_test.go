package language

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
		expectedResult *result.Changelogs
	}{
		{
			name: "Test new minor version",
			from: "1.20",
			to:   "1.20",
			expectedResult: &result.Changelogs{
				{
					Title: "1.20",
					Body:  "Golang changelog for version \"1.20\" is available on \"https://go.dev/doc/go1.20\"",
					URL:   "https://go.dev/doc/go1.20",
				},
			},
		},
		{
			name: "Test new patch version",
			from: "1.20.1",
			to:   "1.20.1",
			expectedResult: &result.Changelogs{
				{
					Title: "1.20.1",
					Body:  "Golang changelog for version \"1.20.1\" is available on \"https://go.dev/doc/devel/release#go1.20.minor\"",
					URL:   "https://go.dev/doc/devel/release#go1.20.minor",
				},
			},
		},
		{
			name:           "Test without intput",
			expectedResult: nil,
		},
	}

	language, err := New(Spec{})
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := language.Changelog(tt.from, tt.to)
			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}
