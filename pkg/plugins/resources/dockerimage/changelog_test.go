package dockerimage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChangelog(t *testing.T) {

	testdata := []struct {
		image             string
		version           string
		expectedChangelog string
	}{
		{
			image:             "updatecli/updatecli",
			version:           "v0.80.0",
			expectedChangelog: "",
		},
		{
			image:             "ghcr.io/updatecli/policies/updatecli/autodiscovery",
			version:           "0.2.0",
			expectedChangelog: "",
		},
	}

	for _, tt := range testdata {
		t.Run(tt.image, func(t *testing.T) {
			di, err := New(Spec{
				Image: tt.image,
			})

			require.NoError(t, err)

			di.foundVersion.OriginalVersion = tt.version
			di.foundVersion.ParsedVersion = tt.version

			gotChangelog := di.Changelog()

			assert.Equal(t, tt.expectedChangelog, gotChangelog)

			t.Fail()

		})
	}

}
