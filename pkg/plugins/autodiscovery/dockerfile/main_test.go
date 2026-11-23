package dockerfile

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile/simpletextparser"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestDiscoverManifests(t *testing.T) {

	testdata := []struct {
		name              string
		rootDir           string
		digest            bool
		versionFilter     *version.Filter
		expectedPipelines []string
	}{
		{
			name:          "Scenario 1",
			rootDir:       "testdata/updatecli-action",
			digest:        true,
			versionFilter: nil,
			expectedPipelines: []string{`name: 'deps(dockerfile): bump "updatecli/updatecli" digest'
sources:
  updatecli/updatecli:
    name: 'get latest image tag for "updatecli/updatecli"'
    kind: 'dockerimage'
    spec:
      image: 'updatecli/updatecli'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=v0.25.0'
  updatecli/updatecli-digest:
    name: 'get latest image "updatecli/updatecli" digest'
    kind: 'dockerdigest'
    spec:
      image: 'updatecli/updatecli'
      tag: '{{ source "updatecli/updatecli" }}'
    dependson:
      - 'updatecli/updatecli'
targets:
  updatecli/updatecli:
    name: 'deps: update Docker image "updatecli/updatecli" to "{{ source "updatecli/updatecli" }}"'
    kind: 'dockerfile'
    spec:
      file: 'Dockerfile'
      instruction:
        keyword: 'FROM'
        matcher: 'updatecli/updatecli:v0.25.0@sha256:1fafb0905264413501df60d90a92ca32df8a2011cbfb4876ddff5ceb20c8f165'
    sourceid: 'updatecli/updatecli-digest'
`},
		},
		{
			name:          "Scenario 2: arg with suffix",
			rootDir:       "testdata/jenkins",
			digest:        true,
			versionFilter: nil,
			expectedPipelines: []string{`name: 'deps(dockerfile): bump "jenkins/jenkins" digest'
sources:
  jenkins/jenkins:
    name: 'get latest image tag for "jenkins/jenkins"'
    kind: 'dockerimage'
    spec:
      image: 'jenkins/jenkins'
      tagfilter: '^\d*(\.\d*){2}-lts$'
      versionfilter:
        kind: 'semver'
        pattern: '>=2.235.1-lts'
  jenkins/jenkins-digest:
    name: 'get latest image "jenkins/jenkins" digest'
    kind: 'dockerdigest'
    spec:
      image: 'jenkins/jenkins'
      tag: '{{ source "jenkins/jenkins" }}'
    dependson:
      - 'jenkins/jenkins'
targets:
  jenkins/jenkins:
    name: 'deps: update Docker image "jenkins/jenkins" to "{{ source "jenkins/jenkins" }}"'
    kind: 'dockerfile'
    spec:
      file: 'Dockerfile'
      instruction:
        keyword: 'ARG'
        matcher: 'jenkins_version'
    sourceid: 'jenkins/jenkins-digest'
`},
		},
		{
			name:          "Scenario 3: Digest disabled",
			rootDir:       "testdata/updatecli-action",
			digest:        false,
			versionFilter: nil,
			expectedPipelines: []string{`name: 'deps(dockerfile): bump "updatecli/updatecli" tag'
sources:
  updatecli/updatecli:
    name: 'get latest image tag for "updatecli/updatecli"'
    kind: 'dockerimage'
    spec:
      image: 'updatecli/updatecli'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=v0.25.0'
targets:
  updatecli/updatecli:
    name: 'deps: update Docker image "updatecli/updatecli" to "{{ source "updatecli/updatecli" }}"'
    kind: 'dockerfile'
    spec:
      file: 'Dockerfile'
      instruction:
        keyword: 'FROM'
        matcher: 'updatecli/updatecli:v0.25.0@sha256:1fafb0905264413501df60d90a92ca32df8a2011cbfb4876ddff5ceb20c8f165'
    sourceid: 'updatecli/updatecli'
`},
		},
		{
			name:          "Scenario 4: Reuse base image and scratch",
			rootDir:       "testdata/scratch-and-base",
			digest:        true,
			versionFilter: nil,
			expectedPipelines: []string{`name: 'deps(dockerfile): bump "updatecli/updatecli" digest'
sources:
  updatecli/updatecli:
    name: 'get latest image tag for "updatecli/updatecli"'
    kind: 'dockerimage'
    spec:
      image: 'updatecli/updatecli'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=v0.25.0'
  updatecli/updatecli-digest:
    name: 'get latest image "updatecli/updatecli" digest'
    kind: 'dockerdigest'
    spec:
      image: 'updatecli/updatecli'
      tag: '{{ source "updatecli/updatecli" }}'
    dependson:
      - 'updatecli/updatecli'
targets:
  updatecli/updatecli:
    name: 'deps: update Docker image "updatecli/updatecli" to "{{ source "updatecli/updatecli" }}"'
    kind: 'dockerfile'
    spec:
      file: 'Dockerfile'
      instruction:
        keyword: 'ARG'
        matcher: 'updatecli_version'
    sourceid: 'updatecli/updatecli-digest'
`},
		},
		{
			name:          "Scenario 5: Should not update stage name as image",
			rootDir:       "testdata/similar-stage-and-image",
			digest:        true,
			versionFilter: nil,
			expectedPipelines: []string{`name: 'deps(dockerfile): bump "python" digest'
sources:
  python:
    name: 'get latest image tag for "python"'
    kind: 'dockerimage'
    spec:
      image: 'python'
      tagfilter: '^\d*(\.\d*){1}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=3.13'
  python-digest:
    name: 'get latest image "python" digest'
    kind: 'dockerdigest'
    spec:
      image: 'python'
      tag: '{{ source "python" }}'
    dependson:
      - 'python'
targets:
  python:
    name: 'deps: update Docker image "python" to "{{ source "python" }}"'
    kind: 'dockerfile'
    spec:
      file: 'Dockerfile'
      instruction:
        keyword: 'FROM'
        matcher: 'python:3.13'
    sourceid: 'python-digest'
`},
		},
		{
			name:          "Scenario 6: Alpine",
			rootDir:       "testdata/alpine",
			digest:        false,
			versionFilter: nil,
			expectedPipelines: []string{`name: 'deps(dockerfile): bump "alpine" tag'
sources:
  alpine:
    name: 'get latest image tag for "alpine"'
    kind: 'dockerimage'
    spec:
      image: 'alpine'
      tagfilter: '^\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=3.16.3'
targets:
  alpine:
    name: 'deps: update Docker image "alpine" to "{{ source "alpine" }}"'
    kind: 'dockerfile'
    spec:
      file: 'Dockerfile'
      instruction:
        keyword: 'ARG'
        matcher: 'alpine_version'
    sourceid: 'alpine'
`, `name: 'deps(dockerfile): bump "debian" tag'
sources:
  debian:
    name: 'get latest image tag for "debian"'
    kind: 'dockerimage'
    spec:
      image: 'debian'
      tagfilter: '^\d*$'
      versionfilter:
        kind: 'semver'
        pattern: '>=8'
targets:
  debian:
    name: 'deps: update Docker image "debian" to "{{ source "debian" }}"'
    kind: 'dockerfile'
    spec:
      file: 'Dockerfile'
      instruction:
        keyword: 'ARG'
        matcher: 'debian_version'
    sourceid: 'debian'
`, `name: 'deps(dockerfile): bump "opensuse" tag'
sources:
  opensuse:
    name: 'get latest image tag for "opensuse"'
    kind: 'dockerimage'
    spec:
      image: 'opensuse'
      tagfilter: '^\d*(\.\d*){1}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=15.4'
targets:
  opensuse:
    name: 'deps: update Docker image "opensuse" to "{{ source "opensuse" }}"'
    kind: 'dockerfile'
    spec:
      file: 'Dockerfile'
      instruction:
        keyword: 'FROM'
        matcher: 'opensuse:15.4'
    sourceid: 'opensuse'
`},
		},
		{
			name:          "Scenario 7: Multi-variable instructions should be ignored",
			rootDir:       "testdata/multi-variable",
			digest:        false,
			versionFilter: nil,
			expectedPipelines: []string{`name: 'deps(dockerfile): bump "node" tag'
sources:
  node:
    name: 'get latest image tag for "node"'
    kind: 'dockerimage'
    spec:
      image: 'node'
      tagfilter: '^\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=18.0.0'
targets:
  node:
    name: 'deps: update Docker image "node" to "{{ source "node" }}"'
    kind: 'dockerfile'
    spec:
      file: 'Dockerfile'
      instruction:
        keyword: 'ARG'
        matcher: 'node_version'
    sourceid: 'node'
`, `name: 'deps(dockerfile): bump "node" tag'
sources:
  node:
    name: 'get latest image tag for "node"'
    kind: 'dockerimage'
    spec:
      image: 'node'
      tagfilter: '^\d*$'
      versionfilter:
        kind: 'semver'
        pattern: '>=20'
targets:
  node:
    name: 'deps: update Docker image "node" to "{{ source "node" }}"'
    kind: 'dockerfile'
    spec:
      file: 'Dockerfile'
      instruction:
        keyword: 'ARG'
        matcher: 'my_version'
    sourceid: 'node'
`},
		},
		{
			name:    "Scenario 8: Version filter with regex/semver",
			rootDir: "testdata/python-slim",
			digest:  false,
			versionFilter: &version.Filter{
				Kind:    "regex/semver",
				Pattern: ">=3.14",
				Regex:   `^(\d+\.\d+)-slim$`,
			},
			expectedPipelines: []string{`name: 'deps(dockerfile): bump "python" tag'
sources:
  python:
    name: 'get latest image tag for "python"'
    kind: 'dockerimage'
    spec:
      image: 'python'
      tagfilter: '^\d*(\.\d*){1}-slim$'
      versionfilter:
        kind: 'regex/semver'
        pattern: '>=3.14'
        regex: '^(\d+\.\d+)-slim$'
targets:
  python:
    name: 'deps: update Docker image "python" to "{{ source "python" }}"'
    kind: 'dockerfile'
    spec:
      file: 'Dockerfile'
      instruction:
        keyword: 'FROM'
        matcher: 'python:3.14-slim'
    sourceid: 'python'
`},
		},
		{
			name:          "Scenario 9: Same image with different variant tags",
			rootDir:       "testdata/multi-variant",
			digest:        false,
			versionFilter: nil,
			expectedPipelines: []string{`name: 'deps(dockerfile): bump "caddy" tag'
sources:
  caddy:
    name: 'get latest image tag for "caddy"'
    kind: 'dockerimage'
    spec:
      image: 'caddy'
      tagfilter: '^\d*(\.\d*){2}-builder$'
      versionfilter:
        kind: 'semver'
        pattern: '>=2.11.4-builder'
targets:
  caddy:
    name: 'deps: update Docker image "caddy" to "{{ source "caddy" }}"'
    kind: 'dockerfile'
    spec:
      file: 'Dockerfile'
      instruction:
        keyword: 'FROM'
        matcher: 'caddy:2.11.4-builder'
    sourceid: 'caddy'
`, `name: 'deps(dockerfile): bump "caddy" tag'
sources:
  caddy:
    name: 'get latest image tag for "caddy"'
    kind: 'dockerimage'
    spec:
      image: 'caddy'
      tagfilter: '^\d*(\.\d*){2}-alpine$'
      versionfilter:
        kind: 'semver'
        pattern: '>=2.11.4-alpine'
targets:
  caddy:
    name: 'deps: update Docker image "caddy" to "{{ source "caddy" }}"'
    kind: 'dockerfile'
    spec:
      file: 'Dockerfile'
      instruction:
        keyword: 'FROM'
        matcher: 'caddy:2.11.4-alpine'
    sourceid: 'caddy'
`},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			digest := tt.digest
			spec := Spec{
				Digest: &digest,
			}
			if tt.versionFilter != nil {
				spec.VersionFilter = *tt.versionFilter
			}
			dockerfile, err := New(spec, tt.rootDir, "", "")
			require.NoError(t, err)

			rawPipelines, err := dockerfile.DiscoverManifests()
			require.NoError(t, err)

			if len(rawPipelines) == 0 {
				t.Errorf("No pipelines found for %s", tt.name)
			}

			var pipelines []string
			assert.Equal(t, len(tt.expectedPipelines), len(rawPipelines))

			for i := range rawPipelines {
				// We expect manifest generated by the autodiscovery to use the yaml syntax
				pipelines = append(pipelines, string(rawPipelines[i]))
				assert.Equal(t, tt.expectedPipelines[i], pipelines[i])
			}
		})
	}

}

