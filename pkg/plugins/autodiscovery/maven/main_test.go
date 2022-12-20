package maven

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	discoveryConfig "github.com/updatecli/updatecli/pkg/core/pipeline/autodiscovery/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/plugins/resources/maven"
	"github.com/updatecli/updatecli/pkg/plugins/resources/xml"
)

func TestDiscoverManifests(t *testing.T) {

	testdata := []struct {
		name              string
		rootDir           string
		expectedPipelines []config.Spec
	}{
		{
			name:    "Scenario 1",
			rootDir: "testdata",
			expectedPipelines: []config.Spec{
				{

					Name: "Bump Maven dependency com.jcraft/jsch",
					Sources: map[string]source.Config{
						"com.jcraft/jsch": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get latest Maven Artifact version: \"com.jcraft/jsch\"",
								Kind: "maven",
								Spec: maven.Spec{
									Repositories: []string{
										"https://repo.jenkins-ci.org/public/",
									},
									GroupID:    "com.jcraft",
									ArtifactID: "jsch",
								},
							},
						},
					},
					Conditions: map[string]condition.Config{

						"com.jcraft": {
							DisableSourceInput: true,
							ResourceConfig: resource.ResourceConfig{
								Name: "Ensure dependency groupId \"com.jcraft\" is specified",
								Kind: "xml",
								Spec: xml.Spec{
									File:  "pom.xml",
									Path:  "/project/dependencies/dependency[1]/groupId",
									Value: "com.jcraft",
								},
							},
						},
						"jsch": {
							DisableSourceInput: true,
							ResourceConfig: resource.ResourceConfig{
								Name: "Ensure dependency artifactId \"jsch\" is specified",
								Kind: "xml",
								Spec: xml.Spec{
									File:  "pom.xml",
									Path:  "/project/dependencies/dependency[1]/artifactId",
									Value: "jsch",
								},
							},
						},
					},
					Targets: map[string]target.Config{

						"com.jcraft/jsch": {
							SourceID: "com.jcraft/jsch",
							ResourceConfig: resource.ResourceConfig{
								Name: "Bump dependency version for \"com.jcraft/jsch\"",
								Kind: "xml",
								Spec: xml.Spec{
									File: "pom.xml",
									Path: "/project/dependencies/dependency[1]/version",
								},
							},
						},
					},
				},
				{

					Name: "Bump Maven dependencyManagement io.jenkins.tools.bom/bom-2.346.x",
					Sources: map[string]source.Config{
						"io.jenkins.tools.bom/bom-2.346.x": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get latest Maven Artifact version: \"io.jenkins.tools.bom/bom-2.346.x\"",
								Kind: "maven",
								Spec: maven.Spec{
									Repositories: []string{
										"https://repo.jenkins-ci.org/public/",
									},
									GroupID:    "io.jenkins.tools.bom",
									ArtifactID: "bom-2.346.x",
								},
							},
						},
					},
					Conditions: map[string]condition.Config{

						"io.jenkins.tools.bom": {
							DisableSourceInput: true,
							ResourceConfig: resource.ResourceConfig{
								Name: "Ensure dependencyManagement groupId \"io.jenkins.tools.bom\" is specified",
								Kind: "xml",
								Spec: xml.Spec{
									File:  "pom.xml",
									Path:  "/project/dependencyManagement/dependencies/dependency[1]/groupId",
									Value: "io.jenkins.tools.bom",
								},
							},
						},
						"bom-2.346.x": {
							DisableSourceInput: true,
							ResourceConfig: resource.ResourceConfig{
								Name: "Ensure dependencyManagement artifactId \"bom-2.346.x\" is specified",
								Kind: "xml",
								Spec: xml.Spec{
									File:  "pom.xml",
									Path:  "/project/dependencyManagement/dependencies/dependency[1]/artifactId",
									Value: "bom-2.346.x",
								},
							},
						},
					},
					Targets: map[string]target.Config{

						"io.jenkins.tools.bom/bom-2.346.x": {
							SourceID: "io.jenkins.tools.bom/bom-2.346.x",
							ResourceConfig: resource.ResourceConfig{
								Name: "Bump dependencyManagement version for \"io.jenkins.tools.bom/bom-2.346.x\"",
								Kind: "xml",
								Spec: xml.Spec{
									File: "pom.xml",
									Path: "/project/dependencyManagement/dependencies/dependency[1]/version",
								},
							},
						},
					},
				},
				{

					Name: "Bump Maven parent Pom org.jenkins-ci.plugins/plugin",
					Sources: map[string]source.Config{
						"org.jenkins-ci.plugins/plugin": {
							ResourceConfig: resource.ResourceConfig{
								Name: "Get latest Parent Pom Artifact version: \"org.jenkins-ci.plugins/plugin\"",
								Kind: "maven",
								Spec: maven.Spec{
									Repositories: []string{
										"https://repo.jenkins-ci.org/public/",
									},
									GroupID:    "org.jenkins-ci.plugins",
									ArtifactID: "plugin",
								},
							},
						},
					},
					Conditions: map[string]condition.Config{

						"org.jenkins-ci.plugins": {
							DisableSourceInput: true,
							ResourceConfig: resource.ResourceConfig{
								Name: "Ensure parent pom.xml groupId \"org.jenkins-ci.plugins\" is specified",
								Kind: "xml",
								Spec: xml.Spec{
									File:  "pom.xml",
									Path:  "/project/parent/groupId",
									Value: "org.jenkins-ci.plugins",
								},
							},
						},
						"plugin": {
							DisableSourceInput: true,
							ResourceConfig: resource.ResourceConfig{
								Name: "Ensure parent artifactId \"plugin\" is specified",
								Kind: "xml",
								Spec: xml.Spec{
									File:  "pom.xml",
									Path:  "/project/parent/artifactId",
									Value: "plugin",
								},
							},
						},
					},
					Targets: map[string]target.Config{

						"org.jenkins-ci.plugins/plugin": {
							SourceID: "org.jenkins-ci.plugins/plugin",
							ResourceConfig: resource.ResourceConfig{
								Name: "Bump parent pom version for \"org.jenkins-ci.plugins/plugin\"",
								Kind: "xml",
								Spec: xml.Spec{
									File: "pom.xml",
									Path: "/project/parent/version",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			maven, err := New(
				Spec{
					RootDir: tt.rootDir,
				}, "", "")

			require.NoError(t, err)

			pipelines, err := maven.DiscoverManifests(discoveryConfig.Input{})

			require.NoError(t, err)
			assert.Equal(t, tt.expectedPipelines, pipelines)
		})
	}

}
