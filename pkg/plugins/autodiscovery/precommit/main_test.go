package precommit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverManifests(t *testing.T) {

	testdata := []struct {
		name              string
		rootDir           string
		expectedPipelines []string
		ignoreRules       MatchingRules
		onlyRules         MatchingRules
	}{
		{
			name:    "Scenario 1",
			rootDir: "testdata/simple",
			expectedPipelines: []string{`name: 'Bump "https://github.com/psf/black" repo version'

sources:
  'gittag':
    name: 'Get "https://github.com/psf/black" repo version'
    kind: 'gittag'
    spec:
      versionfilter:
        kind: 'semver'
        pattern: '>=20.10.0'
      url: 'https://github.com/psf/black'

targets:
  '.pre-commit-config.yaml':
    name: 'deps(precommit): bump "https://github.com/psf/black" repo version to {{ source "gittag" }}'
    kind: yaml
    sourceid: 'gittag'
    spec:
      file: '.pre-commit-config.yaml'
      key: "$.repos[?(@.repo == 'https://github.com/psf/black')].rev"
      engine: 'yamlpath'
`},
		},
		{
			name:    "Scenario 2 - Only",
			rootDir: "testdata/precommit",
			onlyRules: MatchingRules{
				MatchingRule{
					Repos: map[string]string{
						"https://github.com/pre-commit/pre-commit-hooks": "v4.6.0",
					},
				},
			},
			expectedPipelines: []string{`name: 'Bump "https://github.com/pre-commit/pre-commit-hooks" repo version'

sources:
  'gittag':
    name: 'Get "https://github.com/pre-commit/pre-commit-hooks" repo version'
    kind: 'gittag'
    spec:
      versionfilter:
        kind: 'semver'
        pattern: '>=4.6.0'
      url: 'https://github.com/pre-commit/pre-commit-hooks'

targets:
  '.pre-commit-config.yaml':
    name: 'deps(precommit): bump "https://github.com/pre-commit/pre-commit-hooks" repo version to {{ source "gittag" }}'
    kind: yaml
    sourceid: 'gittag'
    spec:
      file: '.pre-commit-config.yaml'
      key: "$.repos[?(@.repo == 'https://github.com/pre-commit/pre-commit-hooks')].rev"
      engine: 'yamlpath'
`},
		},
		{
			name:    "Scenario 3 - Ignore",
			rootDir: "testdata/precommit",
			expectedPipelines: []string{`name: 'Bump "https://github.com/asottile/add-trailing-comma" repo version'

sources:
  'gittag':
    name: 'Get "https://github.com/asottile/add-trailing-comma" repo version'
    kind: 'gittag'
    spec:
      versionfilter:
        kind: 'semver'
        pattern: '>=3.1.0'
      url: 'https://github.com/asottile/add-trailing-comma'

targets:
  '.pre-commit-config.yaml':
    name: 'deps(precommit): bump "https://github.com/asottile/add-trailing-comma" repo version to {{ source "gittag" }}'
    kind: yaml
    sourceid: 'gittag'
    spec:
      file: '.pre-commit-config.yaml'
      key: "$.repos[?(@.repo == 'https://github.com/asottile/add-trailing-comma')].rev"
      engine: 'yamlpath'
`},
			ignoreRules: MatchingRules{
				MatchingRule{
					Repos: map[string]string{
						"https://github.com/pre-commit/pre-commit-hooks": "v4.6.0",
					},
				},
				MatchingRule{
					Repos: map[string]string{
						"https://github.com/asottile/setup-cfg-fmt": "v2.5.0",
					},
				},
				MatchingRule{
					Repos: map[string]string{
						"https://github.com/asottile/reorder-python-imports": "v3.13.0",
					},
				},
				MatchingRule{
					Repos: map[string]string{
						"https://github.com/asottile/pyupgrade": "v3.17.0",
					},
				},
				MatchingRule{
					Repos: map[string]string{
						"https://github.com/hhatto/autopep8": "v2.3.1",
					},
				},
				MatchingRule{
					Repos: map[string]string{
						"https://github.com/PyCQA/flake8": "7.1.1",
					},
				},
				MatchingRule{
					Repos: map[string]string{
						"https://github.com/pre-commit/mirrors-mypy": "1.11.2",
					},
				},
			},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			a, err := New(
				Spec{
					Only:   tt.onlyRules,
					Ignore: tt.ignoreRules,
				}, tt.rootDir, "", "")
			require.NoError(t, err)

			bytesPipelines, err := a.DiscoverManifests()
			require.NoError(t, err)

			assert.Equal(t, len(tt.expectedPipelines), len(bytesPipelines))
			stringPipelines := []string{}
			for i := range bytesPipelines {
				stringPipelines = append(stringPipelines, string(bytesPipelines[i]))
			}

			pipelines := []string{}

			for i := range stringPipelines {
				pipelines = append(pipelines, stringPipelines[i])
				assert.Equal(t, tt.expectedPipelines[i], stringPipelines[i])
			}
		})
	}

}
