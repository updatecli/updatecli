package updatecli

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/core/transformer"
	updatecli "github.com/updatecli/updatecli/pkg/plugins/autodiscovery/updatecli"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerdigest"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
	"github.com/updatecli/updatecli/pkg/plugins/resources/yaml"
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
			name:    "Scenario 1",
			rootDir: "testdata",
			expectedPipelines: []config.Spec{
				{

					Name: `deps(updatecli/policy): bump "ghcr.io/updatecli/policies/policies/nodejs/githubaction" Updatecli policy version`,
					Sources: map[string]source.Config{
						"version": {
							ResourceConfig: resource.ResourceConfig{
								Name: `Get latest "ghcr.io/updatecli/policies/policies/nodejs/githubaction" Updatecli policy version`,
								Kind: "dockerimage",
								Spec: dockerimage.Spec{
									Image: "ghcr.io/updatecli/policies/policies/nodejs/githubaction",
									VersionFilter: version.Filter{
										Kind:    "semver",
										Pattern: "*",
									},
								},
							},
						},
						"digest": {
							ResourceConfig: resource.ResourceConfig{
								Name: `Get latest "ghcr.io/updatecli/policies/policies/nodejs/githubaction" Updatecli policy digest`,
								Kind: "dockerdigest",
								DependsOn: []string{
									"version",
								},
								Spec: dockerdigest.Spec{
									Image: "ghcr.io/updatecli/policies/policies/nodejs/githubaction",
									Tag:   `{{ source "version" }}`,
								},
							},
						},
					},
					Targets: map[string]target.Config{
						"compose": {
							SourceID: "digest",
							ResourceConfig: resource.ResourceConfig{
								Name: `deps(updatecli/policy): bump "ghcr.io/updatecli/policies/policies/nodejs/githubaction" Updatecli version policy`,
								Kind: "yaml",
								Spec: yaml.Spec{
									File: "testdata/website/update-compose.yaml",
									Key:  "$.policies[1].policy",
								},
								Transformers: []transformer.Transformer{
									{
										AddPrefix: "ghcr.io/updatecli/policies/policies/nodejs/githubaction:",
									},
								},
							},
						},
					},
				},
				{

					Name: `deps(updatecli/policy): bump "ghcr.io/updatecli/policies/policies/nodejs/netlify" Updatecli policy version`,
					Sources: map[string]source.Config{
						"version": {
							ResourceConfig: resource.ResourceConfig{
								Name: `Get latest "ghcr.io/updatecli/policies/policies/nodejs/netlify" Updatecli policy version`,
								Kind: "dockerimage",
								Spec: dockerimage.Spec{
									Image: "ghcr.io/updatecli/policies/policies/nodejs/netlify",
									VersionFilter: version.Filter{
										Kind:    "semver",
										Pattern: "*",
									},
								},
							},
						},
						"digest": {
							ResourceConfig: resource.ResourceConfig{
								Name: `Get latest "ghcr.io/updatecli/policies/policies/nodejs/netlify" Updatecli policy digest`,
								Kind: "dockerdigest",
								DependsOn: []string{
									"version",
								},
								Spec: dockerdigest.Spec{
									Image: "ghcr.io/updatecli/policies/policies/nodejs/netlify",
									Tag:   `{{ source "version" }}`,
								},
							},
						},
					},
					Targets: map[string]target.Config{
						"compose": {
							SourceID: "digest",
							ResourceConfig: resource.ResourceConfig{
								Name: `deps(updatecli/policy): bump "ghcr.io/updatecli/policies/policies/nodejs/netlify" Updatecli version policy`,
								Kind: "yaml",
								Spec: yaml.Spec{
									File: "testdata/website/update-compose.yaml",
									Key:  "$.policies[2].policy",
								},
								Transformers: []transformer.Transformer{
									{
										AddPrefix: "ghcr.io/updatecli/policies/policies/nodejs/netlify:",
									},
								},
							},
						},
					},
				},
				{

					Name: `deps(updatecli/policy): bump "ghcr.io/updatecli/policies/policies/hugo/netlify" Updatecli policy version`,
					Sources: map[string]source.Config{
						"version": {
							ResourceConfig: resource.ResourceConfig{
								Name: `Get latest "ghcr.io/updatecli/policies/policies/hugo/netlify" Updatecli policy version`,
								Kind: "dockerimage",
								Spec: dockerimage.Spec{
									Image: "ghcr.io/updatecli/policies/policies/hugo/netlify",
									VersionFilter: version.Filter{
										Kind:    "semver",
										Pattern: "*",
									},
								},
							},
						},
						"digest": {
							ResourceConfig: resource.ResourceConfig{
								Name: `Get latest "ghcr.io/updatecli/policies/policies/hugo/netlify" Updatecli policy digest`,
								Kind: "dockerdigest",
								DependsOn: []string{
									"version",
								},
								Spec: dockerdigest.Spec{
									Image: "ghcr.io/updatecli/policies/policies/hugo/netlify",
									Tag:   `{{ source "version" }}`,
								},
							},
						},
					},
					Targets: map[string]target.Config{
						"compose": {
							SourceID: "digest",
							ResourceConfig: resource.ResourceConfig{
								Name: `deps(updatecli/policy): bump "ghcr.io/updatecli/policies/policies/hugo/netlify" Updatecli version policy`,
								Kind: "yaml",
								Spec: yaml.Spec{
									File: "testdata/website/update-compose.yaml",
									Key:  "$.policies[3].policy",
								},
								Transformers: []transformer.Transformer{
									{
										AddPrefix: "ghcr.io/updatecli/policies/policies/hugo/netlify:",
									},
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
			updatecli, err := updatecli.New(
				updatecli.Spec{
					RootDir: tt.rootDir,
					Files:   []string{"update-compose.yaml"},
				}, "", "")
			require.NoError(t, err)

			pipelines, err := updatecli.DiscoverManifests()
			require.NoError(t, err)

			for i := range tt.expectedPipelines {
				test.AssertConfigSpecEqualByteArray(t, &tt.expectedPipelines[i], string(pipelines[i]))
			}

			require.Equal(t, len(tt.expectedPipelines), len(pipelines))
		})
	}
}
