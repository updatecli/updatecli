package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/updatecli/updatecli/pkg/core/pipeline/action"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

func TestConfig_EnsureLocalScm(t *testing.T) {
	tests := []struct {
		name         string
		config       *Config
		wantLocalScm scm.Config
		wantErr      error
	}{
		{
			name: "passing case with local SCM disabled",
			config: &Config{
				Spec: Spec{
					SCMs: map[string]scm.Config{
						LOCALSCMIDENTIFIER: {
							Disabled: true,
						},
					},
					Actions: map[string]action.Config{
						"default": {
							ScmID: LOCALSCMIDENTIFIER,
						},
					},
				},
			},
			wantLocalScm: scm.Config{
				Disabled: true,
			},
		},
		{
			name: "passing case with no local SCM (because no target or action reference it)",
			config: &Config{
				Spec: Spec{
					Actions: map[string]action.Config{
						"default": {
							ScmID: "different_than_" + LOCALSCMIDENTIFIER,
						},
					},
				},
			},
			wantLocalScm: scm.Config{},
		},
		{
			name: "passing case with no local SCM specified (autoguess) - SSH origin remote",
			config: &Config{
				gitHandler: gitgeneric.MockGit{
					Remotes: map[string]string{"origin": "git@github.com:olblak/updatecli.git"},
				},
				Spec: Spec{
					Sources: map[string]source.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								SCMID: LOCALSCMIDENTIFIER,
							},
						},
					},
				},
			},
			wantLocalScm: scm.Config{
				Kind: "github",
				Spec: github.Spec{
					Owner:      "olblak",
					Repository: "updatecli",
					Branch:     "main",
				},
			},
		},
		{
			name: "passing case with no local SCM specified (autoguess) - HTTPS origin remote",
			config: &Config{
				gitHandler: gitgeneric.MockGit{
					Remotes: map[string]string{"origin": "https://localhost:2222/olblak/updatecli"},
				},
				Spec: Spec{
					Actions: map[string]action.Config{
						"default": {
							ScmID: LOCALSCMIDENTIFIER,
						},
					},
				},
			},
			wantLocalScm: scm.Config{
				Kind: "git",
				Spec: git.Spec{
					URL:    "https://localhost:2222/olblak/updatecli",
					Branch: "main",
				},
			},
		},
		{
			name: "passing case with a custom local SCM specified merged with the autoguess",
			config: &Config{
				gitHandler: gitgeneric.MockGit{
					Remotes: map[string]string{"origin": "https://localhost:2222/olblak/updatecli.git"},
				},
				Spec: Spec{
					SCMs: map[string]scm.Config{
						LOCALSCMIDENTIFIER: {
							Kind: "git",
							Spec: git.Spec{
								Branch: "production",
							},
						},
					},
					Conditions: map[string]condition.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								SCMID: LOCALSCMIDENTIFIER,
							},
						},
					},
				},
			},
			wantLocalScm: scm.Config{
				Kind: "git",
				Spec: git.Spec{
					URL:    "https://localhost:2222/olblak/updatecli.git",
					Branch: "production",
				},
			},
		},
		{
			name: "failing case with incompatible type beetween autoguess and specified SCM",
			config: &Config{
				gitHandler: gitgeneric.MockGit{
					Remotes: map[string]string{"origin": "https://localhost:2222/olblak/updatecli.git"},
				},
				Spec: Spec{
					SCMs: map[string]scm.Config{
						LOCALSCMIDENTIFIER: {
							Kind: "github",
							Spec: github.Spec{
								Branch: "production",
							},
						},
					},
					Targets: map[string]target.Config{
						"default": {
							ResourceConfig: resource.ResourceConfig{
								SCMID: LOCALSCMIDENTIFIER,
							},
						},
					},
				},
			},
			wantErr: fmt.Errorf("the SCM discovered in the directory \"\" has a different type ('git') than the specified SCM configuration \"local\"."),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotError := tt.config.EnsureLocalScm()

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, gotError)
				return
			}

			assert.Equal(t, tt.wantLocalScm, tt.config.Spec.SCMs[LOCALSCMIDENTIFIER])
		})
	}
}
