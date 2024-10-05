package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	PrecommitAutodiscovery "github.com/updatecli/updatecli/pkg/plugins/autodiscovery/precommit"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gittag"
	"github.com/updatecli/updatecli/pkg/plugins/resources/yaml"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git"
	"github.com/updatecli/updatecli/pkg/plugins/utils/test"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestDiscoverManifests(t *testing.T) {

	testdata := []struct {
		name              string
		rootDir           string
		expectedPipelines []config.Spec
		ignoreRules       PrecommitAutodiscovery.MatchingRules
		onlyRules         PrecommitAutodiscovery.MatchingRules
	}{
		{
			name:    "Scenario 1",
			rootDir: "testdata/simple",
			expectedPipelines: []config.Spec{
				{
					Name: "Bump \"https://github.com/psf/black\" repo version",
					SCMs: map[string]scm.Config{
						"https://github.com/psf/black": {
							Kind: "git",
							Spec: git.Spec{
								URL: "https://github.com/psf/black",
							},
						},
					},
					Sources: map[string]source.Config{
						"https://github.com/psf/black": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get \"https://github.com/psf/black\" repo version",
								Kind: "gittag",
								Spec: gittag.Spec{
									VersionFilter: version.Filter{
										Kind:    "semver",
										Pattern: ">=20.10.0",
									},
								},
								SCMID: "https://github.com/psf/black",
							},
						},
					},
					Targets: map[string]target.Config{
						".pre-commit-config.yaml": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Bump \"https://github.com/psf/black\" repo version to {{ source \"https://github.com/psf/black\" }}",
								Kind: "yaml",
								Spec: yaml.Spec{
									File:   ".pre-commit-config.yaml",
									Key:    "$.repos[?(@.repo == 'https://github.com/psf/black')].rev",
									Engine: yaml.EngineYamlPath,
								},
							},
							SourceID: "https://github.com/psf/black",
						},
					},
				},
			},
		},
		{
			name:    "Scenario 2 - Only",
			rootDir: "testdata/precommit",
			onlyRules: PrecommitAutodiscovery.MatchingRules{
				PrecommitAutodiscovery.MatchingRule{
					Repos: map[string]string{
						"https://github.com/pre-commit/pre-commit-hooks": "v4.6.0",
					},
				},
			},
			expectedPipelines: []config.Spec{
				{
					Name: "Bump \"https://github.com/pre-commit/pre-commit-hooks\" repo version",
					SCMs: map[string]scm.Config{
						"https://github.com/pre-commit/pre-commit-hooks": {
							Kind: "git",
							Spec: git.Spec{
								URL: "https://github.com/pre-commit/pre-commit-hooks",
							},
						},
					},
					Sources: map[string]source.Config{
						"https://github.com/pre-commit/pre-commit-hooks": {
							ResourceConfig: resource.ResourceConfig{
								Name:  "Get \"https://github.com/pre-commit/pre-commit-hooks\" repo version",
								Kind:  "gittag",
								SCMID: "https://github.com/pre-commit/pre-commit-hooks",
								Spec: gittag.Spec{
									VersionFilter: version.Filter{
										Kind:    "semver",
										Pattern: ">=4.6.0",
									},
								},
							},
						},
					},
					Targets: map[string]target.Config{
						".pre-commit-config.yaml": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Bump \"https://github.com/pre-commit/pre-commit-hooks\" repo version to {{ source \"https://github.com/pre-commit/pre-commit-hooks\" }}",
								Kind: "yaml",
								Spec: yaml.Spec{
									File:   ".pre-commit-config.yaml",
									Key:    "$.repos[?(@.repo == 'https://github.com/pre-commit/pre-commit-hooks')].rev",
									Engine: yaml.EngineYamlPath,
								},
							},
							SourceID: "https://github.com/pre-commit/pre-commit-hooks",
						},
					},
				},
			},
		},
		{
			name:    "Scenario 3 - Ignore",
			rootDir: "testdata/precommit",
			ignoreRules: PrecommitAutodiscovery.MatchingRules{
				PrecommitAutodiscovery.MatchingRule{
					Repos: map[string]string{
						"https://github.com/pre-commit/pre-commit-hooks": "v4.6.0",
					},
				},
				PrecommitAutodiscovery.MatchingRule{
					Repos: map[string]string{
						"https://github.com/asottile/setup-cfg-fmt": "v2.5.0",
					},
				},
				PrecommitAutodiscovery.MatchingRule{
					Repos: map[string]string{
						"https://github.com/asottile/reorder-python-imports": "v3.13.0",
					},
				},
				PrecommitAutodiscovery.MatchingRule{
					Repos: map[string]string{
						"https://github.com/asottile/pyupgrade": "v3.17.0",
					},
				},
				PrecommitAutodiscovery.MatchingRule{
					Repos: map[string]string{
						"https://github.com/hhatto/autopep8": "v2.3.1",
					},
				},
				PrecommitAutodiscovery.MatchingRule{
					Repos: map[string]string{
						"https://github.com/PyCQA/flake8": "7.1.1",
					},
				},
				PrecommitAutodiscovery.MatchingRule{
					Repos: map[string]string{
						"https://github.com/pre-commit/mirrors-mypy": "1.11.2",
					},
				},
			},
			expectedPipelines: []config.Spec{
				{
					Name: "Bump \"https://github.com/asottile/add-trailing-comma\" repo version",
					SCMs: map[string]scm.Config{
						"https://github.com/asottile/add-trailing-comma": {
							Kind: "git",
							Spec: git.Spec{
								URL: "https://github.com/asottile/add-trailing-comma",
							},
						},
					},
					Sources: map[string]source.Config{
						"https://github.com/asottile/add-trailing-comma": {
							ResourceConfig: resource.ResourceConfig{
								Name:  "Get \"https://github.com/asottile/add-trailing-comma\" repo version",
								Kind:  "gittag",
								SCMID: "https://github.com/asottile/add-trailing-comma",
								Spec: gittag.Spec{
									VersionFilter: version.Filter{
										Kind:    "semver",
										Pattern: ">=3.1.0",
									},
								},
							},
						},
					},
					Targets: map[string]target.Config{
						".pre-commit-config.yaml": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Bump \"https://github.com/asottile/add-trailing-comma\" repo version to {{ source \"https://github.com/asottile/add-trailing-comma\" }}",
								Kind: "yaml",
								Spec: yaml.Spec{
									File:   ".pre-commit-config.yaml",
									Key:    "$.repos[?(@.repo == 'https://github.com/asottile/add-trailing-comma')].rev",
									Engine: yaml.EngineYamlPath,
								},
							},
							SourceID: "https://github.com/asottile/add-trailing-comma",
						},
					},
				},
			},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			a, err := PrecommitAutodiscovery.New(
				PrecommitAutodiscovery.Spec{
					Only:   tt.onlyRules,
					Ignore: tt.ignoreRules,
				}, tt.rootDir, "")
			require.NoError(t, err)

			pipelines, err := a.DiscoverManifests()
			require.NoError(t, err)

			if len(pipelines) != len(tt.expectedPipelines) {
				t.Logf("%v pipeline detected but expecting %v", len(pipelines), len(tt.expectedPipelines))
				t.Fail()
				return
			}
			// We sort both the pipelines and the expectedPipelines using the same algorithm
			// to ensure the order is the same as map in Golang are unordered
			test.SortConfigSpecArray(t, tt.expectedPipelines, pipelines)
			for i := range pipelines {
				test.AssertConfigSpecEqualByteArray(t, &tt.expectedPipelines[i], string(pipelines[i]))
			}
		})
	}

}