// TestGeneratedFromTargetsUpdateOnlyTheirOwnLine feeds the matchers produced by
// the autodiscovery through the same simple text parser a dockerfile target
// uses at runtime. On a Dockerfile reusing one image with two variant tags,
// each generated target must rewrite exactly its own FROM line and leave the
// other one untouched, otherwise the targets overwrite each other.
func TestGeneratedFromTargetsUpdateOnlyTheirOwnLine(t *testing.T) {
	digest := false
	dockerfile, err := New(Spec{Digest: &digest}, "testdata/multi-variant", "", "")
	require.NoError(t, err)

	rawPipelines, err := dockerfile.DiscoverManifests()
	require.NoError(t, err)
	require.Len(t, rawPipelines, 2)

	matcherRegex := regexp.MustCompile(`(?m)^\s*matcher: '([^']+)'$`)
	keywordRegex := regexp.MustCompile(`(?m)^\s*keyword: '([^']+)'$`)

	var generatedMatchers []string
	for _, rawPipeline := range rawPipelines {
		pipeline := string(rawPipeline)

		keyword := keywordRegex.FindStringSubmatch(pipeline)
		require.Len(t, keyword, 2)
		require.Equal(t, "FROM", keyword[1])

		matcher := matcherRegex.FindStringSubmatch(pipeline)
		require.Len(t, matcher, 2)
		generatedMatchers = append(generatedMatchers, matcher[1])
	}
	require.ElementsMatch(t, []string{"caddy:2.11.4-builder", "caddy:2.11.4-alpine"}, generatedMatchers)

	originalContent, err := os.ReadFile(filepath.Join("testdata", "multi-variant", "Dockerfile"))
	require.NoError(t, err)

	// Each target resolves to a newer tag of its own variant; the pinned
	// matcher restricts which FROM line the target is allowed to rewrite.
	testdata := []struct {
		name                string
		matcher             string
		sourceValue         string
		expectedChangedLine int
		expectedDockerfile  []string
	}{
		{
			name:                "the builder target only updates the builder FROM line",
			matcher:             "caddy:2.11.4-builder",
			sourceValue:         "2.11.5-builder",
			expectedChangedLine: 1,
			expectedDockerfile: []string{
				"FROM caddy:2.11.5-builder AS builder",
				"RUN xcaddy build",
				"FROM caddy:2.11.4-alpine",
			},
		},
		{
			name:                "the alpine target only updates the alpine FROM line",
			matcher:             "caddy:2.11.4-alpine",
			sourceValue:         "2.11.5-alpine",
			expectedChangedLine: 3,
			expectedDockerfile: []string{
				"FROM caddy:2.11.4-builder AS builder",
				"RUN xcaddy build",
				"FROM caddy:2.11.5-alpine",
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := simpletextparser.NewSimpleTextDockerfileParser(map[string]string{
				"keyword": "FROM",
				"matcher": tt.matcher,
			})
			require.NoError(t, err)

			updatedContent, changedLines, err := parser.ReplaceInstructions(originalContent, tt.sourceValue, "")
			require.NoError(t, err)
			require.Len(t, changedLines, 1)
			require.Contains(t, changedLines, tt.expectedChangedLine)

			assert.Equal(t, strings.Join(tt.expectedDockerfile, "\n")+"\n", string(updatedContent))
		})
	}
}
