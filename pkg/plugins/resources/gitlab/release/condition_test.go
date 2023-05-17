package release

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestCondition(t *testing.T) {

	tests := []struct {
		name     string
		manifest struct {
			URL        string
			Token      string
			Owner      string
			Repository string
			Tag        string
		}
		wantResult bool
		wantErr    bool
	}{
		{
			name: "repository olblak/updatecli-nonexistent should not exist",
			manifest: struct {
				URL        string
				Token      string
				Owner      string
				Repository string
				Tag        string
			}{
				URL:        "gitlab.com",
				Token:      "",
				Owner:      "olblak",
				Repository: "updatecli-nonexistent",
			},
			wantResult: false,
			wantErr:    true,
		},
		{
			name: "repository cicd-devroom/FOSDEM22 should exist but no release",
			manifest: struct {
				URL        string
				Token      string
				Owner      string
				Repository string
				Tag        string
			}{
				URL:        "gitlab.com",
				Token:      "",
				Owner:      "cicd-devroom",
				Repository: "FOSDEM22",
			},
			wantResult: false,
		},
		{
			name: "repository should exist with no release 2.0.0",
			manifest: struct {
				URL        string
				Token      string
				Owner      string
				Repository string
				Tag        string
			}{
				URL:        "gitlab.com",
				Token:      "",
				Owner:      "cicd-devroom",
				Repository: "FOSDEM22",
				Tag:        "2.0.0",
			},
			wantResult: false,
		},
		{
			name: "repository should exist with release v0.46.2",
			manifest: struct {
				URL        string
				Token      string
				Owner      string
				Repository string
				Tag        string
			}{
				URL:        "gitlab.com",
				Token:      "",
				Owner:      "olblak",
				Repository: "updatecli",
				Tag:        "v0.46.2",
			},
			wantResult: true,
			wantErr:    false,
		},
		{
			name: "repository should exist with no release v0.0.99",
			manifest: struct {
				URL        string
				Token      string
				Owner      string
				Repository string
				Tag        string
			}{
				URL:        "gitlab.com",
				Token:      "",
				Owner:      "olblak",
				Repository: "updatecli",
				Tag:        "v0.0.99",
			},
			wantResult: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			g, gotErr := New(tt.manifest)
			require.NoError(t, gotErr)

			gotResult := result.Condition{}
			gotErr = g.Condition("", nil, &gotResult)

			if tt.wantErr {
				require.Error(t, gotErr)
			} else {
				require.NoError(t, gotErr)
			}

			assert.Equal(t, tt.wantResult, gotResult.Pass)

		})

	}
}
