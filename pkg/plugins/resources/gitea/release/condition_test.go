package release

import (
	"fmt"
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
		wantResult     bool
		wantErr        bool
		wantErrMessage error
	}{
		{
			name: "repository olblak/updatecli should not exist",
			manifest: struct {
				URL        string
				Token      string
				Owner      string
				Repository string
				Tag        string
			}{
				URL:        "codeberg.org",
				Token:      "",
				Owner:      "updatecli",
				Repository: "updatecli-nonexistent",
			},
			wantResult:     false,
			wantErr:        true,
			wantErrMessage: fmt.Errorf("looking for Gitea release: Not Found"),
		},
		{
			name: "repository olblak/updatecli-mirror should exist but no release",
			manifest: struct {
				URL        string
				Token      string
				Owner      string
				Repository string
				Tag        string
			}{
				URL:        "codeberg.org",
				Token:      "",
				Owner:      "updatecli",
				Repository: "demo-terminal",
			},
			wantResult:     false,
			wantErr:        true,
			wantErrMessage: fmt.Errorf("no Gitea release found"),
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
				URL:        "codeberg.org",
				Token:      "",
				Owner:      "updatecli",
				Repository: "demo-terminal",
				Tag:        "2.0.0",
			},
			wantResult:     false,
			wantErr:        true,
			wantErrMessage: fmt.Errorf("no Gitea release found"),
		},
		{
			name: "repository should exist with release v2.15.0",
			manifest: struct {
				URL        string
				Token      string
				Owner      string
				Repository string
				Tag        string
			}{
				URL:        "codeberg.org",
				Token:      "",
				Owner:      "updatecli",
				Repository: "updatecli-action",
				Tag:        "v2.15.0",
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
				if assert.Error(t, gotErr) {
					assert.Equal(t, gotErr.Error(), tt.wantErrMessage.Error())
				}
			} else {
				require.NoError(t, gotErr)
			}

			assert.Equal(t, tt.wantResult, gotResult.Pass)

		})

	}
}
