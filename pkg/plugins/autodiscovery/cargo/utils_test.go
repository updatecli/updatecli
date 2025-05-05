package cargo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindCargoFiles(t *testing.T) {

	dataset := []struct {
		name               string
		rootDir            string
		expectedFoundFiles []string
	}{
		{
			name:    "Default working scenario",
			rootDir: "testdata",
			expectedFoundFiles: []string{
				"testdata/simple_crate/Cargo.toml",
				"testdata/simple_crate_lock/Cargo.toml",
				"testdata/workspace/Cargo.toml",
				"testdata/workspace_lock/Cargo.toml",
			},
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			foundFiles, err := findCargoFiles(d.rootDir, ValidFiles[:])
			require.NoError(t, err)

			assert.Equal(t, foundFiles, d.expectedFoundFiles)
		})
	}
}

func TestGetCrateMetadata(t *testing.T) {
	testdata := []struct {
		name             string
		rootDir          string
		expectedMetadata crateMetadata
	}{
		{
			name:    "Scenario 1 -- simple_crate",
			rootDir: "testdata/simple_crate",
			expectedMetadata: crateMetadata{
				Name:      "test-crate",
				CargoFile: "testdata/simple_crate/Cargo.toml",
				Dependencies: []crateDependency{
					{Name: "anyhow", Version: "1.0.1", Inlined: true},
					{Name: "rand", Version: "0.8.0"},
				},
				DevDependencies: []crateDependency{
					{Name: "futures", Version: "0.3.21"},
				},
			},
		},
		{
			name:    "Scenario 2 -- simple_crate_lock",
			rootDir: "testdata/simple_crate_lock",
			expectedMetadata: crateMetadata{
				Name:          "test-crate",
				CargoFile:     "testdata/simple_crate_lock/Cargo.toml",
				CargoLockFile: "testdata/simple_crate_lock/Cargo.lock",
				Dependencies: []crateDependency{
					{Name: "anyhow", Version: "1.0.1", Inlined: true},
					{Name: "rand", Version: "0.8.0"},
				},
				DevDependencies: []crateDependency{
					{Name: "futures", Version: "0.3.21"},
				},
			},
		},
		{
			name:    "Scenario 3 -- workspace",
			rootDir: "testdata/workspace",
			expectedMetadata: crateMetadata{
				CargoFile: "testdata/workspace/Cargo.toml",
				Workspace: true,
				WorkspaceDependencies: []crateDependency{
					{Name: "anyhow", Version: "1", Inlined: true},
				},
				WorkspaceMembers: []crateMetadata{
					{Name: "simple_crate",
						CargoFile:        "testdata/workspace/crates/simple_crate/Cargo.toml",
						WorkspaceMembers: []crateMetadata{},
						Dependencies: []crateDependency{
							{Name: "rand", Version: "0.8.0"},
						},
						DevDependencies: []crateDependency{
							{Name: "futures", Version: "0.3.21"},
						}},
				},
			},
		},
		{
			name:    "Scenario 4 -- workspace lock",
			rootDir: "testdata/workspace_lock",
			expectedMetadata: crateMetadata{
				CargoFile:     "testdata/workspace_lock/Cargo.toml",
				CargoLockFile: "testdata/workspace_lock/Cargo.lock",
				Workspace:     true,
				WorkspaceDependencies: []crateDependency{
					{Name: "anyhow", Version: "1", Inlined: true},
				},
				WorkspaceMembers: []crateMetadata{
					{Name: "simple_crate",
						CargoFile:        "testdata/workspace_lock/crates/simple_crate/Cargo.toml",
						CargoLockFile:    "testdata/workspace_lock/Cargo.lock",
						WorkspaceMembers: []crateMetadata{},
						Dependencies: []crateDependency{
							{Name: "rand", Version: "0.8.0"},
						},
						DevDependencies: []crateDependency{
							{Name: "futures", Version: "0.3.21"},
						}},
				},
			},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			metadata, err := getCrateMetadata(tt.rootDir)

			require.NoError(t, err)

			assert.Equal(t, metadata.Name, tt.expectedMetadata.Name)
			assert.Equal(t, metadata.CargoFile, tt.expectedMetadata.CargoFile)
			assert.Equal(t, metadata.CargoLockFile, tt.expectedMetadata.CargoLockFile)
			assert.Equal(t, metadata.Workspace, tt.expectedMetadata.Workspace)
			assert.ElementsMatch(t, metadata.WorkspaceDependencies, tt.expectedMetadata.WorkspaceDependencies)
			assert.ElementsMatch(t, metadata.Dependencies, tt.expectedMetadata.Dependencies)
			assert.ElementsMatch(t, metadata.DevDependencies, tt.expectedMetadata.DevDependencies)
			assert.ElementsMatch(t, metadata.WorkspaceDevDependencies, tt.expectedMetadata.WorkspaceDevDependencies)
			assert.ElementsMatch(t, metadata.WorkspaceMembers, tt.expectedMetadata.WorkspaceMembers)
		})
	}
}
