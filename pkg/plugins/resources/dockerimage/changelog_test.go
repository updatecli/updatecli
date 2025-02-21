package dockerimage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestChangelog(t *testing.T) {

	testdata := []struct {
		name              string
		image             string
		version           string
		expectedChangelog result.Changelogs
	}{
		//{
		//	// We can't test this scenario from pullrequest as we don't have access to Dockerhub credentials
		//	name:              "Get changelog from a docker image without changelog labels",
		//	image:             "updatecli/updatecli",
		//	version:           "v0.80.0",
		//	expectedChangelog: nil,
		//},
		//{
		//	// We can't test this scenario from pullrequest as we don't have access to Dockerhub credentials
		//	name:    "Get changelog from an Updatecli policy stored on Dockerhub with changelog labels",
		//	image:   "olblak/updatecli-docusaurus",
		//	version: "0.1.0",
		//	expectedChangelog: result.Changelogs{
		//		{
		//			Title: "0.1.0",
		//			Body:  "Init release",
		//		},
		//	},
		//},
		{
			name:              "Get changelog from an Updatecli policy without labels defined",
			image:             "ghcr.io/updatecli/policies/updatecli/autodiscovery",
			version:           "0.2.0",
			expectedChangelog: nil,
		},
		//{
		//	name:    "Get changelog from an Updatecli policy with the right label defined",
		//	image:   "ghcr.io/olblak/policies/updatecli/autodiscovery",
		//	version: "0.3.0",
		//	expectedChangelog: result.Changelogs{
		//		{
		//			Title: "0.3.0",
		//			Body:  "- Allow to set commit with GitHub GraphQL API using `scm.commitusingapi`",
		//		},
		//	},
		//},
		{
			name:              "Get changelog from an Updatecli policy with the label set to an url without markdown extension",
			image:             "ghcr.io/olblak/policies/updatecli/autodiscovery",
			version:           "0.3.2",
			expectedChangelog: nil,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {

			di, err := New(Spec{
				Image: tt.image,
			})
			require.NoError(t, err)

			di.foundVersion.OriginalVersion = tt.version
			di.foundVersion.ParsedVersion = tt.version

			gotChangelog := di.Changelog("", "")

			assert.Equal(t, tt.expectedChangelog, gotChangelog)
		})
	}

}
