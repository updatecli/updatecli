package maven

import (
	"testing"

	"github.com/beevik/etree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetParentFromPom(t *testing.T) {
	testdata := []struct {
		name           string
		file           string
		expectedParent parentPom
	}{
		{
			name: "Test 1",
			file: "testdata/default/pom.xml",
			expectedParent: parentPom{
				GroupID:    "org.jenkins-ci.plugins",
				ArtifactID: "plugin",
				Version:    "4.43.1",
			},
		},
	}

	for _, tt := range testdata {
		doc := etree.NewDocument()
		if err := doc.ReadFromFile(tt.file); err != nil {
			require.NoError(t, err)
		}
		t.Run(tt.name, func(t *testing.T) {
			gotParent := getParentFromPom(doc)
			assert.Equal(t, tt.expectedParent, gotParent)
		})
	}
}

func TestGetRepositoriesFromPom(t *testing.T) {
	testdata := []struct {
		name                 string
		file                 string
		expectedRepositories []repository
	}{
		{
			name: "Test 1",
			file: "testdata/default/pom.xml",
			expectedRepositories: []repository{
				{
					URL: "https://repo.jenkins-ci.org/public/",
					ID:  "repo.jenkins-ci.org",
				},
			},
		},
	}

	for _, tt := range testdata {
		doc := etree.NewDocument()
		if err := doc.ReadFromFile(tt.file); err != nil {
			require.NoError(t, err)
		}
		t.Run(tt.name, func(t *testing.T) {
			gotRepositories := getRepositoriesFromPom(doc)
			assert.Equal(t, tt.expectedRepositories, gotRepositories)
		})
	}
}

func TestGetDependenciesFromPom(t *testing.T) {
	testdata := []struct {
		name                 string
		file                 string
		expectedDependencies []dependency
	}{
		{
			name: "Test 1",
			file: "testdata/default/pom.xml",
			expectedDependencies: []dependency{
				{
					GroupID:    "com.jcraft",
					ArtifactID: "jsch",
					Version:    "0.1.55",
				},
				{
					GroupID:    "org.jenkins-ci.plugins",
					ArtifactID: "trilead-api",
				},
				{

					GroupID:    "org.jenkins-ci.plugins",
					ArtifactID: "credentials",
				},
				{
					GroupID:    "io.jenkins.plugins.mina-sshd-api",
					ArtifactID: "mina-sshd-api-core",
				},
				{
					GroupID:    "org.jenkins-ci.plugins",
					ArtifactID: "cloudbees-folder",
				},
				{
					GroupID:    "io.jenkins",
					ArtifactID: "configuration-as-code",
				},
				{
					GroupID:    "io.jenkins.configuration-as-code",
					ArtifactID: "test-harness",
				},
			},
		},
	}

	for _, tt := range testdata {
		doc := etree.NewDocument()
		if err := doc.ReadFromFile(tt.file); err != nil {
			require.NoError(t, err)
		}
		t.Run(tt.name, func(t *testing.T) {
			gotDependencies := getDependenciesFromPom(doc)
			assert.Equal(t, tt.expectedDependencies, gotDependencies)
		})
	}
}

func TestGetDependencyManagementsFromPom(t *testing.T) {
	testdata := []struct {
		name                 string
		file                 string
		expectedDependencies []dependency
	}{
		{
			name: "Test 1",
			file: "testdata/default/pom.xml",
			expectedDependencies: []dependency{
				{
					GroupID:    "io.jenkins.tools.bom",
					ArtifactID: "bom-2.346.x",
					Version:    "1508.v4b_d09ff0e893",
				},
			},
		},
	}

	for _, tt := range testdata {
		doc := etree.NewDocument()
		if err := doc.ReadFromFile(tt.file); err != nil {
			require.NoError(t, err)
		}
		t.Run(tt.name, func(t *testing.T) {
			gotDependencies := getDependencyManagementsFromPom(doc)
			assert.Equal(t, tt.expectedDependencies, gotDependencies)
		})
	}
}

func TestGetMavenRepositoriesURL(t *testing.T) {
	testdata := []struct {
		name                 string
		file                 string
		expectedRepositories []string
	}{
		{
			name: "Test 1",
			file: "testdata/default/pom.xml",
			expectedRepositories: []string{
				"https://repo.jenkins-ci.org/public/",
				"https://mirror.example.com/maven2",
			},
		},
		{
			name: "Test 2",
			file: "testdata/exclude/pom.xml",
			expectedRepositories: []string{
				"https://foo:bar@mirror.example.com/maven",
				"http://bar-repo.example.com/maven",
			},
		},
		{
			name: "Test 3",
			file: "testdata/jenkins-datadog-plugin/pom.xml",
			expectedRepositories: []string{
				"http://repo.jenkins-ci.org/public/",
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			doc := etree.NewDocument()
			if err := doc.ReadFromFile(tt.file); err != nil {
				require.NoError(t, err)
			}
			gotRepositories := getMavenRepositoriesURL(tt.file, doc)
			assert.Equal(t, tt.expectedRepositories, gotRepositories)
		})
	}
}
