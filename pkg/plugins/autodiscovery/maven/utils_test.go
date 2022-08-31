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
			file: "testdata/pom.xml",
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
			file: "testdata/pom.xml",
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
			file: "testdata/pom.xml",
			expectedDependencies: []dependency{
				{
					GroupID:    "com.jcraft",
					ArtifactID: "jsch",
					Version:    "0.1.55",
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
			file: "testdata/pom.xml",
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
