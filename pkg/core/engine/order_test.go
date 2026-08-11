package engine

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
)

func TestOrderPipelines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		engine        Engine
		expectedOrder []string
		expectedError string
	}{
		{
			name: "sorts by manifest dependencies while preserving insertion order for siblings",
			engine: Engine{
				Pipelines: []*pipeline.Pipeline{
					newPipelineForOrderTest("app", "app", []string{"base"}),
					newPipelineForOrderTest("base", "base", nil),
					newPipelineForOrderTest("docs", "docs", nil),
				},
			},
			expectedOrder: []string{"base", "docs", "app"},
		},
		{
			name: "keeps legacy manifests without explicit ids in insertion order",
			engine: Engine{
				Pipelines: []*pipeline.Pipeline{
					newLegacyPipelineForOrderTest("first", "seed-1"),
					newLegacyPipelineForOrderTest("second", "seed-2"),
				},
			},
			expectedOrder: []string{"first", "second"},
		},
		{
			name: "depends on all manifests sharing the same id",
			engine: Engine{
				Pipelines: []*pipeline.Pipeline{
					newPipelineForOrderTest("app", "app", []string{"shared"}),
					newPipelineForOrderTest("first", "shared", nil),
					newPipelineForOrderTest("second", "shared", nil),
				},
			},
			expectedOrder: []string{"first", "second", "app"},
		},
		{
			name: "fails on unknown dependency target",
			engine: Engine{
				Pipelines: []*pipeline.Pipeline{
					newPipelineForOrderTest("app", "app", []string{"missing"}),
				},
			},
			expectedError: `depends on unknown manifest id "missing"`,
		},
		{
			name: "fails on dependency cycles",
			engine: Engine{
				Pipelines: []*pipeline.Pipeline{
					newPipelineForOrderTest("one", "one", []string{"two"}),
					newPipelineForOrderTest("two", "two", []string{"one"}),
				},
			},
			expectedError: "manifest dependency cycle detected",
		},
		{
			name: "fails when depending on its own shared id",
			engine: Engine{
				Pipelines: []*pipeline.Pipeline{
					newPipelineForOrderTest("one", "shared", []string{"shared"}),
					newPipelineForOrderTest("two", "shared", nil),
				},
			},
			expectedError: `cannot depend on its own id "shared"`,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.engine.OrderPipelines()

			if test.expectedError != "" {
				require.ErrorContains(t, err, test.expectedError)
				return
			}

			require.NoError(t, err)
			require.Len(t, test.engine.Pipelines, len(test.expectedOrder))
			require.Len(t, test.engine.configurations, len(test.expectedOrder))

			for i, expectedName := range test.expectedOrder {
				require.Equal(t, expectedName, test.engine.Pipelines[i].Name)
				require.Same(t, test.engine.Pipelines[i].Config, test.engine.configurations[i])
			}
		})
	}
}

func TestMergeManifestDependencies(t *testing.T) {
	t.Parallel()

	got := mergeManifestDependencies([]string{"child", "shared"}, []string{"parent", "shared"})

	require.Equal(t, []string{"parent", "shared", "child"}, got)
}

func newPipelineForOrderTest(name, id string, dependsOn []string) *pipeline.Pipeline {
	cfg := &config.Config{
		Spec: config.Spec{
			Name:      name,
			ID:        id,
			DependsOn: dependsOn,
		},
	}
	cfg.SetManifestID("seed/" + name)

	return &pipeline.Pipeline{
		Name:   name,
		Config: cfg,
	}
}

func newLegacyPipelineForOrderTest(name, seed string) *pipeline.Pipeline {
	cfg := &config.Config{
		Spec: config.Spec{
			Name: name,
		},
	}
	cfg.SetManifestID(seed)

	return &pipeline.Pipeline{
		Name:   name,
		Config: cfg,
	}
}
