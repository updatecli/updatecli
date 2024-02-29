package githubaction

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
		token             string
	}{
		{
			name:    "Scenario - helmrelease Simple",
			rootDir: "testdata/updatecli",
			token:   "xxx",
			expectedPipelines: []string{`name: 'deps: bump actions/checkout GitHub workflow'

scms:
  github-action:
    kind: 'git'
    spec:
      url: "https://github.com/actions/checkout.git"
      password: '{{ requiredEnv "GITHUB_TOKEN" }}'

sources:
  release:
    name: 'Get latest GitHub Release for actions/checkout'
    kind: 'githubrelease'
    spec:
      owner: 'actions'
      repository: 'checkout'
      token: '{{ requiredEnv "GITHUB_TOKEN" }}'
      versionfilter:
        kind: 'semver'
        pattern: '*'

  tag:
    name: 'Get latest tag for actions/checkout'
    kind: 'gittag'
    scmid: 'github-action'
    spec:
      versionfilter:
        kind: 'semver'
        pattern: '*'

  branch:
    name: 'Get latest branch for actions/checkout'
    kind: 'gitbranch'
    scmid: 'github-action'
    spec:
      versionfilter:
        kind: 'semver'
        pattern: '*'

conditions:
  release:
    name: 'Check if actions/checkout@v4 is a GitHub release'
    kind: 'githubrelease'
    disablesourceinput: true
    spec:
      owner: actions
      repository: checkout
      token: '{{ requiredEnv "GITHUB_TOKEN" }}'
      tag: '{{ source "release" }}'

  tag:
    name: 'Check if actions/checkout@v4 is a tag'
    kind: 'gittag'
    scmid: 'github-action'
    disablesourceinput: true
    spec:
      tag: 'v4'

  branch:
    name: 'Check if actions/checkout@v4 is a branch'
    kind: 'gitbranch'
    scmid: 'github-action'
    disablesourceinput: true
    spec:
      branch: 'v4'

targets:
  release:
    name: 'deps(github): bump Action release for actions/checkout from v4 to {{ source "release" }}'
    kind: 'yaml'
    sourceid: 'release'
    conditionids:
      - 'release'
    transformers:
      - addprefix: '"actions/checkout@'
      - addsuffix: '"'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'

  tag:
    name: 'deps(github): bump Action tag for actions/checkout from v4 to {{ source "tag" }}'
    kind: 'yaml'
    sourceid: 'tag'
    conditionids:
      - 'tag'
    transformers:
      - addprefix: '"actions/checkout@'
      - addsuffix: '"'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'

  branch:
    name: 'deps(github): bump Action branch for actions/checkout from v4 to {{ source "branch" }}'
    kind: yaml
    sourceid: branch
    conditionids:
      - branch
    transformers:
      - addprefix: '"actions/checkout@'
      - addsuffix: '"'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'
`},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			g, err := New(
				Spec{
					RootDir: tt.rootDir,
					Token:   tt.token,
				}, "", "")

			require.NoError(t, err)

			var pipelines []string
			rawPipelines, err := g.DiscoverManifests()
			require.NoError(t, err)

			assert.Equal(t, len(tt.expectedPipelines), len(rawPipelines))

			for i := range rawPipelines {
				pipelines = append(pipelines, string(rawPipelines[i]))
				assert.Equal(t, tt.expectedPipelines[i], pipelines[i])
			}
		})
	}
}
