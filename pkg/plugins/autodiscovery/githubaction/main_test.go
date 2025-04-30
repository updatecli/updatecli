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
		digest            bool
	}{
		{
			name:    "Scenario - GitHub Action using Docker image",
			rootDir: "testdata/docker",
			expectedPipelines: []string{`name: 'deps: bump Docker image "ghcr.io/updatecli/udash"'

sources:
  image:
    name: 'get latest image tag for "ghcr.io/updatecli/udash"'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/updatecli/udash'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=v0.1.0'
targets:
  workflow:
    name: 'deps: bump Docker image "ghcr.io/updatecli/udash" to {{ source "image" }}'
    kind: 'yaml'
    spec:
      file: '.github/workflows/docker-01.yaml'
      key: '$.jobs.container-updatecli.container.image'
    sourceid: 'image'
    transformers:
      - addprefix: 'ghcr.io/updatecli/udash:'
`,
				`name: 'deps: bump Docker image "ghcr.io/updatecli/updatecli"'

sources:
  image:
    name: 'get latest image tag for "ghcr.io/updatecli/updatecli"'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=v0.67.0'
targets:
  workflow:
    name: 'deps: bump Docker image "ghcr.io/updatecli/updatecli" to {{ source "image" }}'
    kind: 'yaml'
    spec:
      file: '.github/workflows/docker-02.yaml'
      key: '$.jobs.updatecli.steps[0].uses'
    sourceid: 'image'
    transformers:
      - addprefix: 'docker://ghcr.io/updatecli/updatecli:'
`},
		},
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
      - addprefix: 'actions/checkout@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'
      engine: 'yamlpath'

  tag:
    dependson:
      - 'condition#tag:and'
    disableconditions: true
    name: 'deps(github): bump Action tag for actions/checkout from v4 to {{ source "tag" }}'
    kind: 'yaml'
    sourceid: 'tag'
    transformers:
      - addprefix: 'actions/checkout@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'
      engine: 'yamlpath'

  branch:
    dependson:
      - 'condition#branch:and'
    disableconditions: true
    name: 'deps(github): bump Action branch for actions/checkout from v4 to {{ source "branch" }}'
    kind: yaml
    sourceid: branch
    transformers:
      - addprefix: 'actions/checkout@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'
      engine: 'yamlpath'
`},
		},
		{
			name:    "Scenario - GitHub Action with a single workflow file and digest",
			rootDir: "testdata/digest",
			credentials: map[string]gitProviderToken{
				"github.com": {
					Kind:  "github",
					Token: "xxx",
				},
			},
			digest: true,
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

  release_digest:
    dependson:
      - 'condition#release:and'
    name: 'Get latest GitHub Release for actions/checkout'
    kind: 'githubrelease'
    spec:
      owner: 'actions'
      repository: 'checkout'
      url: 'https://github.com'
      token: 'xxx'
      key: 'taghash'
      versionfilter:
        kind: 'regex'
        pattern: '{{ source "release" }}'

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

  tag_digest:
    dependson:
      - 'condition#tag:and'
    name: 'Get latest tag for actions/checkout'
    kind: 'gittag'
    spec:
      url: "https://github.com/actions/checkout.git"
      password: 'xxx'
      key: 'hash'
      versionfilter:
        kind: 'regex'
        pattern: '{{ source "tag" }}'

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

  branch_digest:
    dependson:
      - 'condition#branch:and'
    name: 'Get latest branch for actions/checkout'
    kind: 'gitbranch'
    spec:
      url: "https://github.com/actions/checkout.git"
      password: 'xxx'
      key: 'hash'
      versionfilter:
        kind: 'regex'
        pattern: '{{ source "branch" }}'

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
    name: 'deps(github): bump Action release for actions/checkout from v4 to {{ source "release_digest" }} (Pinned from {{ source "release" }})'
    kind: 'yaml'
    sourceid: 'release_digest'
    transformers:
      - addprefix: 'actions/checkout@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'
      engine: 'yamlpath'
      comment: '{{ source "release" }}'

  tag:
    dependson:
      - 'condition#tag:and'
    disableconditions: true
    name: 'deps(github): bump Action tag for actions/checkout from v4 to {{ source "tag_digest" }} (Pinned from {{ source "tag" }})'
    kind: 'yaml'
    sourceid: 'tag_digest'
    transformers:
      - addprefix: 'actions/checkout@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'
      engine: 'yamlpath'
      comment: '{{ source "tag" }}'

  branch:
    dependson:
      - 'condition#branch:and'
    disableconditions: true
    name: 'deps(github): bump Action branch for actions/checkout from v4 to {{ source "branch_digest" }} (Pinned from {{ source "branch" }})'
    kind: yaml
    sourceid: branch_digest
    transformers:
      - addprefix: 'actions/checkout@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'
      engine: 'yamlpath'
      comment: '{{ source "branch" }}'
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

  release_digest:
    dependson:
      - 'condition#release:and'
    name: 'Get latest GitHub Release for actions/checkout'
    kind: 'githubrelease'
    spec:
      owner: 'actions'
      repository: 'checkout'
      url: 'https://github.com'
      token: 'xxx'
      key: 'taghash'
      versionfilter:
        kind: 'regex'
        pattern: '{{ source "release" }}'

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

  tag_digest:
    dependson:
      - 'condition#tag:and'
    name: 'Get latest tag for actions/checkout'
    kind: 'gittag'
    spec:
      url: "https://github.com/actions/checkout.git"
      password: 'xxx'
      key: 'hash'
      versionfilter:
        kind: 'regex'
        pattern: '{{ source "tag" }}'

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

  branch_digest:
    dependson:
      - 'condition#branch:and'
    name: 'Get latest branch for actions/checkout'
    kind: 'gitbranch'
    spec:
      url: "https://github.com/actions/checkout.git"
      password: 'xxx'
      key: 'hash'
      versionfilter:
        kind: 'regex'
        pattern: '{{ source "branch" }}'

conditions:
  release:
    name: 'Check if actions/checkout@v4.2.2 is a GitHub release'
    kind: 'githubrelease'
    disablesourceinput: true
    spec:
      owner: 'actions'
      repository: 'checkout'
      url: 'https://github.com'
      token: 'xxx'
      tag: 'v4.2.2'

  tag:
    name: 'Check if actions/checkout@v4.2.2 is a tag'
    kind: 'gittag'
    disablesourceinput: true
    spec:
      url: "https://github.com/actions/checkout.git"
      password: 'xxx'
      versionfilter:
        kind: 'regex'
        pattern: '^v4.2.2$'

  branch:
    name: 'Check if actions/checkout@v4.2.2 is a branch'
    kind: 'gitbranch'
    disablesourceinput: true
    spec:
      branch: 'v4.2.2'
      url: "https://github.com/actions/checkout.git"
      password: 'xxx'

targets:
  release:
    dependson:
      - 'condition#release:and'
    disableconditions: true
    name: 'deps(github): bump Action release for actions/checkout from 11bd71901bbe5b1630ceea73d27597364c9af683 to {{ source "release_digest" }} (Pinned from {{ source "release" }})'
    kind: 'yaml'
    sourceid: 'release_digest'
    transformers:
      - addprefix: 'actions/checkout@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[1].uses'
      engine: 'yamlpath'
      comment: '{{ source "release" }}'

  tag:
    dependson:
      - 'condition#tag:and'
    disableconditions: true
    name: 'deps(github): bump Action tag for actions/checkout from 11bd71901bbe5b1630ceea73d27597364c9af683 to {{ source "tag_digest" }} (Pinned from {{ source "tag" }})'
    kind: 'yaml'
    sourceid: 'tag_digest'
    transformers:
      - addprefix: 'actions/checkout@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[1].uses'
      engine: 'yamlpath'
      comment: '{{ source "tag" }}'

  branch:
    dependson:
      - 'condition#branch:and'
    disableconditions: true
    name: 'deps(github): bump Action branch for actions/checkout from 11bd71901bbe5b1630ceea73d27597364c9af683 to {{ source "branch_digest" }} (Pinned from {{ source "branch" }})'
    kind: yaml
    sourceid: branch_digest
    transformers:
      - addprefix: 'actions/checkout@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[1].uses'
      engine: 'yamlpath'
      comment: '{{ source "branch" }}'
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
        kind: 'latest'

  release_digest:
    dependson:
      - 'condition#release:and'
    name: 'Get latest GitHub Release for actions/checkout'
    kind: 'githubrelease'
    spec:
      owner: 'actions'
      repository: 'checkout'
      url: 'https://github.com'
      token: 'xxx'
      key: 'taghash'
      versionfilter:
        kind: 'regex'
        pattern: '{{ source "release" }}'

  tag:
    dependson:
      - 'condition#tag:and'
    name: 'Get latest tag for actions/checkout'
    kind: 'gittag'
    spec:
      url: "https://github.com/actions/checkout.git"
      password: 'xxx'
      versionfilter:
        kind: 'latest'

  tag_digest:
    dependson:
      - 'condition#tag:and'
    name: 'Get latest tag for actions/checkout'
    kind: 'gittag'
    spec:
      url: "https://github.com/actions/checkout.git"
      password: 'xxx'
      key: 'hash'
      versionfilter:
        kind: 'regex'
        pattern: '{{ source "tag" }}'

  branch:
    dependson:
      - 'condition#branch:and'
    name: 'Get latest branch for actions/checkout'
    kind: 'gitbranch'
    spec:
      url: "https://github.com/actions/checkout.git"
      password: 'xxx'
      versionfilter:
        kind: 'latest'

  branch_digest:
    dependson:
      - 'condition#branch:and'
    name: 'Get latest branch for actions/checkout'
    kind: 'gitbranch'
    spec:
      url: "https://github.com/actions/checkout.git"
      password: 'xxx'
      key: 'hash'
      versionfilter:
        kind: 'regex'
        pattern: '{{ source "branch" }}'

conditions:
  release:
    name: 'Check if actions/checkout@main is a GitHub release'
    kind: 'githubrelease'
    disablesourceinput: true
    spec:
      owner: 'actions'
      repository: 'checkout'
      url: 'https://github.com'
      token: 'xxx'
      tag: 'main'

  tag:
    name: 'Check if actions/checkout@main is a tag'
    kind: 'gittag'
    disablesourceinput: true
    spec:
      url: "https://github.com/actions/checkout.git"
      password: 'xxx'
      versionfilter:
        kind: 'regex'
        pattern: '^main$'

  branch:
    name: 'Check if actions/checkout@main is a branch'
    kind: 'gitbranch'
    disablesourceinput: true
    spec:
      branch: 'main'
      url: "https://github.com/actions/checkout.git"
      password: 'xxx'

targets:
  release:
    dependson:
      - 'condition#release:and'
    disableconditions: true
    name: 'deps(github): bump Action release for actions/checkout from main to {{ source "release_digest" }} (Pinned from {{ source "release" }})'
    kind: 'yaml'
    sourceid: 'release_digest'
    transformers:
      - addprefix: 'actions/checkout@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[2].uses'
      engine: 'yamlpath'
      comment: '{{ source "release" }}'

  tag:
    dependson:
      - 'condition#tag:and'
    disableconditions: true
    name: 'deps(github): bump Action tag for actions/checkout from main to {{ source "tag_digest" }} (Pinned from {{ source "tag" }})'
    kind: 'yaml'
    sourceid: 'tag_digest'
    transformers:
      - addprefix: 'actions/checkout@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[2].uses'
      engine: 'yamlpath'
      comment: '{{ source "tag" }}'

  branch:
    dependson:
      - 'condition#branch:and'
    disableconditions: true
    name: 'deps(github): bump Action branch for actions/checkout from main to {{ source "branch_digest" }} (Pinned from {{ source "branch" }})'
    kind: yaml
    sourceid: branch_digest
    transformers:
      - addprefix: 'actions/checkout@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[2].uses'
      engine: 'yamlpath'
      comment: '{{ source "branch" }}'
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
			digest: true,
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
      - addprefix: 'https://gitea.com/actions/checkout@'
    spec:
      file: '.gitea/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'
      engine: 'yamlpath'

  tag:
    dependson:
      - 'condition#tag:and'
    disableconditions: true
    name: 'deps(gitea): bump https://gitea.com/actions/checkout from v4 to {{ source "tag" }}'
    kind: 'yaml'
    sourceid: 'tag'
    transformers:
      - addprefix: 'https://gitea.com/actions/checkout@'
    spec:
      file: '.gitea/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'
      engine: 'yamlpath'

  branch:
    dependson:
      - 'condition#branch:and'
    disableconditions: true
    name: 'deps(gitea): bump https://gitea.com/actions/checkout from v4 to {{ source "branch" }}'
    kind: yaml
    sourceid: branch
    transformers:
      - addprefix: 'https://gitea.com/actions/checkout@'
    spec:
      file: '.gitea/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'
      engine: 'yamlpath'
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
      - addprefix: 'tibdex/github-app-token@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'
      engine: 'yamlpath'

  tag:
    dependson:
      - 'condition#tag:and'
    disableconditions: true
    name: 'deps(github): bump Action tag for tibdex/github-app-token from v2.1 to {{ source "tag" }}'
    kind: 'yaml'
    sourceid: 'tag'
    transformers:
      - addprefix: 'tibdex/github-app-token@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'
      engine: 'yamlpath'

  branch:
    dependson:
      - 'condition#branch:and'
    disableconditions: true
    name: 'deps(github): bump Action branch for tibdex/github-app-token from v2.1 to {{ source "branch" }}'
    kind: yaml
    sourceid: branch
    transformers:
      - addprefix: 'tibdex/github-app-token@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[0].uses'
      engine: 'yamlpath'
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
      - addprefix: 'tibdex/github-app-token@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[1].uses'
      engine: 'yamlpath'

  tag:
    dependson:
      - 'condition#tag:and'
    disableconditions: true
    name: 'deps(github): bump Action tag for tibdex/github-app-token from v2.1 to {{ source "tag" }}'
    kind: 'yaml'
    sourceid: 'tag'
    transformers:
      - addprefix: 'tibdex/github-app-token@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[1].uses'
      engine: 'yamlpath'

  branch:
    dependson:
      - 'condition#branch:and'
    disableconditions: true
    name: 'deps(github): bump Action branch for tibdex/github-app-token from v2.1 to {{ source "branch" }}'
    kind: yaml
    sourceid: branch
    transformers:
      - addprefix: 'tibdex/github-app-token@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[1].uses'
      engine: 'yamlpath'
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
      - addprefix: 'actions/checkout@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[2].uses'
      engine: 'yamlpath'

  tag:
    dependson:
      - 'condition#tag:and'
    disableconditions: true
    name: 'deps(github): bump Action tag for actions/checkout from v4 to {{ source "tag" }}'
    kind: 'yaml'
    sourceid: 'tag'
    transformers:
      - addprefix: 'actions/checkout@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[2].uses'
      engine: 'yamlpath'

  branch:
    dependson:
      - 'condition#branch:and'
    disableconditions: true
    name: 'deps(github): bump Action branch for actions/checkout from v4 to {{ source "branch" }}'
    kind: yaml
    sourceid: branch
    transformers:
      - addprefix: 'actions/checkout@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[2].uses'
      engine: 'yamlpath'
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
      - addprefix: 'updatecli/updatecli-action@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[3].uses'
      engine: 'yamlpath'

  tag:
    dependson:
      - 'condition#tag:and'
    disableconditions: true
    name: 'deps(github): bump Action tag for updatecli/updatecli-action from v2 to {{ source "tag" }}'
    kind: 'yaml'
    sourceid: 'tag'
    transformers:
      - addprefix: 'updatecli/updatecli-action@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[3].uses'
      engine: 'yamlpath'

  branch:
    dependson:
      - 'condition#branch:and'
    disableconditions: true
    name: 'deps(github): bump Action branch for updatecli/updatecli-action from v2 to {{ source "branch" }}'
    kind: yaml
    sourceid: branch
    transformers:
      - addprefix: 'updatecli/updatecli-action@'
    spec:
      file: '.github/workflows/updatecli.yaml'
      key: '$.jobs.updatecli.steps[3].uses'
      engine: 'yamlpath'
`},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			g, err := New(
				Spec{
					Credentials: tt.credentials,
					Digest:      &tt.digest,
				}, tt.rootDir, "", "")

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
