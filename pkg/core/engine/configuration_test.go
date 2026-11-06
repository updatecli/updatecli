package engine

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/engine/manifest"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/scms/gitlab"
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
		{
			// A templating error must point at the partial the broken line lives in,
			// and at its line number in that file, rather than at an offset into the
			// concatenation of every partial and the manifest.
			name: "Failure - Templating error inside a partial",
			wd:   "testdata/partialBrokenTemplate",
			engine: Engine{
				Options: Options{
					Manifests: []manifest.Manifest{
						{},
					},
				},
			},
			expectedPipelines: 0,
			expectedReports:   1,
			wantErr:           true,
			expectedError: `failed loading pipeline(s)
	* updatecli.d/updatecli.yaml:
		* Partial files:
			* updatecli.d/_source.yaml
		* Error:
			updatecli.d/_source.yaml:8: unexpected "{" in command
			
			  5 |   broken:
			  6 |     kind: json
			  7 |     spec:
			> 8 |       value: {{{ .version }}
			  9 |       file: jenkins.json`,
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

func TestLoadConfigurationsSearch(t *testing.T) {
	for _, tt := range []struct {
		name  string
		empty bool
	}{
		{name: "Two searches"},
		{name: "No matching repositories", empty: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.URL.Path == "/api/v4/groups/first/projects":
					if tt.empty {
						_, _ = w.Write([]byte(`[]`))
					} else {
						_, _ = w.Write([]byte(`[{"path_with_namespace":"first/one"},{"path_with_namespace":"first/two"}]`))
					}
				case r.URL.Path == "/api/v4/groups/second/projects":
					_, _ = w.Write([]byte(`[{"path_with_namespace":"second/three"},{"path_with_namespace":"second/four"},{"path_with_namespace":"second/five"}]`))
				case strings.HasSuffix(r.URL.Path, "/repository/branches"):
					_, _ = w.Write([]byte(`[{"name":"main"}]`))
				default:
					t.Errorf("unexpected request: %s", r.URL)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			filename := filepath.Join(t.TempDir(), "updatecli.yaml")
			require.NoError(t, os.WriteFile(filename, []byte(fmt.Sprintf(`name: Search expansion
options:
  relativepaths: manifest
scms:
  first:
    kind: gitlabsearch
    spec:
      url: %s
      group: first
      directory: checkouts/first
  second:
    kind: gitlabsearch
    spec:
      url: %s
      group: second
      directory: checkouts/second
sources:
  version:
    kind: shell
    spec:
      command: echo unused
targets:
  version:
    kind: file
    sourceid: version
    spec:
      file: version.txt
      content: '{{ source "version" }}'
`, server.URL, server.URL)), 0600))
			e := Engine{Options: Options{Manifests: []manifest.Manifest{{Manifests: []string{filename}}}}}
			err := e.LoadConfigurations()
			if tt.empty {
				require.EqualError(t, err, `no scm discovered for gitlabsearch scm "first"`)
				assert.Empty(t, e.Pipelines)
				return
			}
			require.NoError(t, err)
			require.Len(t, e.Pipelines, 6)
			require.Len(t, e.configurations, 6)
			var pairs []string
			for _, p := range e.Pipelines {
				first := p.SCMs["first"].Handler.(*gitlab.Gitlab)
				second := p.SCMs["second"].Handler.(*gitlab.Gitlab)
				pair := first.Spec.Repository + "/" + second.Spec.Repository
				pairs = append(pairs, pair)
				assert.Equal(t, "main", first.Spec.Branch)
				assert.Equal(t, "main", second.Spec.Branch)
				assert.Equal(t, filepath.Join(filepath.Dir(filename), "checkouts", "first"), first.GetDirectory())
				assert.Equal(t, filepath.Join(filepath.Dir(filename), "checkouts", "second"), second.GetDirectory())

				version := p.Sources["version"]
				version.Output = pair
				version.Result.Result = result.SUCCESS
				p.Sources["version"] = version
				require.NoError(t, p.Update())
				assert.Equal(t, map[string]any{"file": "version.txt", "content": pair}, p.Targets["version"].Config.Spec)
			}
			assert.ElementsMatch(t, []string{
				"one/three", "one/four", "one/five",
				"two/three", "two/four", "two/five",
			}, pairs)
			for i, c := range e.configurations {
				assert.Equal(t, filepath.Dir(filename), c.BaseDir())
				assert.NotEmpty(t, c.ManifestID())
				assert.Equal(t, e.configurations[0].ManifestID(), c.ManifestID())
				scmConfig := c.Spec.SCMs["first"]
				scmConfig.Kind = fmt.Sprintf("copy-%d", i)
				c.Spec.SCMs["first"] = scmConfig
			}
			for i, c := range e.configurations {
				assert.Equal(t, fmt.Sprintf("copy-%d", i), c.Spec.SCMs["first"].Kind)
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
