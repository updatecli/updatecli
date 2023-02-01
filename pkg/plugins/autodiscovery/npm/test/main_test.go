package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	NPMAutodiscovery "github.com/updatecli/updatecli/pkg/plugins/autodiscovery/npm"
	"github.com/updatecli/updatecli/pkg/plugins/resources/json"
	"github.com/updatecli/updatecli/pkg/plugins/resources/npm"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell/success/checksum"
	"github.com/updatecli/updatecli/pkg/plugins/utils/test"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestDiscoverManifests(t *testing.T) {

	testdata := []struct {
		name              string
		rootDir           string
		expectedPipelines []config.Spec
	}{
		{
			name:    "Npm lockfile",
			rootDir: "testdata/npmlockfile",
			expectedPipelines: []config.Spec{
				{

					Name: "Bump \"axios\" package version",
					Sources: map[string]source.Config{
						"npm": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get \"axios\" package version",
								Kind: "npm",
								Spec: npm.Spec{
									Name: "axios",
									VersionFilter: version.Filter{
										Kind:    "semver",
										Pattern: "^1.0.0",
									},
								},
							},
						},
					},
					Targets: map[string]target.Config{
						"package-lock.json": {
							DisableSourceInput: true,
							ResourceConfig: resource.ResourceConfig{
								Name: "Bump \"axios\" package version to {{ source \"npm\" }}",
								Kind: "shell",
								Spec: shell.Spec{
									WorkDir: ".",
									Command: "\nnpm install --package-lock-only --dry-run=$DRY_RUN axios@{{ source \"npm\" }}",
									Environments: shell.Environments{
										{Name: "PATH"},
									},
									ChangedIf: shell.SpecChangedIf{
										Kind: "file/checksum",
										Spec: checksum.Spec{
											Files: []string{
												"package-lock.json",
												"package.json",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:    "Scenario 1",
			rootDir: "testdata/nolockfile",
			expectedPipelines: []config.Spec{
				{

					Name: "Bump \"@mdi/font\" package version",
					Sources: map[string]source.Config{
						"npm": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get \"@mdi/font\" package version",
								Kind: "npm",
								Spec: npm.Spec{
									Name: "@mdi/font",
									VersionFilter: version.Filter{
										Kind:    "semver",
										Pattern: ">=5.9.55",
									},
								},
							},
						},
					},
					Targets: map[string]target.Config{
						"npm": {
							SourceID: "npm",
							ResourceConfig: resource.ResourceConfig{
								Name: "Bump \"@mdi/font\" package version to {{ source \"npm\" }}",
								Kind: "json",
								Spec: json.Spec{
									File: "package.json",
									Key:  "dependencies.@mdi/font",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			resource, err := NPMAutodiscovery.New(
				NPMAutodiscovery.Spec{
					RootDir: tt.rootDir,
				}, "", "")
			require.NoError(t, err)

			pipelines, err := resource.DiscoverManifests()
			require.NoError(t, err)

			for i := range pipelines {
				test.AssertConfigSpecEqualByteArray(t, &tt.expectedPipelines[i], string(pipelines[i]))
			}
		})
	}

}
