package gittag

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

type mockNativeGitHandler struct {
	gitgeneric.GitHandler
	tagRefs      []gitgeneric.DatedTag
	tagRefsError error
}

func (m *mockNativeGitHandler) TagRefs(workingDir string) (refs []gitgeneric.DatedTag, err error) {
	return m.tagRefs, m.tagRefsError
}

func TestGitTag_Source(t *testing.T) {
	tests := []struct {
		name                   string
		workingDir             string
		mockedNativeGitHandler gitgeneric.GitHandler
		versionFilter          version.Filter
		spec                   Spec
		wantValue              string
		wantErr                bool
	}{
		{
			name:       "3 tags found, filter with latest",
			workingDir: "github.com/updatecli/updatecli",
			mockedNativeGitHandler: &mockNativeGitHandler{
				tagRefs: []gitgeneric.DatedTag{
					{
						Name: "1.0.0",
						Hash: "abc123",
					},
					{
						Name: "2.0.0",
						Hash: "def456",
					},
					{
						Name: "3.0.0",
						Hash: "ghi789",
					},
				},
			},
			versionFilter: version.Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			spec:      Spec{},
			wantValue: "3.0.0",
			wantErr:   false,
		},
		{
			name:                   "Error: O tags found, filter with latest",
			workingDir:             "github.com/updatecli/updatecli",
			mockedNativeGitHandler: &mockNativeGitHandler{tagRefs: []gitgeneric.DatedTag{}},
			versionFilter: version.Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			spec:      Spec{},
			wantValue: "",
			wantErr:   true,
		},
		{
			name:       "Error: 3 tags found, filter with semver on 2.1.y",
			workingDir: "github.com/updatecli/updatecli",
			mockedNativeGitHandler: &mockNativeGitHandler{
				tagRefs: []gitgeneric.DatedTag{
					{
						Name: "1.0.0",
						Hash: "abc123",
					},
					{
						Name: "2.0.0",
						Hash: "def456",
					},
					{
						Name: "3.0.0",
						Hash: "ghi789",
					},
				},
			},
			versionFilter: version.Filter{
				Kind:    "semver",
				Pattern: "~2.1",
			},
			spec:      Spec{},
			wantValue: "",
			wantErr:   true,
		},
		{
			name:       "3 tags found, filter with semver on 2.1.y",
			workingDir: "github.com/updatecli/updatecli",
			mockedNativeGitHandler: &mockNativeGitHandler{
				tagRefs: []gitgeneric.DatedTag{
					{
						Name: "1.0.0",
						Hash: "abc123",
					},
					{
						Name: "2.0.0",
						Hash: "def456",
					},
					{
						Name: "2.1.1",
						Hash: "ghi789",
					},
				},
			},
			versionFilter: version.Filter{
				Kind:    "semver",
				Pattern: "~2.1",
			},
			spec:      Spec{},
			wantValue: "2.1.1",
			wantErr:   false,
		},
		{
			name:       "Error: error while retrieving tags",
			workingDir: "github.com/updatecli/updatecli",
			mockedNativeGitHandler: &mockNativeGitHandler{
				tagRefsError: fmt.Errorf("Unexpected error while retrieving tags."),
			},
			versionFilter: version.Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			spec:      Spec{},
			wantValue: "",
			wantErr:   true,
		},
		{
			name:       "3 tags found, filter with semver on 2.1.y, return hash",
			workingDir: "github.com/updatecli/updatecli",
			mockedNativeGitHandler: &mockNativeGitHandler{
				tagRefs: []gitgeneric.DatedTag{
					{
						Name: "1.0.0",
						Hash: "abc123",
					},
					{
						Name: "2.0.0",
						Hash: "def456",
					},
					{
						Name: "2.1.1",
						Hash: "ghi789",
					},
				},
			},
			versionFilter: version.Filter{
				Kind:    "semver",
				Pattern: "~2.1",
			},
			spec: Spec{
				Key:  "hash",
				Path: "github.com/updatecli/updatecli",
			},
			wantValue: "ghi789",
			wantErr:   false,
		},
		{
			name:       "3 tags found, filter with semver on 2.1.y, return name",
			workingDir: "github.com/updatecli/updatecli",
			mockedNativeGitHandler: &mockNativeGitHandler{
				tagRefs: []gitgeneric.DatedTag{
					{
						Name: "1.0.0",
						Hash: "abc123",
					},
					{
						Name: "2.0.0",
						Hash: "def456",
					},
					{
						Name: "2.1.1",
						Hash: "ghi789",
					},
				},
			},
			versionFilter: version.Filter{
				Kind:    "semver",
				Pattern: "~2.1",
			},
			spec: Spec{
				Key:  "name",
				Path: "github.com/updatecli/updatecli",
			},
			wantValue: "2.1.1",
			wantErr:   false,
		},
		{
			name:       "5 tags found, filter with regex on 'gopls/', return last tag's name",
			workingDir: "github.com/updatecli/updatecli",
			mockedNativeGitHandler: &mockNativeGitHandler{
				tagRefs: []gitgeneric.DatedTag{
					{
						Name: "1.0.0",
						Hash: "abc123",
					},
					{
						Name: "2.0.0",
						Hash: "def456",
					},
					{
						Name: "2.1.1",
						Hash: "ghi789",
					},
					{
						Name: "gopls/v2.1.1",
						Hash: "jkl012",
					},
					{
						Name: "gopls/v3.0.0",
						Hash: "mno345",
					},
				},
			},
			versionFilter: version.Filter{
				Kind:    "regex",
				Pattern: "^gopls\\/v.*",
			},
			spec: Spec{
				Key:  "name",
				Path: "github.com/updatecli/updatecli",
			},
			wantValue: "gopls/v3.0.0",
			wantErr:   false,
		},
		{
			name:       "5 tags found, filter with regex on 'gopls/', return last tag's hash",
			workingDir: "github.com/updatecli/updatecli",
			mockedNativeGitHandler: &mockNativeGitHandler{
				tagRefs: []gitgeneric.DatedTag{
					{
						Name: "1.0.0",
						Hash: "abc123",
					},
					{
						Name: "2.0.0",
						Hash: "def456",
					},
					{
						Name: "2.1.1",
						Hash: "ghi789",
					},
					{
						Name: "gopls/v2.1.1",
						Hash: "jkl012",
					},
					{
						Name: "gopls/v3.0.0",
						Hash: "mno345",
					},
				},
			},
			versionFilter: version.Filter{
				Kind:    "regex",
				Pattern: "^gopls\\/v.*",
			},
			spec: Spec{
				Key:  "hash",
				Path: "github.com/updatecli/updatecli",
			},
			wantValue: "mno345",
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// A version filter is required for all test cases
			require.NotNil(t, tt.versionFilter)

			gr := &GitTag{
				nativeGitHandler: tt.mockedNativeGitHandler,
				versionFilter:    tt.versionFilter,
				spec:             tt.spec,
			}

			gotResult := result.Source{}
			err := gr.Source(tt.workingDir, &gotResult)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantValue, gotResult.Information)
		})
	}
}
