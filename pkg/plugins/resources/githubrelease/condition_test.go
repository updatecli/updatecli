package githubrelease

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

func TestGitHubRelease_Condition_NoRelease(t *testing.T) {
	tests := []struct {
		name       string
		tags       []string
		typeFilter github.ReleaseType
	}{
		{
			name: "no releases or tags",
		},
		{
			name:       "explicit type filter excludes tags without releases",
			tags:       []string{"v1.0.0"},
			typeFilter: github.ReleaseType{Release: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gr := GitHubRelease{
				ghHandler:  &mockGhHandler{tags: tt.tags},
				spec:       Spec{Tag: "v1.0.0", TypeFilter: tt.typeFilter},
				typeFilter: tt.typeFilter,
			}
			gotResult, _, err := gr.Condition(context.Background(), "", nil, pathresolver.Resolver{})
			require.NoError(t, err)
			assert.False(t, gotResult)
		})
	}
}
