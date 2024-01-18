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
		expectedError  bool
	}{
		{
			spec: Spec{

				Image: "ghcr.io/updatecli/updatecli",
				Tag:   "v0.64.1",
			},
			expectedResult: result.Source{
				Information: "v0.64.1@sha256:0c7a19c9607b349cb90af92cb0ec98a70074192eb1a3463c3883f10663024ce5",
			},
		},
		{
			name: "Nominal case with arm64 architecture",
			spec: Spec{
				Image:        "ghcr.io/updatecli/updatecli",
				Tag:          "v0.64.1",
				Architecture: "arm64",
			},
			expectedResult: result.Source{
				Information: "v0.64.1@sha256:80c8393095062a96e74ac7e23aab35c8b639b890d94f2fe8fbcbbc983993e1de",
			},
		},
		{
			name: "Nominal case with amd64 architecture",
			spec: Spec{
				Image:        "ghcr.io/updatecli/updatecli",
				Tag:          "v0.64.1",
				Architecture: "amd64",
			},
			expectedResult: result.Source{
				Information: "v0.64.1@sha256:4eb444331ca649b24fa8905a98cc1bf922111b7b07292b86a30d3977e5e8f56d",
			},
		},
		{
			name: "Failure - No architecture",
			spec: Spec{
				Image:        "ghcr.io/updatecli/updatecli",
				Tag:          "v0.35.0",
				Architecture: "i386",
			},
			expectedError: true,
		},
		{
			name: "Nominal case with hidden tag from digest",
			spec: Spec{
				Image:        "ghcr.io/updatecli/updatecli",
				Tag:          "v0.35.0",
				Architecture: "amd64",
				HideTag:      true,
			},
			expectedResult: result.Source{
				Information: "@sha256:6e1833e5240ac52ecf7609623f18ec4536151e0f58b7243b92fd71ecdf3b94df",
			},
		},
		{
			name: "Get latest digest reusing the tag prefix",
			spec: Spec{
				Image:        "ghcr.io/updatecli/updatecli",
				Tag:          "v0.35.0@sha256:6e1833e5240ac52ecf7609623f18ec4536151e0f58b7243b92fd71ecdf3b94df",
				Architecture: "amd64",
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

			err = DockerDigest.Source("", &gotResult)

			if TestCases[i].expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, TestCases[i].expectedResult.Information, gotResult.Information)
		})
	}
}
