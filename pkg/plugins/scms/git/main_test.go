package git

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/commit"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
)

type DataSet []Data

type Data struct {
	s                     Spec
	expectedDirectoryName string
}

var (
	Dataset DataSet = DataSet{
		{
			s: Spec{
				URL:      "https://github.com/updatecli/updatecli.git",
				Username: "git",
				Password: "password",
			}, expectedDirectoryName: "github_com_updatecli_updatecli_git",
		},
		{
			s: Spec{
				URL:      "https://github.com/updatecli/updatecli.git",
				Username: "git",
				Password: "password",
			}, expectedDirectoryName: "github_com_updatecli_updatecli_git",
		},
		{
			s: Spec{
				URL:      "https://github.com/updatecli/updatecli.git",
				Password: "password",
			}, expectedDirectoryName: "github_com_updatecli_updatecli_git",
		},
		{
			s: Spec{
				URL:      "https://github.com/updatecli/updatecli.git",
				Username: "git",
			}, expectedDirectoryName: "github_com_updatecli_updatecli_git",
		},
		{
			s: Spec{
				URL:      "https://@github.com/updatecli/updatecli.git",
				Username: "git",
			}, expectedDirectoryName: "github_com_updatecli_updatecli_git",
		},
		{
			s: Spec{
				URL:      "ssh://server:project.git",
				Username: "bob",
			}, expectedDirectoryName: "server_project_git",
		},
	}
)

func TestSanitizeDirectoryName(t *testing.T) {
	for _, data := range Dataset {
		got := sanitizeDirectoryName(data.s.URL)

		if strings.Compare(got, data.expectedDirectoryName) != 0 {
			t.Errorf("got sanitize directory name %q, expected %q", got, data.expectedDirectoryName)
		}
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
				Branch:    "main",
				Directory: "/tmp",
				Email:     "foo@foo.bar",
				URL:       "git@github.com:olblak/updatecli.git",
				Username:  "olblak",
				User:      "olblak",
				GPG: sign.GPGSpec{
					SigningKey: "mine",
				},
				Force: false,
				CommitMessage: commit.Commit{
					Title: "Bye",
				},
			},
			children: Spec{
				Branch:    "dev",
				Directory: "/home",
				Email:     "root@localhost",
				URL:       "https://github.com/obiwan/jeditemple.git",
				Username:  "obiwan",
				User:      "obiwan",
				GPG: sign.GPGSpec{
					SigningKey: "theirs",
				},
				Force: true,
				CommitMessage: commit.Commit{
					Title: "Hello There",
				},
			},
			want: Spec{
				Branch:    "dev",
				Directory: "/home",
				Email:     "root@localhost",
				URL:       "https://github.com/obiwan/jeditemple.git",
				Username:  "obiwan",
				User:      "obiwan",
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
				Branch:    "main",
				Directory: "/tmp",
				Email:     "foo@foo.bar",
				URL:       "git@github.com:olblak/updatecli.git",
				Username:  "olblak",
				User:      "olblak",
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
				Branch:    "dev",
				Directory: "/tmp",
				Email:     "foo@foo.bar",
				URL:       "git@github.com:olblak/updatecli.git",
				Username:  "olblak",
				User:      "olblak",
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
				Branch:    "main",
				Directory: "/tmp",
				Email:     "foo@foo.bar",
				URL:       "git@github.com:olblak/updatecli.git",
				Username:  "olblak",
				User:      "olblak",
				GPG: sign.GPGSpec{
					SigningKey: "mine",
				},
				Force: false,
				CommitMessage: commit.Commit{
					Title: "Bye",
				},
			},
			children: github.Spec{
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
				"UPDATECLI_SCM_LOCAL_BRANCH":    "main",
				"UPDATECLI_SCM_LOCAL_DIRECTORY": "/tmp",
				"UPDATECLI_SCM_LOCAL_EMAIL":     "foo@bar.com",
				"UPDATECLI_SCM_LOCAL_URL":       "git@github.com:foo/bar.git",
				"UPDATECLI_SCM_LOCAL_USERNAME":  "userName",
				"UPDATECLI_SCM_LOCAL_USER":      "user",
			},
			want: Spec{
				Branch:    "main",
				Directory: "/tmp",
				Email:     "foo@bar.com",
				URL:       "git@github.com:foo/bar.git",
				Username:  "userName",
				User:      "user",
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
