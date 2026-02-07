package gittag

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

type mockGitHandlerForCondition struct {
	gitgeneric.GitHandler
	tagRefs      []gitgeneric.DatedTag
	tagRefsError error
}

func (m *mockGitHandlerForCondition) TagRefs(workingDir string) (refs []gitgeneric.DatedTag, err error) {
	return m.tagRefs, m.tagRefsError
}

func TestGitTag_Condition(t *testing.T) {
	tests := []struct {
		name                   string
		directory              string
		mockedNativeGitHandler gitgeneric.GitHandler
		spec                   Spec
		source                 string
		wantPass               bool
		wantMessage            string
		wantErr                bool
	}{
		{
			name: "Get tags from a remote https URL, filter with v0.99.3",
			spec: Spec{
				URL: "https://github.com/updatecli-test/updatecli.git",
				Tag: "v0.99.3",
			},
			wantPass:    true,
			wantMessage: "git tag \"v0.99.3\" found",
		},
		{
			name:      "Tag specified in spec, tag exists",
			directory: "/tmp/test-repo",
			mockedNativeGitHandler: &mockGitHandlerForCondition{
				tagRefs: []gitgeneric.DatedTag{
					{
						Name: "v1.0.0",
						Hash: "hash1",
					},
					{
						Name: "v2.0.0",
						Hash: "hash2",
					},
					{
						Name: "v3.0.0",
						Hash: "hash3",
					},
				},
			},
			spec: Spec{
				Tag: "v2.0.0",
			},
			source:      "",
			wantPass:    true,
			wantMessage: "git tag \"v2.0.0\" found",
			wantErr:     false,
		},
		{
			name:      "Tag specified in spec, tag doesn't exist",
			directory: "/tmp/test-repo",
			mockedNativeGitHandler: &mockGitHandlerForCondition{
				tagRefs: []gitgeneric.DatedTag{
					{
						Name: "v1.0.0",
						Hash: "hash1",
					},
					{
						Name: "v2.0.0",
						Hash: "hash2",
					},
					{
						Name: "v3.0.0",
						Hash: "hash3",
					},
				},
			},
			spec: Spec{
				Tag: "v4.0.0",
			},
			source:      "",
			wantPass:    false,
			wantMessage: "no git tag found matching \"v4.0.0\"",
			wantErr:     false,
		},
		{
			name:      "Tag specified in spec, source also provided - should prioritize spec.Tag",
			directory: "/tmp/test-repo",
			mockedNativeGitHandler: &mockGitHandlerForCondition{
				tagRefs: []gitgeneric.DatedTag{
					{
						Name: "v1.0.0",
						Hash: "hash1",
					},
					{
						Name: "v2.0.0",
						Hash: "hash2",
					},
					{
						Name: "v3.0.0",
						Hash: "hash3",
					},
				},
			},
			spec: Spec{
				Tag: "v2.0.0",
			},
			source:      "v1.0.0",
			wantPass:    true,
			wantMessage: "git tag \"v2.0.0\" found",
			wantErr:     false,
		},
		{
			name:      "No tag in spec, source provided - should use source",
			directory: "/tmp/test-repo",
			mockedNativeGitHandler: &mockGitHandlerForCondition{
				tagRefs: []gitgeneric.DatedTag{
					{
						Name: "v1.0.0",
						Hash: "hash1",
					},
					{
						Name: "v2.0.0",
						Hash: "hash2",
					},
					{
						Name: "v3.0.0",
						Hash: "hash3",
					},
				},
			},
			spec:        Spec{},
			source:      "v3.0.0",
			wantPass:    true,
			wantMessage: "git tag \"v3.0.0\" found",
			wantErr:     false,
		},
		{
			name:      "No tag in spec, source provided but tag doesn't exist",
			directory: "/tmp/test-repo",
			mockedNativeGitHandler: &mockGitHandlerForCondition{
				tagRefs: []gitgeneric.DatedTag{
					{
						Name: "v1.0.0",
						Hash: "hash1",
					},
					{
						Name: "v2.0.0",
						Hash: "hash2",
					},
					{
						Name: "v3.0.0",
						Hash: "hash3",
					},
				},
			},
			spec:        Spec{},
			source:      "v4.0.0",
			wantPass:    false,
			wantMessage: "no git tag found matching \"v4.0.0\"",
			wantErr:     false,
		},
		{
			name:      "No tag in spec, no source - should fallback to versionFilter",
			directory: "/tmp/test-repo",
			mockedNativeGitHandler: &mockGitHandlerForCondition{
				tagRefs: []gitgeneric.DatedTag{
					{
						Name: "v1.0.0",
						Hash: "hash1",
					},
					{
						Name: "v2.0.0",
						Hash: "hash2",
					},
					{
						Name: "v3.0.0",
						Hash: "hash3",
					},
				},
			},
			spec: Spec{
				VersionFilter: version.Filter{
					Kind:    "latest",
					Pattern: "latest",
				},
			},
			source:      "",
			wantPass:    true,
			wantMessage: "git tag matching \"latest\" found\n",
			wantErr:     false,
		},
		{
			name:      "No tag in spec, no source, versionFilter pattern doesn't match",
			directory: "/tmp/test-repo",
			mockedNativeGitHandler: &mockGitHandlerForCondition{
				tagRefs: []gitgeneric.DatedTag{
					{
						Name: "v1.0.0",
						Hash: "hash1",
					},
					{
						Name: "v2.0.0",
						Hash: "hash2",
					},
					{
						Name: "v3.0.0",
						Hash: "hash3",
					},
				},
			},
			spec: Spec{
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "~4.0",
				},
			},
			source:      "",
			wantPass:    false,
			wantMessage: "",
			wantErr:     true,
		},
		{
			name:      "Error retrieving tags",
			directory: "/tmp/test-repo",
			mockedNativeGitHandler: &mockGitHandlerForCondition{
				tagRefsError: fmt.Errorf("failed to retrieve tags"),
			},
			spec: Spec{
				Tag: "v1.0.0",
			},
			source:      "",
			wantPass:    false,
			wantMessage: "",
			wantErr:     true,
		},
		{
			name:      "No tags found in repository",
			directory: "/tmp/test-repo",
			mockedNativeGitHandler: &mockGitHandlerForCondition{
				tagRefs: []gitgeneric.DatedTag{},
			},
			spec: Spec{
				Tag: "v1.0.0",
			},
			source:      "",
			wantPass:    false,
			wantMessage: "no tags found",
			wantErr:     false,
		},
		{
			name:      "Empty directory - should return error",
			directory: "",
			mockedNativeGitHandler: &mockGitHandlerForCondition{
				tagRefs: []gitgeneric.DatedTag{
					{
						Name: "v1.0.0",
						Hash: "hash1",
					},
				},
			},
			spec: Spec{
				Tag: "v1.0.0",
			},
			source:      "",
			wantPass:    false,
			wantMessage: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize version filter if not already set
			if tt.spec.VersionFilter.Kind == "" {
				tt.spec.VersionFilter = version.Filter{
					Kind:    "latest",
					Pattern: "latest",
				}
			}
			versionFilter, err := tt.spec.VersionFilter.Init()
			require.NoError(t, err)

			gt := &GitTag{
				spec:             tt.spec,
				versionFilter:    versionFilter,
				nativeGitHandler: tt.mockedNativeGitHandler,
				directory:        tt.directory,
			}

			gotPass, gotMessage, gotErr := gt.Condition(tt.source, nil)

			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantPass, gotPass)
			assert.Equal(t, tt.wantMessage, gotMessage)
		})
	}
}
