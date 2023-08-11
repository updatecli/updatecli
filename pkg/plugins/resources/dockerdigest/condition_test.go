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
			name: "Test condition with a digest specified via the manifest",
			spec: Spec{
				Image:  "ghcr.io/updatecli/updatecli",
				Digest: "v0.35.0@sha256:6e1833e5240ac52ecf7609623f18ec4536151e0f58b7243b92fd71ecdf3b94df",
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
				Image:  "ghcr.io/updatecli/updatecli",
				Digest: "v0.35.0@sha256:6e1833e5240ac52ecf7609623f18ec4536151e0f58b7243b92fd71ecdf3b94df",
			},
			expectedResult: result.Condition{
				Result: "SUCCESS",
				Pass:   true,
			},
		},
		{
			name: "Test condition with a digest specified via the manifest without tag",
			spec: Spec{
				Image:  "ghcr.io/updatecli/updatecli",
				Digest: "@sha256:6e1833e5240ac52ecf7609623f18ec4536151e0f58b7243b92fd71ecdf3b94df",
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
				Image:  "ghcr.io/updatecli/updatecli",
				Digest: "@sha256:6e1833e5240ac52ecf7609623f18ec4536151e0f58b7243b92fd71ecdf3b94df",
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

			gotResult := result.Condition{}

			err = DockerDigest.Condition(TestCases[i].sourceOutput, nil, &gotResult)
			require.NoError(t, err)

			assert.Equal(t, TestCases[i].expectedResult.Pass, gotResult.Pass)
		})
	}
}
