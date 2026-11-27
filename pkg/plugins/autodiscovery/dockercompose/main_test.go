package dockercompose

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestDiscoverManifests(t *testing.T) {
	testdata := []struct {
		name              string
		rootDir           string
		digest            bool
		Auths             map[string]docker.InlineKeyChain
		versionFilter     version.Filter
		registryMappings  map[string]string
		expectedPipelines []string
	}{
		{
			name:    "Scenario 1 - digest",
			rootDir: "testdata",
			digest:  true,
			expectedPipelines: []string{
				`name: 'deps(dockercompose): bump "jenkinsci/jenkins" digest'
sources:
  jenkins-lts:
    name: 'get latest image tag for "jenkinsci/jenkins"'
    kind: 'dockerimage'
    spec:
      image: 'jenkinsci/jenkins'
      tagfilter: '^\d*(\.\d*){2}-alpine$'
      versionfilter:
        kind: 'semver'
        pattern: '>=2.150.1-alpine'
  jenkins-lts-digest:
    name: 'get latest image "jenkinsci/jenkins" digest'
    kind: 'dockerdigest'
    spec:
      image: 'jenkinsci/jenkins'
      tag: '{{ source "jenkins-lts" }}'
    dependson:
      - 'jenkins-lts'
targets:
  jenkins-lts:
    name: 'deps: update Docker image "jenkinsci/jenkins" to "{{ source "jenkins-lts" }}"'
    kind: 'yaml'
    spec:
      file: 'docker-compose.yaml'
      key: '$.services.jenkins-lts.image'
    sourceid: 'jenkins-lts-digest'
    transformers:
      - addprefix: 'jenkinsci/jenkins:'
`, `name: 'deps(dockercompose): bump "jenkinsci/jenkins" digest'
sources:
  jenkins-weekly:
    name: 'get latest image tag for "jenkinsci/jenkins"'
    kind: 'dockerimage'
    spec:
      architecture: 'amd64'
      image: 'jenkinsci/jenkins'
      tagfilter: '^\d*(\.\d*){1}-alpine$'
      versionfilter:
        kind: 'semver'
        pattern: '>=2.254-alpine'
  jenkins-weekly-digest:
    name: 'get latest image "jenkinsci/jenkins" digest'
    kind: 'dockerdigest'
    spec:
      image: 'jenkinsci/jenkins'
      tag: '{{ source "jenkins-weekly" }}'
    dependson:
      - 'jenkins-weekly'
targets:
  jenkins-weekly:
    name: 'deps: update Docker image "jenkinsci/jenkins" to "{{ source "jenkins-weekly" }}"'
    kind: 'yaml'
    spec:
      file: 'docker-compose.yaml'
      key: '$.services.jenkins-weekly.image'
    sourceid: 'jenkins-weekly-digest'
    transformers:
      - addprefix: 'jenkinsci/jenkins:'
`,
			},
		},
		{
			name:    "Scenario 2 - no digest",
			rootDir: "testdata",
			digest:  false,
			Auths: map[string]docker.InlineKeyChain{
				"index.docker.io": {
					Token: "mysecretToken",
				},
			},
			expectedPipelines: []string{
				`name: 'deps(dockercompose): bump "jenkinsci/jenkins" tag'
sources:
  jenkins-lts:
    name: 'get latest image tag for "jenkinsci/jenkins"'
    kind: 'dockerimage'
    spec:
      image: 'jenkinsci/jenkins'
      tagfilter: '^\d*(\.\d*){2}-alpine$'
      versionfilter:
        kind: 'semver'
        pattern: '>=2.150.1-alpine'
      token: 'mysecretToken'
targets:
  jenkins-lts:
    name: 'deps: update Docker image "jenkinsci/jenkins" to "{{ source "jenkins-lts" }}"'
    kind: 'yaml'
    spec:
      file: 'docker-compose.yaml'
      key: '$.services.jenkins-lts.image'
    sourceid: 'jenkins-lts'
    transformers:
      - addprefix: 'jenkinsci/jenkins:'
`, `name: 'deps(dockercompose): bump "jenkinsci/jenkins" tag'
sources:
  jenkins-weekly:
    name: 'get latest image tag for "jenkinsci/jenkins"'
    kind: 'dockerimage'
    spec:
      architecture: 'amd64'
      image: 'jenkinsci/jenkins'
      tagfilter: '^\d*(\.\d*){1}-alpine$'
      versionfilter:
        kind: 'semver'
        pattern: '>=2.254-alpine'
      token: 'mysecretToken'
targets:
  jenkins-weekly:
    name: 'deps: update Docker image "jenkinsci/jenkins" to "{{ source "jenkins-weekly" }}"'
    kind: 'yaml'
    spec:
      file: 'docker-compose.yaml'
      key: '$.services.jenkins-weekly.image'
    sourceid: 'jenkins-weekly'
    transformers:
      - addprefix: 'jenkinsci/jenkins:'
`,
			},
		},
		{
			// A tag that Masterminds semver cannot parse at all leaves a nil
			// dockerimage spec and fails GreaterThanPattern. That service has to be
			// skipped without crashing, and without discarding the manifest already
			// found for the service listed before it.
			name:          "Scenario 3 - no digest, tag not parseable as a version",
			rootDir:       "testdata-unparsable-tag",
			digest:        false,
			versionFilter: version.Filter{Kind: "semver", Pattern: "patch"},
			expectedPipelines: []string{
				`name: 'deps(dockercompose): bump "jenkinsci/jenkins" tag'
sources:
  a-pinned:
    name: 'get latest image tag for "jenkinsci/jenkins"'
    kind: 'dockerimage'
    spec:
      image: 'jenkinsci/jenkins'
      tagfilter: '^\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '2.150.x'
targets:
  a-pinned:
    name: 'deps: update Docker image "jenkinsci/jenkins" to "{{ source "a-pinned" }}"'
    kind: 'yaml'
    spec:
      file: 'docker-compose.yaml'
      key: '$.services.a-pinned.image'
    sourceid: 'a-pinned'
    transformers:
      - addprefix: 'jenkinsci/jenkins:'
`,
			},
		},
		{
			name:             "Scenario 4 - registry mappings",
			rootDir:          "testdata-registry-mappings",
			digest:           false,
			registryMappings: map[string]string{"my-registry.com/ghcr-io/": "ghcr.io/"},
			expectedPipelines: []string{`name: 'deps(dockercompose): bump "my-registry.com/ghcr-io/updatecli/updatecli" tag'
sources:
  updatecli:
    name: 'get latest image tag for "ghcr.io/updatecli/updatecli"'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=v0.67.0'
conditions:
  updatecli-mirror:
    name: 'check container image "my-registry.com/ghcr-io/updatecli/updatecli:{{ source "updatecli" }}" is available on the mirror'
    kind: 'dockerimage'
    sourceid: 'updatecli'
    spec:
      image: 'my-registry.com/ghcr-io/updatecli/updatecli'
targets:
  updatecli:
    name: 'deps: update Docker image "my-registry.com/ghcr-io/updatecli/updatecli" to "{{ source "updatecli" }}"'
    kind: 'yaml'
    spec:
      file: 'compose.yaml'
      key: '$.services.updatecli.image'
    sourceid: 'updatecli'
    transformers:
      - addprefix: 'my-registry.com/ghcr-io/updatecli/updatecli:'
`},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			digest := tt.digest
			composefile, err := New(
				Spec{
					Digest:           &digest,
					Auths:            tt.Auths,
					VersionFilter:    tt.versionFilter,
					RegistryMappings: tt.registryMappings,
				}, tt.rootDir, "", "")

			require.NoError(t, err)

			rawPipelines, err := composefile.DiscoverManifests()
			require.NoError(t, err)

			if len(rawPipelines) == 0 {
				t.Errorf("No pipelines found for %s", tt.name)
			}

			var pipelines []string
			assert.Equal(t, len(tt.expectedPipelines), len(rawPipelines))

			require.NoError(t, err)
			for i := range rawPipelines {
				// We expect manifest generated by the autodiscovery to use the yaml syntax
				pipelines = append(pipelines, string(rawPipelines[i]))
				assert.Equal(t, tt.expectedPipelines[i], pipelines[i])
			}
		})
	}
}
