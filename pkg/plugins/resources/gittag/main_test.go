package gittag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		spec    Spec
		want    GitTag
		wantErr bool
	}{
		{
			name: "Nominal case",
			spec: Spec{
				Path: "github.com/updatecli/updatecli",
			},
			want: GitTag{
				spec: Spec{
					Path: "github.com/updatecli/updatecli",
				},
				versionFilter: version.Filter{
					Kind:    "latest",
					Pattern: "latest",
				},
				nativeGitHandler: &gitgeneric.GoGit{},
			},
			wantErr: false,
		},
		{
			name: "Get Hash",
			spec: Spec{
				Path: "github.com/updatecli/updatecli",
				Key:  "hash",
			},
			want: GitTag{
				spec: Spec{
					Path: "github.com/updatecli/updatecli",
					Key:  "hash",
				},
				versionFilter: version.Filter{
					Kind:    "latest",
					Pattern: "latest",
				},
				nativeGitHandler: &gitgeneric.GoGit{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.versionFilter, got.versionFilter)
			assert.Equal(t, tt.want.spec, got.spec)
		})
	}
}
func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		spec    Spec
		want    GitTag
		wantErr bool
	}{
		{
			name: "Nominal case",
			spec: Spec{
				Path: "github.com/updatecli/updatecli",
			},
			want: GitTag{
				spec: Spec{
					Path: "github.com/updatecli/updatecli",
				},
				versionFilter: version.Filter{
					Kind:    "latest",
					Pattern: "latest",
				},
				nativeGitHandler: &gitgeneric.GoGit{},
			},
			wantErr: false,
		},
		{
			name: "Bad Key",
			spec: Spec{
				Path: "github.com/updatecli/updatecli",
				Key:  "commit",
			},
			want: GitTag{
				spec: Spec{},
				versionFilter: version.Filter{
					Kind:    "latest",
					Pattern: "latest",
				},
				nativeGitHandler: &gitgeneric.GoGit{},
			},
			wantErr: true,
		},
		{
			name: "Good Key",
			spec: Spec{
				Path: "github.com/updatecli/updatecli",
				Key:  "hash",
			},
			want: GitTag{
				spec: Spec{},
				versionFilter: version.Filter{
					Kind:    "latest",
					Pattern: "latest",
				},
				nativeGitHandler: &gitgeneric.GoGit{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag, err := New(tt.spec)
			require.NoError(t, err)

			gotErr := tag.Validate()
			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}
