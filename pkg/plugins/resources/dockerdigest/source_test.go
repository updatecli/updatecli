package dockerdigest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// TestSource tests the Source method using integration tests
func TestSource(t *testing.T) {
	TestCases := []struct {
		name           string
		spec           Spec
		expectedResult result.Source
	}{
		{
			name: "Nominal case",
			spec: Spec{
				Image: "ghcr.io/updatecli/updatecli",
				Tag:   "v0.35.0",
			},
			expectedResult: result.Source{
				Information: "v0.35.0@sha256:6e1833e5240ac52ecf7609623f18ec4536151e0f58b7243b92fd71ecdf3b94df",
			},
		},
	}

	for i := range TestCases {

		t.Run(t.Name(), func(t *testing.T) {
			DockerDigest, err := New(TestCases[i].spec)
			require.NoError(t, err)

			gotResult := result.Source{}

			DockerDigest.Source("", &gotResult)

			assert.Equal(t, TestCases[i].expectedResult.Information, gotResult.Information)
		})
	}
}
