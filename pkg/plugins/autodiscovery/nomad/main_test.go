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
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			digest := tt.digest
			nomad, err := New(
				Spec{
					Digest: &digest,
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
