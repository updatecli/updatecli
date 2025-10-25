package dockerfile

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
		expectedPipelines []string
	}{
		{
			name:    "Scenario 1",
			rootDir: "testdata/updatecli-action",
			digest:  true,
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
        matcher: 'updatecli/updatecli'
    sourceid: 'updatecli/updatecli-digest'
`},
		},
		{
			name:    "Scenario 2: arg with suffix",
			rootDir: "testdata/jenkins",
			digest:  true,
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
			name:    "Scenario 3: Digest disabled",
			rootDir: "testdata/updatecli-action",
			digest:  false,
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
        matcher: 'updatecli/updatecli'
    sourceid: 'updatecli/updatecli'
`},
		},
		{
			name:    "Scenario 4: Reuse base image and scratch",
			rootDir: "testdata/scratch-and-base",
			digest:  true,
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
			name:    "Scenario 5: Should not update stage name as image",
			rootDir: "testdata/similar-stage-and-image",
			digest:  true,
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
        matcher: 'python'
    sourceid: 'python-digest'
`},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			digest := tt.digest
			dockerfile, err := New(
				Spec{
					Digest: &digest,
				}, tt.rootDir, "", "")
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
