package dockerdigest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestCondition(t *testing.T) {
	TestCases := []struct {
		name           string
		spec           Spec
		sourceOutput   string
		expectedResult result.Condition
	}{
		{
			name: "Test condition with a digest specified via the manifest without architecture",
			spec: Spec{
				Image:  "ghcr.io/updatecli/updatecli",
				Digest: "v0.64.1@sha256:0c7a19c9607b349cb90af92cb0ec98a70074192eb1a3463c3883f10663024ce5",
			},
			expectedResult: result.Condition{
				Result: "SUCCESS",
				Pass:   true,
			},
		},
		{
			name: "Test condition with a digest specified via the manifest for arm64 architecture",
			spec: Spec{
				Image:        "ghcr.io/updatecli/updatecli",
				Architecture: "arm64",
				Digest:       "v0.64.1@sha256:80c8393095062a96e74ac7e23aab35c8b639b890d94f2fe8fbcbbc983993e1de",
			},
			expectedResult: result.Condition{
				Result: "SUCCESS",
				Pass:   true,
			},
		},
		{
			name: "Test condition with a digest specified via the manifest for amd64 architecture",
			spec: Spec{
				Image:        "ghcr.io/updatecli/updatecli",
				Architecture: "amd64",
				Digest:       "v0.64.1@sha256:4eb444331ca649b24fa8905a98cc1bf922111b7b07292b86a30d3977e5e8f56d",
			},
			expectedResult: result.Condition{
				Result: "SUCCESS",
				Pass:   true,
			},
		},
		{
			name:         "Test condition with a digest specified via the source output",
			sourceOutput: "v0.35.0@sha256:6e1833e5240ac52ecf7609623f18ec4536151e0f58b7243b92fd71ecdf3b94df",
			spec: Spec{
				Image:        "ghcr.io/updatecli/updatecli",
				Digest:       "v0.35.0@sha256:6e1833e5240ac52ecf7609623f18ec4536151e0f58b7243b92fd71ecdf3b94df",
				Architecture: "amd64",
			},
			expectedResult: result.Condition{
				Result: "SUCCESS",
				Pass:   true,
			},
		},
		{
			name: "Test condition with a digest specified via the manifest without tag",
			spec: Spec{
				Image:        "ghcr.io/updatecli/updatecli",
				Digest:       "@sha256:6e1833e5240ac52ecf7609623f18ec4536151e0f58b7243b92fd71ecdf3b94df",
				Architecture: "amd64",
			},
			expectedResult: result.Condition{
				Result: "SUCCESS",
				Pass:   true,
			},
		},
		{
			name:         "Test condition with a digest specified via the source output",
			sourceOutput: "v0.35.0@sha256:6e1833e5240ac52ecf7609623f18ec4536151e0f58b7243b92fd71ecdf3b94df",
			spec: Spec{
				Image:        "ghcr.io/updatecli/updatecli",
				Digest:       "@sha256:6e1833e5240ac52ecf7609623f18ec4536151e0f58b7243b92fd71ecdf3b94df",
				Architecture: "amd64",
			},
			expectedResult: result.Condition{
				Result: "SUCCESS",
				Pass:   true,
			},
		},
	}

	for i := range TestCases {
		t.Run(t.Name(), func(t *testing.T) {
			DockerDigest, err := New(TestCases[i].spec)
			require.NoError(t, err)

			got, _, gotErr := DockerDigest.Condition(TestCases[i].sourceOutput, nil)

			require.NoError(t, gotErr)
			assert.Equal(t, TestCases[i].expectedResult.Pass, got)
		})
	}
}
