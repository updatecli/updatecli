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
					Name: "Update Golang version",
					Sources: map[string]source.Config{
						"golangVersion": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get latest golang version",
								Kind: "golang",
								Spec: goLang.Spec{
									VersionFilter: version.Filter{
										Kind:    "semver",
										Pattern: "*",
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
