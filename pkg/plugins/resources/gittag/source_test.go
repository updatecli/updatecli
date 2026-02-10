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

func boolPtr(b bool) *bool {
	return &b
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
			name: "Get latest tags from a remote https URL using lsremote",
			versionFilter: version.Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			mockedNativeGitHandler: &mockNativeGitHandler{
				tagRefs: []gitgeneric.DatedTag{
					{
						Name: "0.99.0",
						Hash: "abc123",
					},
					{
						Name: "0.100.0",
						Hash: "ghi789",
					},
					{
						Name: "0.99.1",
						Hash: "def456",
					},
					{
						Name: "0.101.0",
						Hash: "ghi789",
					},
				},
			},
			spec: Spec{
				URL:      "https://github.com/updatecli-test/updatecli.git",
				LsRemote: boolPtr(true),
			},
			wantValue: "0.101.0",
		},
		{
			name: "Get latest tags from a remote https URL using lsremote",
			versionFilter: version.Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			mockedNativeGitHandler: &mockNativeGitHandler{
				tagRefs: []gitgeneric.DatedTag{
					{
						Name: "0.99.0",
						Hash: "abc123",
					},
					{
						Name: "0.100.0",
						Hash: "ghi789",
					},
					{
						Name: "0.99.1",
						Hash: "def456",
					},
					{
						Name: "0.101.0",
						Hash: "ghi789",
					},
				},
			},
			spec: Spec{
				URL:      "https://github.com/updatecli-test/updatecli.git",
				LsRemote: boolPtr(true),
			},
			wantValue: "0.101.0",
		},
		{
			name: "Get latest tags from a remote https URL, filter with v0.99.3",
			versionFilter: version.Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			mockedNativeGitHandler: &mockNativeGitHandler{
				tagRefs: []gitgeneric.DatedTag{
					{
						Name: "0.99.0",
						Hash: "abc123",
					},
					{
						Name: "0.100.0",
						Hash: "ghi789",
					},
					{
						Name: "0.99.1",
						Hash: "def456",
					},
					{
						Name: "0.101.0",
						Hash: "ghi789",
					},
				},
			},
			spec: Spec{
				URL: "https://github.com/updatecli-test/updatecli.git",
			},
			wantValue: "0.101.0",
		},
		{
			name: "No tag found matching pattern",
			versionFilter: version.Filter{
				Kind:    "semver",
				Pattern: "0.69.x",
			},
			mockedNativeGitHandler: &mockNativeGitHandler{
				tagRefs: []gitgeneric.DatedTag{
					{
						Name: "0.99.0",
						Hash: "abc123",
					},
				},
			},
			spec: Spec{
				URL: "https://github.com/updatecli-test/updatecli.git",
			},
			wantValue: "",
			wantErr:   true,
		},
		{
			name: "Get latest semver tags from a remote https URL, filter with v0.99.3",
			versionFilter: version.Filter{
				Kind:    "semver",
				Pattern: "0.99.x",
			},
			mockedNativeGitHandler: &mockNativeGitHandler{
				tagRefs: []gitgeneric.DatedTag{
					{
						Name: "0.99.0",
						Hash: "abc123",
					},
					{
						Name: "0.100.0",
						Hash: "ghi789",
					},
					{
						Name: "0.99.1",
						Hash: "def456",
					},
					{
						Name: "0.101.0",
						Hash: "ghi789",
					},
				},
			},
			spec: Spec{
				URL: "https://github.com/updatecli-test/updatecli.git",
			},
			wantValue: "0.99.1",
		},
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
