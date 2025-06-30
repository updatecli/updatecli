package engine

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/engine/manifest"
)

func TestLoadConfigurations(t *testing.T) {
	tests := []struct {
		name              string
		wd                string
		engine            Engine
		expectedPipelines int
		expectedReports   int
		wantErr           bool
		expectedError     string
	}{
		{
			name: "Success - Default file",
			wd:   "testdata/defaultManifestFilename",
			engine: Engine{
				Options: Options{
					Manifests: []manifest.Manifest{
						{},
					},
				},
			},
			expectedPipelines: 1,
		},
		{
			name: "Success - Partial with one manifest",
			wd:   "testdata/partialOneManifest",
			engine: Engine{
				Options: Options{
					Manifests: []manifest.Manifest{
						{},
					},
				},
			},
			expectedPipelines: 1,
		},
		{
			name: "Success - Default manifest directory",
			wd:   "testdata/defaultManifestDirname_single",
			engine: Engine{
				Options: Options{
					Manifests: []manifest.Manifest{
						{},
					},
				},
			},
			expectedPipelines: 1,
		},
		{
			name: "Success - Default manifest directory multiple pipelines",
			wd:   "testdata/defaultManifestDirname_multiple",
			engine: Engine{
				Options: Options{
					Manifests: []manifest.Manifest{
						{},
					},
				},
			},
			expectedPipelines: 2,
		},
		{
			name: "Partial Success - Default manifest directory multiple pipelines one failure",
			wd:   "testdata/defaultManifestDirname_multiple_failure",
			engine: Engine{
				Options: Options{
					Manifests: []manifest.Manifest{
						{},
					},
				},
			},
			expectedPipelines: 1,
			expectedReports:   1,
			wantErr:           true,
			expectedError: `failed loading pipeline(s)
	* updatecli.d/failure.yaml:
		scm ID "updatecli" from source ID "adopters" doesn't exist`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(chdir(t, tt.wd))
			gotErr := tt.engine.LoadConfigurations()
			require.Equal(t, tt.expectedPipelines, len(tt.engine.Pipelines), "Pipelines count unexpected")
			require.Equal(t, tt.expectedReports, len(tt.engine.Reports), "Reports count unexpected")
			if tt.wantErr {
				require.ErrorContains(t, gotErr, tt.expectedError)
			} else {
				require.NoError(t, gotErr)
			}
		})
	}
}

func chdir(t *testing.T, dir string) func() {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	return func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restoring working directory: %v", err)
		}
	}
}
