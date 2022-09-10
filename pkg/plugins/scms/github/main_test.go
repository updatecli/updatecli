package github

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/tmp"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/commit"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		spec       Spec
		pipelineID string
		want       Github
		wantErr    bool
	}{
		{
			name:       "Nominal case",
			pipelineID: "12345",
			spec: Spec{
				Branch:     "main",
				Repository: "updatecli",
				Owner:      "updatecli",
				Directory:  "/home/updatecli",
				Username:   "joe",
				Token:      "superSecretTOkenOfJoe",
				URL:        "github.com",
			},
			want: Github{
				HeadBranch: "updatecli_12345",
				Spec: Spec{
					Branch:     "main",
					Repository: "updatecli",
					Owner:      "updatecli",
					Directory:  "/home/updatecli",
					Username:   "joe",
					Token:      "superSecretTOkenOfJoe",
					URL:        "https://github.com",
				},
			},
		},
		{
			name:       "Nominal case with empty directory",
			pipelineID: "12345",
			spec: Spec{
				Branch:     "main",
				Repository: "updatecli",
				Owner:      "updatecli",
				Username:   "joe",
				Token:      "superSecretTOkenOfJoe",
				URL:        "github.com",
			},
			want: Github{
				HeadBranch: "updatecli_12345",
				Spec: Spec{
					Branch:     "main",
					Repository: "updatecli",
					Owner:      "updatecli",
					Username:   "joe",
					Token:      "superSecretTOkenOfJoe",
					URL:        "https://github.com",
					Directory:  path.Join(tmp.Directory, "github", "updatecli", "updatecli"),
				},
			},
		},
		{
			name:       "Nominal case with empty URL",
			pipelineID: "12345",
			spec: Spec{
				Branch:     "main",
				Repository: "updatecli",
				Owner:      "updatecli",
				Username:   "joe",
				Token:      "superSecretTOkenOfJoe",
				Directory:  "/home/updatecli",
			},
			want: Github{
				HeadBranch: "updatecli_12345",
				Spec: Spec{
					Branch:     "main",
					Repository: "updatecli",
					Owner:      "updatecli",
					Username:   "joe",
					Token:      "superSecretTOkenOfJoe",
					URL:        "https://github.com",
					Directory:  "/home/updatecli",
				},
			},
		},
		{
			name:       "Nominal case with empty URL",
			pipelineID: "12345",
			spec: Spec{
				Branch:     "main",
				Repository: "updatecli",
				Owner:      "updatecli",
				Username:   "joe",
				Token:      "superSecretTOkenOfJoe",
				Directory:  "/home/updatecli",
			},
			want: Github{
				HeadBranch: "updatecli_12345",
				Spec: Spec{
					Branch:     "main",
					Repository: "updatecli",
					Owner:      "updatecli",
					Username:   "joe",
					Token:      "superSecretTOkenOfJoe",
					URL:        "https://github.com",
					Directory:  "/home/updatecli",
				},
			},
		},
		{
			name:       "Custom URL",
			pipelineID: "12345",
			spec: Spec{
				Branch:     "main",
				Repository: "updatecli",
				Owner:      "updatecli",
				Username:   "joe",
				Token:      "superSecretTOkenOfJoe",
				Directory:  "/home/updatecli",
				URL:        "github.project.com",
			},
			want: Github{
				HeadBranch: "updatecli_12345",
				Spec: Spec{
					Branch:     "main",
					Repository: "updatecli",
					Owner:      "updatecli",
					Username:   "joe",
					Token:      "superSecretTOkenOfJoe",
					URL:        "https://github.project.com",
					Directory:  "/home/updatecli",
				},
			},
		},
		{
			name:       "Custom http URL",
			pipelineID: "12345",
			spec: Spec{
				Branch:     "main",
				Repository: "updatecli",
				Owner:      "updatecli",
				Username:   "joe",
				Token:      "superSecretTOkenOfJoe",
				Directory:  "/home/updatecli",
				URL:        "http://github.project.com",
			},
			want: Github{
				HeadBranch: "updatecli_12345",
				Spec: Spec{
					Branch:     "main",
					Repository: "updatecli",
					Owner:      "updatecli",
					Username:   "joe",
					Token:      "superSecretTOkenOfJoe",
					URL:        "http://github.project.com",
					Directory:  "/home/updatecli",
				},
			},
		},
		{
			name:       "Custom https URL",
			pipelineID: "12345",
			spec: Spec{
				Branch:     "main",
				Repository: "updatecli",
				Owner:      "updatecli",
				Username:   "joe",
				Token:      "superSecretTOkenOfJoe",
				Directory:  "/home/updatecli",
				URL:        "https://github.project.com",
			},
			want: Github{
				HeadBranch: "updatecli_12345",
				Spec: Spec{
					Branch:     "main",
					Repository: "updatecli",
					Owner:      "updatecli",
					Username:   "joe",
					Token:      "superSecretTOkenOfJoe",
					URL:        "https://github.project.com",
					Directory:  "/home/updatecli",
				},
			},
		},
		{
			name:       "Validation Error (missing token)",
			pipelineID: "12345",
			spec: Spec{
				Branch:     "main",
				Repository: "updatecli",
				Owner:      "updatecli",
				Directory:  "/tmp/updatecli",
				Username:   "joe",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec, tt.pipelineID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.Spec, got.Spec)
			assert.Equal(t, tt.want.HeadBranch, got.HeadBranch)
			assert.NotNil(t, got.client)
		})
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name     string
		spec     Spec
		children interface{}
		want     Spec
		wantErr  bool
	}{
		{
			name: "Passing case with all arguments overridden",
			spec: Spec{
				Branch:     "main",
				Directory:  "/tmp",
				Email:      "foo@foo.bar",
				Owner:      "olblak",
				Repository: "updatecli",
				Token:      "SuperSecret",
				URL:        "git@github.com:olblak/updatecli.git",
				Username:   "olblak",
				User:       "olblak",
				GPG: sign.GPGSpec{
					SigningKey: "mine",
				},
				Force: false,
				CommitMessage: commit.Commit{
					Title: "Bye",
				},
			},
			children: Spec{
				Branch:     "dev",
				Directory:  "/home",
				Email:      "root@localhost",
				Owner:      "obiwan",
				Repository: "jeditemple",
				Token:      "GotABadFeeling",
				URL:        "https://github.com/obiwan/jeditemple.git",
				Username:   "obiwan",
				User:       "obiwan",
				GPG: sign.GPGSpec{
					SigningKey: "theirs",
				},
				Force: true,
				CommitMessage: commit.Commit{
					Title: "Hello There",
				},
			},
			want: Spec{
				Branch:     "dev",
				Directory:  "/home",
				Email:      "root@localhost",
				Owner:      "obiwan",
				Repository: "jeditemple",
				Token:      "GotABadFeeling",
				URL:        "https://github.com/obiwan/jeditemple.git",
				Username:   "obiwan",
				User:       "obiwan",
				GPG: sign.GPGSpec{
					SigningKey: "theirs",
				},
				Force: true,
				CommitMessage: commit.Commit{
					Title: "Hello There",
				},
			},
		},
		{
			name: "Passing case with partial arguments overridden",
			spec: Spec{
				Branch:     "main",
				Directory:  "/tmp",
				Email:      "foo@foo.bar",
				Owner:      "olblak",
				Repository: "updatecli",
				Token:      "SuperSecret",
				URL:        "git@github.com:olblak/updatecli.git",
				Username:   "olblak",
				User:       "olblak",
				GPG: sign.GPGSpec{
					SigningKey: "mine",
				},
				Force: false,
				CommitMessage: commit.Commit{
					Title: "Bye",
				},
			},
			children: Spec{
				Branch: "dev",
			},
			want: Spec{
				Branch:     "dev",
				Directory:  "/tmp",
				Email:      "foo@foo.bar",
				Owner:      "olblak",
				Repository: "updatecli",
				Token:      "SuperSecret",
				URL:        "git@github.com:olblak/updatecli.git",
				Username:   "olblak",
				User:       "olblak",
				GPG: sign.GPGSpec{
					SigningKey: "mine",
				},
				Force: false,
				CommitMessage: commit.Commit{
					Title: "Bye",
				},
			},
		},
		{
			name: "Failing case with incompatible types",
			spec: Spec{
				Branch:     "main",
				Directory:  "/tmp",
				Email:      "foo@foo.bar",
				Owner:      "olblak",
				Repository: "updatecli",
				Token:      "SuperSecret",
				URL:        "git@github.com:olblak/updatecli.git",
				Username:   "olblak",
				User:       "olblak",
				GPG: sign.GPGSpec{
					SigningKey: "mine",
				},
				Force: false,
				CommitMessage: commit.Commit{
					Title: "Bye",
				},
			},
			children: git.Spec{
				Branch: "dev",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.spec.Merge(tt.children)

			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.want, tt.spec)
		})
	}
}

