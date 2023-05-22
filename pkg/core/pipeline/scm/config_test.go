package scm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name           string
		config         Config
		wantErrMessage string
	}{
		{
			name: "Passing case with a valid Config object (enabled SCM)",
			config: Config{
				Kind: "github",
				Spec: github.Spec{
					Directory: "/tmp",
				},
			},
		},
		{
			name: "Passing case with a valid Config object (disabled SCM)",
			config: Config{
				Disabled: true,
			},
		},
		{
			name: "Failing case with no Kind",
			config: Config{
				Spec: github.Spec{
					Directory: "/tmp",
				},
			},
			wantErrMessage: "wrong scm configuration: missing value for parameter 'kind'",
		},
		{
			name: "Failing case with no spec",
			config: Config{
				Kind: "github",
			},
			wantErrMessage: "wrong scm configuration: missing value for parameter 'value'",
		},
		{
			name: "Failing case with disabled SCM and kind specified",
			config: Config{
				Disabled: true,
				Kind:     "git",
			},
			wantErrMessage: "wrong scm configuration: specified value for 'kind' found while SCM is disabled",
		},
		{
			name: "Failing case with disabled SCM and spec specified",
			config: Config{
				Disabled: true,
				Spec: github.Spec{
					Directory: "/tmp",
				},
			},
			wantErrMessage: "wrong scm configuration: specified value for 'spec' found while SCM is disabled",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.config.Validate()
			if tt.wantErrMessage != "" {
				assert.ErrorContains(t, gotErr, tt.wantErrMessage)
				return
			}

			require.NoError(t, gotErr)
		})
	}
}

func TestAutoGuess(t *testing.T) {
	tests := []struct {
		name            string
		workingDir      string
		configUnderTest Config
		mockRemotes     map[string]string
		mockError       error
		want            Config
		wantErr         bool
	}{
		{
			name:       "Passing case with a github repository (SSH URL)",
			workingDir: "/tmp",
			mockRemotes: map[string]string{
				"origin": "git@github.com:olblak/updatecli.git",
				"fork":   "git@github.com:user/updatecli.git",
			},
			want: Config{
				Kind: "github",
				Spec: github.Spec{
					Directory:  "/tmp",
					Repository: "updatecli",
					Owner:      "olblak",
					Branch:     "main",
				},
			},
		},
		{
			name:       "Passing case with a github repository (HTTPS URL)",
			workingDir: "/tmp",
			mockRemotes: map[string]string{
				"origin": "https://github.com/olblak/updatecli",
				"fork":   "https://github.com/user/updatecli",
			},
			want: Config{
				Kind: "github",
				Spec: github.Spec{
					Directory:  "/tmp",
					Repository: "updatecli",
					Owner:      "olblak",
					Branch:     "main",
				},
			},
		},
		{
			name:       "Passing case with a git repository (SSH URL)",
			workingDir: "/tmp",
			mockRemotes: map[string]string{
				"origin": "git@company.org:9222:olblak/updatecli.git",
				"fork":   "git@company.org:9222:user/updatecli.git",
			},
			want: Config{
				Kind: "git",
				Spec: git.Spec{
					Directory: "/tmp",
					URL:       "git@company.org:9222:olblak/updatecli.git",
					Branch:    "main",
				},
			},
		},
		{
			name:       "Passing case with a git repository (HTTPS URL)",
			workingDir: "/tmp",
			mockRemotes: map[string]string{
				"origin": "https://10.0.2.4:443/olblak/updatecli",
				"fork":   "https://10.0.2.4:443/user/updatecli",
			},
			want: Config{
				Kind: "git",
				Spec: git.Spec{
					Directory: "/tmp",
					URL:       "https://10.0.2.4:443/olblak/updatecli",
					Branch:    "main",
				},
			},
		},
		{
			name:       "Passing case with existing github config  and a github repository (HTTPS URL)",
			workingDir: "/tmp",
			configUnderTest: Config{
				Kind: "github",
				Spec: github.Spec{
					Branch: "production",
				},
			},
			mockRemotes: map[string]string{
				"origin": "https://github.com/olblak/updatecli",
				"fork":   "https://github.com/user/updatecli",
			},
			want: Config{
				Kind: "github",
				Spec: github.Spec{
					Directory:  "/tmp",
					Repository: "updatecli",
					Owner:      "olblak",
					Branch:     "production",
				},
			},
		},
		{
			name:       "Passing case with existing git config and a git repository (HTTPS URL)",
			workingDir: "/tmp",
			configUnderTest: Config{
				Kind: "git",
				Spec: git.Spec{
					Branch: "production",
				},
			},
			mockRemotes: map[string]string{
				"origin": "https://10.0.2.4:443/olblak/updatecli",
				"fork":   "https://10.0.2.4:443/user/updatecli",
			},
			want: Config{
				Kind: "git",
				Spec: git.Spec{
					Directory: "/tmp",
					URL:       "https://10.0.2.4:443/olblak/updatecli",
					Branch:    "production",
				},
			},
		},
		{
			name:       "Failing case with existing github config and a git repository (HTTPS URL)",
			workingDir: "/tmp",
			configUnderTest: Config{
				Kind: "github",
				Spec: github.Spec{
					Branch: "production",
				},
			},
			mockRemotes: map[string]string{
				"origin": "https://10.0.2.4:443/olblak/updatecli",
				"fork":   "https://10.0.2.4:443/user/updatecli",
			},
			wantErr: true,
		},
		{
			name:       "Failing case with existing git config  and a github repository (HTTPS URL)",
			workingDir: "/tmp",
			configUnderTest: Config{
				Kind: "git",
				Spec: git.Spec{
					Branch: "production",
				},
			},
			mockRemotes: map[string]string{
				"origin": "https://github.com/olblak/updatecli",
				"fork":   "https://github.com/user/updatecli",
			},
			wantErr: true,
		},
		{
			name:       "Failing case with an error when retrieving the list of remotes",
			workingDir: "/tmp",
			mockError:  fmt.Errorf("ERROR: unable to retrieve a list of remotes for the git repository '/tmp'."),
			wantErr:    true,
		},
		{
			name:       "Failing case with no remote named `origin`",
			workingDir: "/tmp",
			mockRemotes: map[string]string{
				"fork": "https://10.0.2.4:443/user/updatecli",
			},
			wantErr: true,
		},
		{
			name:       "Failing case with a malformed github repository (SSH URL)",
			workingDir: "/tmp",
			mockRemotes: map[string]string{
				"origin": "git@github.com:updatecli.git",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitHandler := gitgeneric.MockGit{
				Remotes: tt.mockRemotes,
				Err:     tt.mockError,
			}

			gotErr := tt.configUnderTest.AutoGuess("default", tt.workingDir, gitHandler)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.want, tt.configUnderTest)
		})
	}
}
