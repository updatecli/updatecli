package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/cache"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell/success/exitcode"
)

func TestRunSources(t *testing.T) {

	testdata := []struct {
		conf                   config.Config
		expectedSourcesResult  map[string]string
		expectedPipelineResult string
		expectedError          bool
	}{
		{
			conf: config.Config{
				Spec: config.Spec{
					Name: "Test a simple successful source pipeline",
					Sources: map[string]source.Config{
						"success": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "success",
								Spec: shell.Spec{
									Command: "true",
									ChangedIf: shell.SpecChangedIf{
										Kind: "exitcode",
										Spec: exitcode.Spec{
											Warning: 1, Success: 0, Failure: 2,
										},
									},
								},
							},
						},
					},
				},
			},
			expectedSourcesResult: map[string]string{
				"success": result.SUCCESS,
			},
			expectedPipelineResult: result.SUCCESS,
		},
		{
			conf: config.Config{
				Spec: config.Spec{
					Name: "Test a case with one source of each result type",
					Sources: map[string]source.Config{
						"success": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "success",
								Spec: shell.Spec{
									Command: "true",
								},
							},
						},
						"failed": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "failure",
								Spec: shell.Spec{
									Command: "false",
								},
							},
						},
					},
				},
			},
			expectedError: true,
			expectedSourcesResult: map[string]string{
				"success": result.SUCCESS,
				"failed":  result.FAILURE,
			},
			expectedPipelineResult: result.FAILURE,
		},
		{
			conf: config.Config{
				Spec: config.Spec{
					Name: "Test a case with a skipped source",
					Sources: map[string]source.Config{
						"success": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "success",
								Spec: shell.Spec{
									Command: "true",
								},
								DependsOn: []string{"condition#skip"},
							},
						},
					},
					Conditions: map[string]condition.Config{
						"skip": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "skip",
								Spec: shell.Spec{
									Command: "false",
								},
							},
						},
					},
				},
			},
			expectedError: false,
			expectedSourcesResult: map[string]string{
				"success": result.SKIPPED,
			},
			// As expected, the source is skipped because the condition is not met
			// so the pipeline result is considered as success
			expectedPipelineResult: result.SUCCESS,
		},
		{
			conf: config.Config{
				Spec: config.Spec{
					Name: "Test a case with a skipped source and warning second source",
					Sources: map[string]source.Config{
						"success": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "success",
								Spec: shell.Spec{
									Command: "true",
								},
								DependsOn: []string{"condition#skip"},
							},
						},
						"failed": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "failure",
								Spec: shell.Spec{
									Command: "false",
								},
							},
						},
					},
					Conditions: map[string]condition.Config{
						"skip": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "skip",
								Spec: shell.Spec{
									Command: "false",
								},
							},
							DisableSourceInput: true,
						},
					},
				},
			},
			expectedError: true,
			expectedSourcesResult: map[string]string{
				"success": result.SKIPPED,
				"failed":  result.FAILURE,
			},
			expectedPipelineResult: result.FAILURE,
		},
		{
			conf: config.Config{
				Spec: config.Spec{
					Name: "Test a case with a skipped source and a success second source",
					Sources: map[string]source.Config{
						"success": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "success",
								Spec: shell.Spec{
									Command: "true",
								},
								DependsOn: []string{"condition#skip"},
							},
						},
						"successbis": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "successbis",
								Spec: shell.Spec{
									Command: "true",
								},
							},
						},
					},
					Conditions: map[string]condition.Config{
						"skip": {
							ResourceConfig: resource.ResourceConfig{
								Kind: "shell",
								Name: "skip",
								Spec: shell.Spec{
									Command: "false",
								},
							},
							DisableSourceInput: true,
						},
					},
				},
			},
			expectedError: true,
			expectedSourcesResult: map[string]string{
				"success":    result.SKIPPED,
				"successbis": result.SUCCESS,
			},
			expectedPipelineResult: result.SUCCESS,
		},
	}

	for _, data := range testdata {
		t.Run(data.conf.Spec.Name, func(t *testing.T) {
			p := Pipeline{}
			err := p.Init(&data.conf, Options{})
			require.NoError(t, err)

			err = p.Run(context.Background())
			if !data.expectedError {
				require.NoError(t, err)
			}

			require.Equal(t, len(data.expectedSourcesResult), len(p.Sources))
			for id, result := range p.Sources {
				require.Equal(t, data.expectedSourcesResult[id], result.Result.Result)
			}
			require.Equal(t, data.expectedPipelineResult, p.Report.Result)
		})
	}

}

