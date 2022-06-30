package config

import (
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git"
)

func TestMigrateToGitTmpWorkingBranch(t *testing.T) {
	tests := []struct {
		name       string
		config     Config
		wantResult bool
		wantErr    bool
	}{
		{
			name: "default case",
			config: Config{
				Spec: Spec{
					SCMs: map[string]scm.Config{
						"default": {
							Kind: "git",
						},
					},
				},
			},
			wantResult: true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			c := tt.config

			gotErr := c.migrateToGitTmpWorkingBranch()

			if tt.wantErr {
				require.Error(t, gotErr)
			} else {
				require.NoError(t, gotErr)
			}

			gitSpec := git.Spec{}
			err := mapstructure.Decode(c.Spec.SCMs["default"].Spec, &gitSpec)
			if err != nil {
				t.Error(err)
			}

			assert.Equal(t, tt.wantResult, gitSpec.DisableWorkingBranch)

		})

	}

}
