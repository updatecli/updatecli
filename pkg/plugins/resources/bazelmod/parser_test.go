package bazelmod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseModuleFile(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedDeps  int
		expectedNames []string
		wantErr       bool
	}{
		{
			name: "Parse simple MODULE.bazel",
			content: `module(name = "test_project", version = "1.0.0")

bazel_dep(name = "rules_go", version = "0.42.0")
bazel_dep(name = "gazelle", version = "0.34.0")`,
			expectedDeps:  2,
			expectedNames: []string{"rules_go", "gazelle"},
			wantErr:       false,
		},
		{
			name: "Parse MODULE.bazel with multi-line bazel_dep",
			content: `module(name = "test_project", version = "1.0.0")

bazel_dep(name = "rules_go", version = "0.42.0")
bazel_dep(
    name = "protobuf",
    version = "21.7",
    repo_name = "com_google_protobuf",
)`,
			expectedDeps:  2,
			expectedNames: []string{"rules_go", "protobuf"},
			wantErr:       false,
		},
		{
			name: "Parse MODULE.bazel with comments",
			content: `module(name = "test_project", version = "1.0.0")

# Go rules
bazel_dep(name = "rules_go", version = "0.42.0")
# Protobuf
bazel_dep(name = "protobuf", version = "21.7")`,
			expectedDeps:  2,
			expectedNames: []string{"rules_go", "protobuf"},
			wantErr:       false,
		},
		{
			name:          "Parse empty file",
			content:       `module(name = "test_project", version = "1.0.0")`,
			expectedDeps:  0,
			expectedNames: []string{},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			moduleFile, err := ParseModuleFile(tt.content)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedDeps, len(moduleFile.Deps))

			names := make([]string, len(moduleFile.Deps))
			for i, dep := range moduleFile.Deps {
				names[i] = dep.Name
			}
			assert.Equal(t, tt.expectedNames, names)
		})
	}
}

func TestFindDepByName(t *testing.T) {
	content := `module(name = "test_project", version = "1.0.0")

bazel_dep(name = "rules_go", version = "0.42.0")
bazel_dep(name = "gazelle", version = "0.34.0")`

	moduleFile, err := ParseModuleFile(content)
	require.NoError(t, err)

	tests := []struct {
		name        string
		moduleName  string
		expectedDep *BazelDep
		expectedNil bool
	}{
		{
			name:       "Find existing module",
			moduleName: "rules_go",
			expectedDep: &BazelDep{
				Name:    "rules_go",
				Version: "0.42.0",
			},
			expectedNil: false,
		},
		{
			name:        "Find non-existing module",
			moduleName:  "nonexistent",
			expectedDep: nil,
			expectedNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := moduleFile.FindDepByName(tt.moduleName)
			if tt.expectedNil {
				assert.Nil(t, dep)
			} else {
				require.NotNil(t, dep)
				assert.Equal(t, tt.expectedDep.Name, dep.Name)
				assert.Equal(t, tt.expectedDep.Version, dep.Version)
			}
		})
	}
}

func TestUpdateDepVersion(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		moduleName    string
		newVersion    string
		expectedError bool
		checkVersion  func(t *testing.T, updatedContent string)
	}{
		{
			name: "Update single-line bazel_dep",
			content: `module(name = "test_project", version = "1.0.0")

bazel_dep(name = "rules_go", version = "0.42.0")`,
			moduleName:    "rules_go",
			newVersion:    "0.43.0",
			expectedError: false,
			checkVersion: func(t *testing.T, updatedContent string) {
				assert.Contains(t, updatedContent, `version = "0.43.0"`)
				assert.NotContains(t, updatedContent, `version = "0.42.0"`)
			},
		},
		{
			name: "Update multi-line bazel_dep",
			content: `module(name = "test_project", version = "1.0.0")

bazel_dep(
    name = "protobuf",
    version = "21.7",
    repo_name = "com_google_protobuf",
)`,
			moduleName:    "protobuf",
			newVersion:    "22.0",
			expectedError: false,
			checkVersion: func(t *testing.T, updatedContent string) {
				assert.Contains(t, updatedContent, `version = "22.0"`)
				assert.NotContains(t, updatedContent, `version = "21.7"`)
				assert.Contains(t, updatedContent, `repo_name = "com_google_protobuf"`)
			},
		},
		{
			name: "Update non-existing module",
			content: `module(name = "test_project", version = "1.0.0")

bazel_dep(name = "rules_go", version = "0.42.0")`,
			moduleName:    "nonexistent",
			newVersion:    "1.0.0",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			moduleFile, err := ParseModuleFile(tt.content)
			require.NoError(t, err)

			updatedContent, err := moduleFile.UpdateDepVersion(tt.moduleName, tt.newVersion)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.checkVersion != nil {
				tt.checkVersion(t, updatedContent)
			}
		})
	}
}
