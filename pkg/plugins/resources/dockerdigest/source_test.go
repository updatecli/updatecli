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
			name: "Nominal case",
			spec: Spec{
				Image: "ghcr.io/updatecli/updatecli",
				Tag:   "v0.35.0",
			},
			expectedResult: result.Source{
				Information: "v0.35.0@sha256:6e1833e5240ac52ecf7609623f18ec4536151e0f58b7243b92fd71ecdf3b94df",
			},
		},
		{
			name: "Nominal case with architecture",
			spec: Spec{
				Image:        "ghcr.io/updatecli/updatecli",
				Tag:          "v0.35.0",
				Architecture: "arm64",
			},
			expectedResult: result.Source{
				Information: "v0.35.0@sha256:5fa1ad470832ab88c5e94f942fe15b3298fa7f54a660dbe023b937fec1ad2128",
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
				Image:   "ghcr.io/updatecli/updatecli",
				Tag:     "v0.35.0",
				HideTag: true,
			},
			expectedResult: result.Source{
				Information: "@sha256:6e1833e5240ac52ecf7609623f18ec4536151e0f58b7243b92fd71ecdf3b94df",
			},
		},
		{
			name: "Get latest digest reusing the tag prefix",
			spec: Spec{
				Image: "ghcr.io/updatecli/updatecli",
				Tag:   "v0.35.0@sha256:6e1833e5240ac52ecf7609623f18ec4536151e0f58b7243b92fd71ecdf3b94df",
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
