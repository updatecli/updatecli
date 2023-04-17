package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/golang"
	"github.com/updatecli/updatecli/pkg/plugins/resources/go/gomod"
	goLang "github.com/updatecli/updatecli/pkg/plugins/resources/go/language"
	gomodule "github.com/updatecli/updatecli/pkg/plugins/resources/go/module"
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
			name:    "Golang Version",
			rootDir: "testdata/noModule",
			expectedPipelines: []config.Spec{
				{
					Name: "Update Golang module gopkg.in/yaml.v3",
					Sources: map[string]source.Config{
						"golangModuleVersion": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get latest golang module gopkg.in/yaml.v3 version",
								Kind: "golang/module",
								Spec: gomodule.Spec{
									Module: "gopkg.in/yaml.v3",
									VersionFilter: version.Filter{
										Kind:    "semver",
										Pattern: ">=3.0.1",
									},
								},
							},
						},
					},
					Targets: map[string]target.Config{
						"golangModuleVersion": {
							SourceID: "golangModuleVersion",
							ResourceConfig: resource.ResourceConfig{
								Name: "Update gopkg.in/yaml.v3 Golang module version",
								Kind: "golang/gomod",
								Spec: gomod.Spec{
									File:   "go.mod",
									Module: "gopkg.in/yaml.v3",
								},
							},
						},
						"goModTidy": {
							DisableSourceInput: true,
							ResourceConfig: resource.ResourceConfig{
								Name:      "Run Go mod tidy",
								Kind:      "shell",
								DependsOn: []string{"golangModuleVersion"},
								Spec: shell.Spec{
									Command: "go mod tidy",
									Environments: shell.Environments{
										{Name: "HOME"},
										{Name: "PATH"},
									},
									ChangedIf: shell.SpecChangedIf{
										Kind: "file/checksum",
										Spec: checksum.Spec{
											Files: []string{"go.mod", "go.sum"},
										},
									},
								},
							},
						},
					},
				},
				{
					Name: "Update Golang version",
					Sources: map[string]source.Config{
						"golangVersion": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get latest golang version",
								Kind: "golang",
								Spec: goLang.Spec{
									VersionFilter: version.Filter{
										Kind:    "semver",
										Pattern: ">=1.20.0",
									},
								},
							},
						},
					},
					Targets: map[string]target.Config{
						"golangVersion": {
							SourceID: "golangVersion",
							ResourceConfig: resource.ResourceConfig{
								Name: "Update Go version",
								Kind: "golang/gomod",
								Spec: gomod.Spec{
									File: "go.mod",
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
			resource, err := golang.New(
				golang.Spec{
					RootDir: tt.rootDir,
				}, "", "")
			require.NoError(t, err)

			pipelines, err := resource.DiscoverManifests()
			require.NoError(t, err)

			if len(pipelines) != len(tt.expectedPipelines) {
				t.Logf("%v pipeline detected but expecting %v", len(pipelines), len(tt.expectedPipelines))
				t.Fail()
				return
			}

			for i := range pipelines {
				test.AssertConfigSpecEqualByteArray(t, &tt.expectedPipelines[i], string(pipelines[i]))
			}
		})
	}
}
