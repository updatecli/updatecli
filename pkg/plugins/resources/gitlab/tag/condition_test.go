package tag

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
			name: "repository olblak/updatecli should exist with tags",
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
			},
			wantResult: false,
		},
		{
			name: "repository should exist with no tag v0.1.11",
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
				Tag:        "v0.1.11",
			},
			wantResult: false,
		},
		{
			name: "repository should exist with tag v0.3.0",
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
				Tag:        "v0.3.0",
			},
			wantResult: true,
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
