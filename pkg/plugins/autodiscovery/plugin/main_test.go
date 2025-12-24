package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {

	tests := []struct {
		name              string
		path              string
		spec              Spec
		scmID             string
		actionID          string
		rootdir           string
		expectedPipelines []string
	}{
		{
			name: "test autodiscovery plugin",
			path: "testdata/demo.wasm",
			spec: Spec{
				Spec: map[string]any{
					"files": []string{"testdata/demo/data.txt"},
				},
			},
			expectedPipelines: []string{`name: 'deps: bump fluent/fluent-bit tag'
sources:
  'fluent/fluent-bit':
    name: 'get latest image tag for "fluent/fluent-bit"'
    kind: 'dockerimage'
    spec:
      image: 'fluent/fluent-bit'
      versionfilter:
        kind: 'semver'
        pattern: '*'
targets:
  'fluent/fluent-bit':
    name: 'deps: update Docker image "fluent/fluent-bit" to {{ source "fluent/fluent-bit" }}'
    kind: 'file'
    spec:
      file: 'testdata/demo/data.txt'
      matchpattern: '(.*) (harvester,release/harvester/v1.4-head,release/harvester/v1.4.3)'
      replacepattern: 'fluent/fluent-bit:{{ source "fluent/fluent-bit" }} harvester,release/harvester/v1.4-head,release/harvester/v1.4.3'
`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin, err := New(tt.spec, tt.rootdir, tt.scmID, tt.actionID, tt.path)
			if err != nil {
				require.NoError(t, err)
			}

			gotPipelines, err := plugin.DiscoverManifests()
			require.NoError(t, err)

			require.Equal(t, len(tt.expectedPipelines), len(gotPipelines))
			for i := range gotPipelines {
				assert.Equal(t, tt.expectedPipelines[i], string(gotPipelines[i]))
			}
		})
	}
}
