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
		credentials       map[string]gitProviderToken
	}{
		{
			name:    "Scenario - GitHub Action with a single workflow file",
			rootDir: "testdata/updatecli",
			credentials: map[string]gitProviderToken{
				"github.com": {
					Kind:  "github",
					Token: "xxx",
				},
			},
			expectedPipelines: []string{`name: 'deps: bump actions/checkout GitHub workflow'

sources:
  release:
    dependson:
      - 'condition#release:and'
    name: 'Get latest GitHub Release for actions/checkout'
    kind: 'githubrelease'
    spec:
      owner: 'actions'
      repository: 'checkout'
      url: 'https://github.com'
      token: 'xxx'
      versionfilter:
        kind: 'semver'
        pattern: '*'

  tag:
    dependson:
      - 'condition#tag:and'
    name: 'Get latest tag for actions/checkout'
    kind: 'gittag'
    spec:
      url: "https://github.com/actions/checkout.git"
      password: 'xxx'
      versionfilter:
        kind: 'semver'
        pattern: '*'

  branch:
    dependson:
      - 'condition#branch:and'
    name: 'Get latest branch for actions/checkout'
    kind: 'gitbranch'
    spec:
      url: "https://github.com/actions/checkout.git"
      password: 'xxx'
      versionfilter:
        kind: 'semver'
        pattern: '*'

conditions:
  release:
    name: 'Check if actions/checkout@v4 is a GitHub release'
    kind: 'githubrelease'
    disablesourceinput: true
    spec:
      owner: 'actions'
      repository: 'checkout'
      url: 'https://github.com'
      token: 'xxx'
      tag: 'v4'

  tag:
    name: 'Check if actions/checkout@v4 is a tag'
    kind: 'gittag'
    disablesourceinput: true
    spec:
      url: "https://github.com/actions/checkout.git"
      password: 'xxx'
      versionfilter:
        kind: 'regex'
        pattern: '^v4$'

  branch:
    name: 'Check if actions/checkout@v4 is a branch'
    kind: 'gitbranch'
    disablesourceinput: true
    spec:
      branch: 'v4'
      url: "https://github.com/actions/checkout.git"
      password: 'xxx'

targets:
  release:
    dependson:
      - 'condition#release:and'
    disableconditions: true
    name: 'deps(github): bump Action release for actions/checkout from v4 to {{ source "release" }}'
    kind: 'yaml'
    sourceid: 'release'
    transformers:
      - addprefix: '"actions/checkout@'
      - addsuffix: '"'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'

  tag:
    dependson:
      - 'condition#tag:and'
    disableconditions: true
    name: 'deps(github): bump Action tag for actions/checkout from v4 to {{ source "tag" }}'
    kind: 'yaml'
    sourceid: 'tag'
    transformers:
      - addprefix: '"actions/checkout@'
      - addsuffix: '"'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'

  branch:
    dependson:
      - 'condition#branch:and'
    disableconditions: true
    name: 'deps(github): bump Action branch for actions/checkout from v4 to {{ source "branch" }}'
    kind: yaml
    sourceid: branch
    transformers:
      - addprefix: '"actions/checkout@'
      - addsuffix: '"'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'
`},
		},
		{
			name:    "Scenario - Gitea Action with a single workflow file",
			rootDir: "testdata/gitea",
			credentials: map[string]gitProviderToken{
				"gitea.com": {
					Kind:  "gitea",
					Token: "xxx",
				},
			},
			expectedPipelines: []string{`name: 'deps: bump https://gitea.com/actions/checkout Gitea workflow'

sources:
  release:
    dependson:
      - 'condition#release:and'
    name: 'Get latest Gitea Release for https://gitea.com/actions/checkout'
    kind: 'gitearelease'
    spec:
      owner: 'actions'
      repository: 'checkout'
      url: 'https://gitea.com'
      token: 'xxx'
      versionfilter:
        kind: 'semver'
        pattern: '*'

  tag:
    dependson:
      - 'condition#tag:and'
    name: 'Get latest tag for https://gitea.com/actions/checkout'
    kind: 'gittag'
    spec:
      url: "https://gitea.com/actions/checkout.git"
      password: 'xxx'
      versionfilter:
        kind: 'semver'
        pattern: '*'

  branch:
    dependson:
      - 'condition#branch:and'
    name: 'Get latest branch for https://gitea.com/actions/checkout'
    kind: 'gitbranch'
    spec:
      url: "https://gitea.com/actions/checkout.git"
      password: 'xxx'
      versionfilter:
        kind: 'semver'
        pattern: '*'

conditions:
  release:
    name: 'Check if https://gitea.com/actions/checkout@v4 is a Gitea release'
    kind: 'gitearelease'
    disablesourceinput: true
    spec:
      owner: 'actions'
      repository: 'checkout'
      url: 'https://gitea.com'
      token: 'xxx'
      tag: 'v4'

  tag:
    name: 'Check if https://gitea.com/actions/checkout@v4 is a tag'
    kind: 'gittag'
    disablesourceinput: true
    spec:
      url: "https://gitea.com/actions/checkout.git"
      password: 'xxx'
      versionfilter:
        kind: 'regex'
        pattern: '^v4$'

  branch:
    name: 'Check if https://gitea.com/actions/checkout@v4 is a branch'
    kind: 'gitbranch'
    disablesourceinput: true
    spec:
      branch: 'v4'
      url: "https://gitea.com/actions/checkout.git"
      password: 'xxx'

targets:
  release:
    dependson:
      - 'condition#release:and'
    disableconditions: true
    name: 'deps(gitea): bump https://gitea.com/actions/checkout from v4 to {{ source "release" }}'
    kind: 'yaml'
    sourceid: 'release'
    transformers:
      - addprefix: '"https://gitea.com/actions/checkout@'
      - addsuffix: '"'
    spec:
      file: '.gitea/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'

  tag:
    dependson:
      - 'condition#tag:and'
    disableconditions: true
    name: 'deps(gitea): bump https://gitea.com/actions/checkout from v4 to {{ source "tag" }}'
    kind: 'yaml'
    sourceid: 'tag'
    transformers:
      - addprefix: '"https://gitea.com/actions/checkout@'
      - addsuffix: '"'
    spec:
      file: '.gitea/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'

  branch:
    dependson:
      - 'condition#branch:and'
    disableconditions: true
    name: 'deps(gitea): bump https://gitea.com/actions/checkout from v4 to {{ source "branch" }}'
    kind: yaml
    sourceid: branch
    transformers:
      - addprefix: '"https://gitea.com/actions/checkout@'
      - addsuffix: '"'
    spec:
      file: '.gitea/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'
`},
		},
		{
			name:    "Scenario - GitHub Action with duplicate_steps",
			rootDir: "testdata/duplicate_steps",
			credentials: map[string]gitProviderToken{
				"github.com": {
					Kind:  "github",
					Token: "xxx",
				},
			},
			expectedPipelines: []string{`name: 'deps: bump tibdex/github-app-token GitHub workflow'

sources:
  release:
    dependson:
      - 'condition#release:and'
    name: 'Get latest GitHub Release for tibdex/github-app-token'
    kind: 'githubrelease'
    spec:
      owner: 'tibdex'
      repository: 'github-app-token'
      url: 'https://github.com'
      token: 'xxx'
      versionfilter:
        kind: 'semver'
        pattern: '*'

  tag:
    dependson:
      - 'condition#tag:and'
    name: 'Get latest tag for tibdex/github-app-token'
    kind: 'gittag'
    spec:
      url: "https://github.com/tibdex/github-app-token.git"
      password: 'xxx'
      versionfilter:
        kind: 'semver'
        pattern: '*'

  branch:
    dependson:
      - 'condition#branch:and'
    name: 'Get latest branch for tibdex/github-app-token'
    kind: 'gitbranch'
    spec:
      url: "https://github.com/tibdex/github-app-token.git"
      password: 'xxx'
      versionfilter:
        kind: 'semver'
        pattern: '*'

conditions:
  release:
    name: 'Check if tibdex/github-app-token@v2.1 is a GitHub release'
    kind: 'githubrelease'
    disablesourceinput: true
    spec:
      owner: 'tibdex'
      repository: 'github-app-token'
      url: 'https://github.com'
      token: 'xxx'
      tag: 'v2.1'

  tag:
    name: 'Check if tibdex/github-app-token@v2.1 is a tag'
    kind: 'gittag'
    disablesourceinput: true
    spec:
      url: "https://github.com/tibdex/github-app-token.git"
      password: 'xxx'
      versionfilter:
        kind: 'regex'
        pattern: '^v2.1$'

  branch:
    name: 'Check if tibdex/github-app-token@v2.1 is a branch'
    kind: 'gitbranch'
    disablesourceinput: true
    spec:
      branch: 'v2.1'
      url: "https://github.com/tibdex/github-app-token.git"
      password: 'xxx'

targets:
  release:
    dependson:
      - 'condition#release:and'
    disableconditions: true
    name: 'deps(github): bump Action release for tibdex/github-app-token from v2.1 to {{ source "release" }}'
    kind: 'yaml'
    sourceid: 'release'
    transformers:
      - addprefix: '"tibdex/github-app-token@'
      - addsuffix: '"'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'

  tag:
    dependson:
      - 'condition#tag:and'
    disableconditions: true
    name: 'deps(github): bump Action tag for tibdex/github-app-token from v2.1 to {{ source "tag" }}'
    kind: 'yaml'
    sourceid: 'tag'
    transformers:
      - addprefix: '"tibdex/github-app-token@'
      - addsuffix: '"'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'

  branch:
    dependson:
      - 'condition#branch:and'
    disableconditions: true
    name: 'deps(github): bump Action branch for tibdex/github-app-token from v2.1 to {{ source "branch" }}'
    kind: yaml
    sourceid: branch
    transformers:
      - addprefix: '"tibdex/github-app-token@'
      - addsuffix: '"'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'
`, `name: 'deps: bump tibdex/github-app-token GitHub workflow'

sources:
  release:
    dependson:
      - 'condition#release:and'
    name: 'Get latest GitHub Release for tibdex/github-app-token'
    kind: 'githubrelease'
    spec:
      owner: 'tibdex'
      repository: 'github-app-token'
      url: 'https://github.com'
      token: 'xxx'
      versionfilter:
        kind: 'semver'
        pattern: '*'

  tag:
    dependson:
      - 'condition#tag:and'
    name: 'Get latest tag for tibdex/github-app-token'
    kind: 'gittag'
    spec:
      url: "https://github.com/tibdex/github-app-token.git"
      password: 'xxx'
      versionfilter:
        kind: 'semver'
        pattern: '*'

  branch:
    dependson:
      - 'condition#branch:and'
    name: 'Get latest branch for tibdex/github-app-token'
    kind: 'gitbranch'
    spec:
      url: "https://github.com/tibdex/github-app-token.git"
      password: 'xxx'
      versionfilter:
        kind: 'semver'
        pattern: '*'

conditions:
  release:
    name: 'Check if tibdex/github-app-token@v2.1 is a GitHub release'
    kind: 'githubrelease'
    disablesourceinput: true
    spec:
      owner: 'tibdex'
      repository: 'github-app-token'
      url: 'https://github.com'
      token: 'xxx'
      tag: 'v2.1'

  tag:
    name: 'Check if tibdex/github-app-token@v2.1 is a tag'
    kind: 'gittag'
    disablesourceinput: true
    spec:
      url: "https://github.com/tibdex/github-app-token.git"
      password: 'xxx'
      versionfilter:
        kind: 'regex'
        pattern: '^v2.1$'

  branch:
    name: 'Check if tibdex/github-app-token@v2.1 is a branch'
    kind: 'gitbranch'
    disablesourceinput: true
    spec:
      branch: 'v2.1'
      url: "https://github.com/tibdex/github-app-token.git"
      password: 'xxx'

targets:
  release:
    dependson:
      - 'condition#release:and'
    disableconditions: true
    name: 'deps(github): bump Action release for tibdex/github-app-token from v2.1 to {{ source "release" }}'
    kind: 'yaml'
    sourceid: 'release'
    transformers:
      - addprefix: '"tibdex/github-app-token@'
      - addsuffix: '"'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[1].uses'

  tag:
    dependson:
      - 'condition#tag:and'
    disableconditions: true
    name: 'deps(github): bump Action tag for tibdex/github-app-token from v2.1 to {{ source "tag" }}'
    kind: 'yaml'
    sourceid: 'tag'
    transformers:
      - addprefix: '"tibdex/github-app-token@'
      - addsuffix: '"'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[1].uses'

  branch:
    dependson:
      - 'condition#branch:and'
    disableconditions: true
    name: 'deps(github): bump Action branch for tibdex/github-app-token from v2.1 to {{ source "branch" }}'
    kind: yaml
    sourceid: branch
    transformers:
      - addprefix: '"tibdex/github-app-token@'
      - addsuffix: '"'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[1].uses'
`, `name: 'deps: bump actions/checkout GitHub workflow'

sources:
  release:
    dependson:
      - 'condition#release:and'
    name: 'Get latest GitHub Release for actions/checkout'
    kind: 'githubrelease'
    spec:
      owner: 'actions'
      repository: 'checkout'
      url: 'https://github.com'
      token: 'xxx'
      versionfilter:
        kind: 'semver'
        pattern: '*'

  tag:
    dependson:
      - 'condition#tag:and'
    name: 'Get latest tag for actions/checkout'
    kind: 'gittag'
    spec:
      url: "https://github.com/actions/checkout.git"
      password: 'xxx'
      versionfilter:
        kind: 'semver'
        pattern: '*'

  branch:
    dependson:
      - 'condition#branch:and'
    name: 'Get latest branch for actions/checkout'
    kind: 'gitbranch'
    spec:
      url: "https://github.com/actions/checkout.git"
      password: 'xxx'
      versionfilter:
        kind: 'semver'
        pattern: '*'

conditions:
  release:
    name: 'Check if actions/checkout@v4 is a GitHub release'
    kind: 'githubrelease'
    disablesourceinput: true
    spec:
      owner: 'actions'
      repository: 'checkout'
      url: 'https://github.com'
      token: 'xxx'
      tag: 'v4'

  tag:
    name: 'Check if actions/checkout@v4 is a tag'
    kind: 'gittag'
    disablesourceinput: true
    spec:
      url: "https://github.com/actions/checkout.git"
      password: 'xxx'
      versionfilter:
        kind: 'regex'
        pattern: '^v4$'

  branch:
    name: 'Check if actions/checkout@v4 is a branch'
    kind: 'gitbranch'
    disablesourceinput: true
    spec:
      branch: 'v4'
      url: "https://github.com/actions/checkout.git"
      password: 'xxx'

targets:
  release:
    dependson:
      - 'condition#release:and'
    disableconditions: true
    name: 'deps(github): bump Action release for actions/checkout from v4 to {{ source "release" }}'
    kind: 'yaml'
    sourceid: 'release'
    transformers:
      - addprefix: '"actions/checkout@'
      - addsuffix: '"'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[2].uses'

  tag:
    dependson:
      - 'condition#tag:and'
    disableconditions: true
    name: 'deps(github): bump Action tag for actions/checkout from v4 to {{ source "tag" }}'
    kind: 'yaml'
    sourceid: 'tag'
    transformers:
      - addprefix: '"actions/checkout@'
      - addsuffix: '"'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[2].uses'

  branch:
    dependson:
      - 'condition#branch:and'
    disableconditions: true
    name: 'deps(github): bump Action branch for actions/checkout from v4 to {{ source "branch" }}'
    kind: yaml
    sourceid: branch
    transformers:
      - addprefix: '"actions/checkout@'
      - addsuffix: '"'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[2].uses'
`, `name: 'deps: bump updatecli/updatecli-action GitHub workflow'

sources:
  release:
    dependson:
      - 'condition#release:and'
    name: 'Get latest GitHub Release for updatecli/updatecli-action'
    kind: 'githubrelease'
    spec:
      owner: 'updatecli'
      repository: 'updatecli-action'
      url: 'https://github.com'
      token: 'xxx'
      versionfilter:
        kind: 'semver'
        pattern: '*'

  tag:
    dependson:
      - 'condition#tag:and'
    name: 'Get latest tag for updatecli/updatecli-action'
    kind: 'gittag'
    spec:
      url: "https://github.com/updatecli/updatecli-action.git"
      password: 'xxx'
      versionfilter:
        kind: 'semver'
        pattern: '*'

  branch:
    dependson:
      - 'condition#branch:and'
    name: 'Get latest branch for updatecli/updatecli-action'
    kind: 'gitbranch'
    spec:
      url: "https://github.com/updatecli/updatecli-action.git"
      password: 'xxx'
      versionfilter:
        kind: 'semver'
        pattern: '*'

conditions:
  release:
    name: 'Check if updatecli/updatecli-action@v2 is a GitHub release'
    kind: 'githubrelease'
    disablesourceinput: true
    spec:
      owner: 'updatecli'
      repository: 'updatecli-action'
      url: 'https://github.com'
      token: 'xxx'
      tag: 'v2'

  tag:
    name: 'Check if updatecli/updatecli-action@v2 is a tag'
    kind: 'gittag'
    disablesourceinput: true
    spec:
      url: "https://github.com/updatecli/updatecli-action.git"
      password: 'xxx'
      versionfilter:
        kind: 'regex'
        pattern: '^v2$'

  branch:
    name: 'Check if updatecli/updatecli-action@v2 is a branch'
    kind: 'gitbranch'
    disablesourceinput: true
    spec:
      branch: 'v2'
      url: "https://github.com/updatecli/updatecli-action.git"
      password: 'xxx'

targets:
  release:
    dependson:
      - 'condition#release:and'
    disableconditions: true
    name: 'deps(github): bump Action release for updatecli/updatecli-action from v2 to {{ source "release" }}'
    kind: 'yaml'
    sourceid: 'release'
    transformers:
      - addprefix: '"updatecli/updatecli-action@'
      - addsuffix: '"'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[3].uses'

  tag:
    dependson:
      - 'condition#tag:and'
    disableconditions: true
    name: 'deps(github): bump Action tag for updatecli/updatecli-action from v2 to {{ source "tag" }}'
    kind: 'yaml'
    sourceid: 'tag'
    transformers:
      - addprefix: '"updatecli/updatecli-action@'
      - addsuffix: '"'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[3].uses'

  branch:
    dependson:
      - 'condition#branch:and'
    disableconditions: true
    name: 'deps(github): bump Action branch for updatecli/updatecli-action from v2 to {{ source "branch" }}'
    kind: yaml
    sourceid: branch
    transformers:
      - addprefix: '"updatecli/updatecli-action@'
      - addsuffix: '"'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[3].uses'
`},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			g, err := New(
				Spec{
					Credentials: tt.credentials,
				}, tt.rootDir, "")

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
