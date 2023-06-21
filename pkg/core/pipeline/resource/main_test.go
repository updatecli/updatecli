package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/resources/githubrelease"
	gomodule "github.com/updatecli/updatecli/pkg/plugins/resources/go/module"
)

func TestGenerateID(t *testing.T) {

	data := []struct {
		resourceConfig ResourceConfig
		expectedHash   string
	}{
		{
			resourceConfig: ResourceConfig{
				Name: "Get value",
				Kind: "golang/module",
				Spec: gomodule.Spec{
					Module: "github.com/updatecli/updatecli",
				},
			},
			expectedHash: "grAgGpDaKH02T6BfoABJboqbD485QhZWsRTOrcg86nw=",
		},
		{
			resourceConfig: ResourceConfig{
				Name: "Get value",
				Kind: "githubrelease",
				Spec: githubrelease.Spec{
					Owner:      "updatecli",
					Repository: "updatecli",
					Token:      "mysecretToken",
				},
			},
			expectedHash: "9fdXCMKm8-44ADLg-FnZoi1I6S9_rXHChbrM1DbxiRI=",
		},
		{
			resourceConfig: ResourceConfig{
				Name: "Get value",
				Kind: "githubrelease",
				Spec: githubrelease.Spec{
					Owner:      "updatecli",
					Repository: "updatecli",
					Username:   "myUsername",
					Token:      "mysecretToken",
				},
			},
			expectedHash: "9fdXCMKm8-44ADLg-FnZoi1I6S9_rXHChbrM1DbxiRI=",
		},
	}

	for i := 0; i < 100; i++ {
		for _, d := range data {
			t.Logf("Iteration %v", i)
			r, err := New(d.resourceConfig)
			require.NoError(t, err)

			id, err := GetAtomicID(r)
			require.NoError(t, err)

			assert.Equal(t, d.expectedHash, id)
		}
	}
}