// TestRunSource_CacheMiss verifies that when the SourceCache has no entry,
// the source executes normally and writes its result to the cache.
func TestRunSource_CacheMiss(t *testing.T) {
	sourceID := "my-source"
	conf := config.Config{
		Spec: config.Spec{
			Name: "cache-miss pipeline",
			Sources: map[string]source.Config{
				sourceID: {
					ResourceConfig: resource.ResourceConfig{
						Kind: "shell",
						Name: sourceID,
						Spec: shell.Spec{
							Command: "echo hello",
							ChangedIf: shell.SpecChangedIf{
								Kind: "exitcode",
								Spec: exitcode.Spec{Warning: 1, Success: 0, Failure: 2},
							},
						},
					},
				},
			},
		},
	}

	p := Pipeline{}
	require.NoError(t, p.Init(&conf, Options{}))

	sc := cache.NewSourceCache()
	p.SourceCache = sc

	err := p.Run(context.Background())
	require.NoError(t, err)

	src := p.Sources[sourceID]
	assert.Equal(t, result.SUCCESS, src.Result.Result)
	assert.Equal(t, "hello", src.Result.Information)

	// Cache must have been populated with exactly one entry.
	assert.Equal(t, 1, sc.Len(), "cache must contain an entry after a successful source run")
}

// TestRunSource_CacheHit verifies that when a second pipeline reuses the same
// source config, the cache prevents re-execution. A temp file counter proves
// the shell command ran exactly once across two pipeline runs.
func TestRunSource_CacheHit(t *testing.T) {
	counterFile := filepath.Join(t.TempDir(), "counter")
	require.NoError(t, os.WriteFile(counterFile, []byte("0"), 0o644))

	// Shell command: increment counter file and echo a value.
	command := fmt.Sprintf(
		`count=$(cat %s); count=$((count+1)); echo $count > %s; echo hello`,
		counterFile, counterFile,
	)

	sourceID := "my-source"
	sourceCfg := source.Config{
		ResourceConfig: resource.ResourceConfig{
			Kind: "shell",
			Name: sourceID,
			Spec: shell.Spec{
				Command: command,
				ChangedIf: shell.SpecChangedIf{
					Kind: "exitcode",
					Spec: exitcode.Spec{Warning: 1, Success: 0, Failure: 2},
				},
			},
		},
	}

	sc := cache.NewSourceCache()

	// First run: executes the shell command, populates the cache.
	conf1 := config.Config{
		Spec: config.Spec{
			Name: "first pipeline",
			Sources: map[string]source.Config{sourceID: sourceCfg},
		},
	}
	p1 := Pipeline{}
	require.NoError(t, p1.Init(&conf1, Options{}))
	p1.SourceCache = sc
	require.NoError(t, p1.Run(context.Background()))
	assert.Equal(t, 1, sc.Len(), "first run must populate the cache")

	// Second run with the same cache: should hit, command must NOT run again.
	conf2 := config.Config{
		Spec: config.Spec{
			Name: "second pipeline",
			Sources: map[string]source.Config{sourceID: sourceCfg},
		},
	}
	p2 := Pipeline{}
	require.NoError(t, p2.Init(&conf2, Options{}))
	p2.SourceCache = sc
	require.NoError(t, p2.Run(context.Background()))

	src := p2.Sources[sourceID]
	assert.Equal(t, result.SUCCESS, src.Result.Result)
	assert.Equal(t, "hello", src.Result.Information,
		"second run must return the cached value")
	assert.Equal(t, 1, sc.Len())

	// The counter file proves the shell command executed exactly once.
	data, err := os.ReadFile(counterFile)
	require.NoError(t, err)
	assert.Equal(t, "1\n", string(data),
		"shell command must have executed exactly once across two pipeline runs")
}
