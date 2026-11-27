package nomad

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverManifests(t *testing.T) {

	testdata := []struct {
		name              string
		rootDir           string
		digest            bool
		scmID             string
		actionID          string
		registryMappings  map[string]string
		expectedPipelines []string
	}{
		{
			name:    "Scenario 1 - digest",
			rootDir: "testdata/simple",
			digest:  true,
			expectedPipelines: []string{`name: 'deps(nomad): update Docker image digest for "nginx" digest'
sources:
  default-digest:
    name: 'get latest image "nginx" digest'
    kind: 'dockerdigest'
    spec:
      image: 'nginx'
      tag: 'latest'
targets:
  default:
    name: 'deps(nomad): update Docker image digest for "nginx:latest"'
    kind: 'hcl'
    spec:
      file: 'nomad.hcl'
      path: 'job.multi-docker-example.group.web-group.task.frontend.config.image'
    sourceid: 'default-digest'
    transformers:
      - addprefix: 'nginx:'
`, `name: 'deps(nomad): update Docker image digest for "hashicorp/http-echo" digest'
sources:
  default-digest:
    name: 'get latest image "hashicorp/http-echo" digest'
    kind: 'dockerdigest'
    spec:
      image: 'hashicorp/http-echo'
      tag: 'latest'
targets:
  default:
    name: 'deps(nomad): update Docker image digest for "hashicorp/http-echo:latest"'
    kind: 'hcl'
    spec:
      file: 'nomad.hcl'
      path: 'job.multi-docker-example.group.web-group.task.backend.config.image'
    sourceid: 'default-digest'
    transformers:
      - addprefix: 'hashicorp/http-echo:'
`},
		},
		{
			name:     "Scenario 2 - with variable task",
			rootDir:  "testdata/variable",
			actionID: "default",
			scmID:    "default",
			digest:   true,
			expectedPipelines: []string{`name: 'deps(nomad): update Docker image digest for "grafana/grafana" digest'
actions:
  default:
    title: 'deps(nomad): update Docker image digest for "grafana/grafana:latest"'
sources:
  default-digest:
    name: 'get latest image "grafana/grafana" digest'
    kind: 'dockerdigest'
    spec:
      image: 'grafana/grafana'
      tag: 'latest'
targets:
  default:
    name: 'deps(nomad): update Docker image digest for "grafana/grafana:latest"'
    kind: 'hcl'
    scmid: 'default'
    spec:
      file: 'grafana.nomad'
      path: 'variable.image_tag.default'
    sourceid: 'default-digest'
`},
		},
		{
			name:     "Scenario 3 - handle podman",
			rootDir:  "testdata/podman",
			actionID: "default",
			scmID:    "default",
			digest:   true,
			expectedPipelines: []string{`name: 'deps(nomad): update Docker image digest "docker.io/library/redis"'
actions:
  default:
    title: 'deps(nomad): update Docker image "docker.io/library/redis" to "{{ source "default" }}"'
sources:
  default:
    name: 'get latest image tag for "docker.io/library/redis"'
    kind: 'dockerimage'
    spec:
      image: 'docker.io/library/redis'
      tagfilter: '^\d*$'
      versionfilter:
        kind: 'semver'
        pattern: '>=7'
  default-digest:
    name: 'get latest image "docker.io/library/redis" digest'
    kind: 'dockerdigest'
    spec:
      image: 'docker.io/library/redis'
      tag: '{{ source "default" }}'
    dependson:
      - 'default'
targets:
  default:
    name: 'deps(nomad): update Docker image "docker.io/library/redis" to "{{ source "default" }}"'
    kind: 'hcl'
    scmid: 'default'
    spec:
      file: 'cache.nomad'
      path: 'job.redis.group.cache.task.redis.config.image'
    sourceid: 'default-digest'
    transformers:
      - addprefix: 'docker.io/library/redis:'
`},
		},
		{
			name:     "Scenario 4 - handle containerd",
			rootDir:  "testdata/containerd",
			actionID: "default",
			scmID:    "default",
			digest:   true,
			expectedPipelines: []string{`name: 'deps(nomad): update Docker image digest "docker.io/library/redis"'
actions:
  default:
    title: 'deps(nomad): update Docker image "docker.io/library/redis" to "{{ source "default" }}"'
sources:
  default:
    name: 'get latest image tag for "docker.io/library/redis"'
    kind: 'dockerimage'
    spec:
      image: 'docker.io/library/redis'
      tagfilter: '^\d*$'
      versionfilter:
        kind: 'semver'
        pattern: '>=7'
  default-digest:
    name: 'get latest image "docker.io/library/redis" digest'
    kind: 'dockerdigest'
    spec:
      image: 'docker.io/library/redis'
      tag: '{{ source "default" }}'
    dependson:
      - 'default'
targets:
  default:
    name: 'deps(nomad): update Docker image "docker.io/library/redis" to "{{ source "default" }}"'
    kind: 'hcl'
    scmid: 'default'
    spec:
      file: 'redis.nomad'
      path: 'job.redis.group.redis-group.task.redis-task.config.image'
    sourceid: 'default-digest'
    transformers:
      - addprefix: 'docker.io/library/redis:'
`},
		},
		{
			name:             "Scenario - registry mappings",
			rootDir:          "testdata/registryMappings",
			digest:           false,
			registryMappings: map[string]string{"my-registry.com/ghcr-io/": "ghcr.io/"},
			expectedPipelines: []string{`name: 'deps(dockercompose): bump "my-registry.com/ghcr-io/updatecli/updatecli" tag'
sources:
  default:
    name: 'get latest image tag for "ghcr.io/updatecli/updatecli"'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=v0.67.0'
conditions:
  default-mirror:
    name: 'check container image "my-registry.com/ghcr-io/updatecli/updatecli:{{ source "default" }}" is available on the mirror'
    kind: 'dockerimage'
    sourceid: 'default'
    spec:
      image: 'my-registry.com/ghcr-io/updatecli/updatecli'
targets:
  default:
    name: 'deps(nomad): update Docker image "my-registry.com/ghcr-io/updatecli/updatecli" to "{{ source "default" }}"'
    kind: 'hcl'
    spec:
      file: 'nomad.hcl'
      path: 'job.mirrored.group.updatecli.task.updatecli.config.image'
    sourceid: 'default'
    transformers:
      - addprefix: 'my-registry.com/ghcr-io/updatecli/updatecli:'
`},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			digest := tt.digest
			nomad, err := New(
				Spec{
					Digest:           &digest,
					RegistryMappings: tt.registryMappings,
				}, tt.rootDir, tt.scmID, tt.actionID)

			require.NoError(t, err)

			rawPipelines, err := nomad.DiscoverManifests()
			require.NoError(t, err)

			if len(rawPipelines) == 0 {
				t.Errorf("No pipelines found for %s", tt.name)
			}

			var pipelines []string
			assert.Equal(t, len(tt.expectedPipelines), len(rawPipelines))

			require.NoError(t, err)
			for i := range rawPipelines {
				pipelines = append(pipelines, string(rawPipelines[i]))
				assert.Equal(t, tt.expectedPipelines[i], pipelines[i])
			}
		})
	}
}
