package maven

import (
	"fmt"
	"path/filepath"

	"regexp"

	"github.com/beevik/etree"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/plugins/resources/maven"
	"github.com/updatecli/updatecli/pkg/plugins/resources/xml"
)

func (m Maven) discoverDependenciesManifests() ([]config.Spec, error) {

	var manifests []config.Spec

	foundPomFiles, err := searchPomFiles(
		m.rootDir,
		pomFileName)

	if err != nil {
		return nil, err
	}

	for _, pomFile := range foundPomFiles {

		relativePomFile, err := filepath.Rel(m.rootDir, pomFile)
		if err != nil {
			// Let's try the next pom.xml if one fail
			logrus.Errorln(err)
			continue
		}

		// Test if the ignore rule based on path is respected
		if len(m.spec.Ignore) > 0 && m.spec.Ignore.isMatchingIgnoreRule(m.rootDir, relativePomFile) {
			logrus.Debugf("Ignoring pom.xml %q as not matching rule(s)\n",
				pomFile)
			continue
		}

		// Test if the only rule based on path is respected
		if len(m.spec.Only) > 0 && !m.spec.Only.isMatchingOnlyRule(m.rootDir, relativePomFile) {
			logrus.Debugf("Ignoring pom.xml %q as not matching rule(s)\n",
				pomFile)
			continue
		}

		doc := etree.NewDocument()
		if err := doc.ReadFromFile(pomFile); err != nil {
			logrus.Errorln(err)
			continue
		}

		// Retrieve repositories from pom.xml
		repositories := getRepositoriesFromPom(doc)

		// Retrieve dependencies

		dependencies := getDependenciesFromPom(doc)

		if len(dependencies) == 0 {
			logrus.Debugf("no maven dependencies found in %q\n", pomFile)
			continue
		}

		containsVariableRegex, err := regexp.Compile(`.*\$\{.*\}.*`)

		if err != nil {
			logrus.Errorln(err)
			continue
		}

		for i, dependency := range dependencies {

			// Test if current version contains a variable, and skip the depend if it's the case
			isContainsVariable := containsVariableRegex.Match([]byte(dependency.Version))

			if err != nil {
				logrus.Errorln(err)
			}

			if isContainsVariable {
				logrus.Printf("Skipping dependency as it relies on property %q", dependency.Version)
				continue
			}

			// No need to update Version if it's not specified
			if len(dependencies[i].Version) == 0 {
				continue
			}

			manifestName := fmt.Sprintf(
				"Bump Maven dependency %s/%s",
				dependency.GroupID,
				dependency.ArtifactID)

			artifactFullName := fmt.Sprintf("%s/%s", dependency.GroupID, dependency.ArtifactID)

			mavenSourceSpec := maven.Spec{
				GroupID:    dependency.GroupID,
				ArtifactID: dependency.ArtifactID,
			}

			for _, repo := range repositories {
				mavenSourceSpec.Repositories = append(mavenSourceSpec.Repositories, repo.URL)
			}

			manifest := config.Spec{
				Name: manifestName,
				Sources: map[string]source.Config{
					artifactFullName: {
						ResourceConfig: resource.ResourceConfig{
							Name: fmt.Sprintf("Get latest Maven Artifact version: %q", artifactFullName),
							Kind: "maven",
							Spec: mavenSourceSpec,
						},
					},
				},
				Conditions: map[string]condition.Config{
					dependency.GroupID: {
						DisableSourceInput: true,
						ResourceConfig: resource.ResourceConfig{
							Name: fmt.Sprintf("Ensure dependency groupId %q is specified", dependency.GroupID),
							Kind: "xml",
							Spec: xml.Spec{
								File:  relativePomFile,
								Path:  fmt.Sprintf("/project/dependencies/dependency[%d]/groupId", i+1),
								Value: dependency.GroupID,
							},
						},
					},
					dependency.ArtifactID: {
						DisableSourceInput: true,
						ResourceConfig: resource.ResourceConfig{
							Name: fmt.Sprintf("Ensure dependency artifactId %q is specified", dependency.ArtifactID),
							Kind: "xml",
							Spec: xml.Spec{
								File:  relativePomFile,
								Path:  fmt.Sprintf("/project/dependencies/dependency[%d]/artifactId", i+1),
								Value: dependency.ArtifactID,
							},
						},
					},
				},
				Targets: map[string]target.Config{
					artifactFullName: {
						SourceID: artifactFullName,
						ResourceConfig: resource.ResourceConfig{
							Name: fmt.Sprintf("Bump dependency version for %q", artifactFullName),
							Kind: "xml",
							Spec: xml.Spec{
								File: relativePomFile,
								Path: fmt.Sprintf("/project/dependencies/dependency[%d]/version", i+1),
							},
						},
					},
				},
			}
			manifests = append(manifests, manifest)

		}
	}

	return manifests, nil
}