func TestSpec_MergeFromEnv(t *testing.T) {
	tests := []struct {
		name      string
		envPrefix string
		spec      Spec
		mockEnv   map[string]string
		want      Spec
	}{
		{
			name:      "Passing case with empty struct",
			envPrefix: "UPDATECLI_SCM_LOCAL",
			mockEnv: map[string]string{
				"UPDATECLI_SCM_LOCAL_BRANCH":     "main",
				"UPDATECLI_SCM_LOCAL_DIRECTORY":  "/tmp",
				"UPDATECLI_SCM_LOCAL_EMAIL":      "foo@bar.com",
				"UPDATECLI_SCM_LOCAL_OWNER":      "foo",
				"UPDATECLI_SCM_LOCAL_REPOSITORY": "bar",
				"UPDATECLI_SCM_LOCAL_TOKEN":      "secret",
				"UPDATECLI_SCM_LOCAL_URL":        "git@github.com:foo/bar.git",
				"UPDATECLI_SCM_LOCAL_USERNAME":   "userName",
				"UPDATECLI_SCM_LOCAL_USER":       "user",
			},
			want: Spec{
				Branch:     "main",
				Directory:  "/tmp",
				Email:      "foo@bar.com",
				Owner:      "foo",
				Repository: "bar",
				Token:      "secret",
				URL:        "git@github.com:foo/bar.git",
				Username:   "userName",
				User:       "user",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.mockEnv {
				t.Setenv(key, value)
			}
			tt.spec.MergeFromEnv(tt.envPrefix)

			assert.Equal(t, tt.want, tt.spec)
		})
	}
}
