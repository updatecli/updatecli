package maven

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverManifests(t *testing.T) {
	testdata := []struct {
		name              string
		rootDir           string
		expectedPipelines []string
	}{
		{
			name:    "Scenario - with external http mirrorof",
			rootDir: "testdata/externalhttp",
			expectedPipelines: []string{`name: 'Bump Maven dependency org.apache.commons/commons-lang3'
sources:
  org.apache.commons/commons-lang3:
    name: 'Get latest Maven Artifact version "org.apache.commons/commons-lang3"'
    kind: 'maven'
    spec:
      artifactid: 'commons-lang3'
      groupid: 'org.apache.commons'
      repositories:
        - 'https://mirror.example.com/maven'
      versionfilter:
        kind: 'latest'
        pattern: 'latest'
conditions:
  artifactid:
    name: 'Ensure dependency artifactId "commons-lang3" is specified'
    kind: 'xml'
    spec:
      file: 'pom.xml'
      path: '/project/dependencies/dependency[1]/artifactId'
      value: 'commons-lang3'
    disablesourceinput: true
  groupid:
    name: 'Ensure dependency groupId "org.apache.commons" is specified'
    kind: 'xml'
    spec:
      file: 'pom.xml'
      path: '/project/dependencies/dependency[1]/groupId'
      value: 'org.apache.commons'
    disablesourceinput: true
targets:
  org.apache.commons/commons-lang3:
    name: 'deps(maven): update "org.apache.commons/commons-lang3" to {{ source "org.apache.commons/commons-lang3" }}'
    kind: 'xml'
    spec:
      file: 'pom.xml'
      path: '/project/dependencies/dependency[1]/version'
    sourceid: 'org.apache.commons/commons-lang3'
`},
		},
		{
			name:    "Scenario - with excluded mirrorof",
			rootDir: "testdata/exclude",
			expectedPipelines: []string{`name: 'Bump Maven dependency org.apache.commons/commons-lang3'
sources:
  org.apache.commons/commons-lang3:
    name: 'Get latest Maven Artifact version "org.apache.commons/commons-lang3"'
    kind: 'maven'
    spec:
      artifactid: 'commons-lang3'
      groupid: 'org.apache.commons'
      repositories:
        - 'https://foo:bar@mirror.example.com/maven'
        - 'http://bar-repo.example.com/maven'
      versionfilter:
        kind: 'latest'
        pattern: 'latest'
conditions:
  artifactid:
    name: 'Ensure dependency artifactId "commons-lang3" is specified'
    kind: 'xml'
    spec:
      file: 'pom.xml'
      path: '/project/dependencies/dependency[1]/artifactId'
      value: 'commons-lang3'
    disablesourceinput: true
  groupid:
    name: 'Ensure dependency groupId "org.apache.commons" is specified'
    kind: 'xml'
    spec:
      file: 'pom.xml'
      path: '/project/dependencies/dependency[1]/groupId'
      value: 'org.apache.commons'
    disablesourceinput: true
targets:
  org.apache.commons/commons-lang3:
    name: 'deps(maven): update "org.apache.commons/commons-lang3" to {{ source "org.apache.commons/commons-lang3" }}'
    kind: 'xml'
    spec:
      file: 'pom.xml'
      path: '/project/dependencies/dependency[1]/version'
    sourceid: 'org.apache.commons/commons-lang3'
`},
		},
		{
			name:    "Scenario - Default",
			rootDir: "testdata/default",
			expectedPipelines: []string{`name: 'Bump Maven dependency com.jcraft/jsch'
sources:
  com.jcraft/jsch:
    name: 'Get latest Maven Artifact version "com.jcraft/jsch"'
    kind: 'maven'
    spec:
      artifactid: 'jsch'
      groupid: 'com.jcraft'
      repositories:
        - 'https://repo.jenkins-ci.org/public/'
        - 'https://mirror.example.com/maven2'
      versionfilter:
        kind: 'latest'
        pattern: 'latest'
conditions:
  artifactid:
    name: 'Ensure dependency artifactId "jsch" is specified'
    kind: 'xml'
    spec:
      file: 'pom.xml'
      path: '/project/dependencies/dependency[1]/artifactId'
      value: 'jsch'
    disablesourceinput: true
  groupid:
    name: 'Ensure dependency groupId "com.jcraft" is specified'
    kind: 'xml'
    spec:
      file: 'pom.xml'
      path: '/project/dependencies/dependency[1]/groupId'
      value: 'com.jcraft'
    disablesourceinput: true
targets:
  com.jcraft/jsch:
    name: 'deps(maven): update "com.jcraft/jsch" to {{ source "com.jcraft/jsch" }}'
    kind: 'xml'
    spec:
      file: 'pom.xml'
      path: '/project/dependencies/dependency[1]/version'
    sourceid: 'com.jcraft/jsch'
`, `name: 'Bump Maven dependencyManagement io.jenkins.tools.bom/bom-2.346.x'
sources:
  io.jenkins.tools.bom/bom-2.346.x:
    name: 'Get latest Maven Artifact version "io.jenkins.tools.bom/bom-2.346.x"'
    kind: 'maven'
    spec:
      artifactid: 'bom-2.346.x'
      groupid: 'io.jenkins.tools.bom'
      repositories:
        - 'https://repo.jenkins-ci.org/public/'
        - 'https://mirror.example.com/maven2'
      versionfilter:
        kind: 'latest'
        pattern: 'latest'
conditions:
  artifactid:
    name: 'Ensure dependencyManagement artifactId "bom-2.346.x" is specified'
    kind: 'xml'
    spec:
      file: 'pom.xml'
      path: '/project/dependencyManagement/dependencies/dependency[1]/artifactId'
      value: 'bom-2.346.x'
    disablesourceinput: true
  groupid:
    name: 'Ensure dependencyManagement groupId "io.jenkins.tools.bom" is specified'
    kind: 'xml'
    spec:
      file: 'pom.xml'
      path: '/project/dependencyManagement/dependencies/dependency[1]/groupId'
      value: 'io.jenkins.tools.bom'
    disablesourceinput: true
targets:
  io.jenkins.tools.bom/bom-2.346.x:
    name: 'deps(maven): update "io.jenkins.tools.bom/bom-2.346.x" to {{ source "io.jenkins.tools.bom/bom-2.346.x" }}'
    kind: 'xml'
    spec:
      file: 'pom.xml'
      path: '/project/dependencyManagement/dependencies/dependency[1]/version'
    sourceid: 'io.jenkins.tools.bom/bom-2.346.x'
`, `name: 'Bump Maven parent Pom org.jenkins-ci.plugins/plugin'
sources:
  org.jenkins-ci.plugins/plugin:
    name: 'Get latest Parent Pom Artifact version "org.jenkins-ci.plugins/plugin"'
    kind: 'maven'
    spec:
      artifactid: 'plugin'
      groupid: 'org.jenkins-ci.plugins'
      repositories:
        - 'https://repo.jenkins-ci.org/public/'
        - 'https://mirror.example.com/maven2'
      versionfilter:
        kind: 'latest'
        pattern: 'latest'
conditions:
  artifactid:
    name: 'Ensure parent artifactId "plugin" is specified'
    kind: 'xml'
    spec:
      file: 'pom.xml'
      path: '/project/parent/artifactId'
      value: 'plugin'
    disablesourceinput: true
  groupid:
    name: 'Ensure parent pom.xml groupId "org.jenkins-ci.plugins" is specified'
    kind: 'xml'
    spec:
      file: 'pom.xml'
      path: '/project/parent/groupId'
      value: 'org.jenkins-ci.plugins'
    disablesourceinput: true
targets:
  org.jenkins-ci.plugins/plugin:
    name: 'deps(maven): update "org.jenkins-ci.plugins/plugin" to {{ source "org.jenkins-ci.plugins/plugin" }}'
    kind: 'xml'
    spec:
      file: 'pom.xml'
      path: '/project/parent/version'
    sourceid: 'org.jenkins-ci.plugins/plugin'
`},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			maven, err := New(
				Spec{}, tt.rootDir, "", "")

			require.NoError(t, err)

			var pipelines []string
			rawPipelines, err := maven.DiscoverManifests()
			require.NoError(t, err)

			assert.Equal(t, len(tt.expectedPipelines), len(rawPipelines))

			for i := range rawPipelines {
				// We expect manifest generated by the autodiscovery to use the yaml syntax
				pipelines = append(pipelines, string(rawPipelines[i]))
				assert.Equal(t, tt.expectedPipelines[i], pipelines[i])
			}
		})
	}

}
