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

func (m Maven) discoverParentPomDependencyManifests() ([]config.Spec, error) {

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
		parentPom := getParentFromPom(doc)

		// No need to update Version if it's not specified
		if len(parentPom.Version) == 0 {
			continue
		}

		// Retrieve dependencies

		containsVariableRegex, err := regexp.Compile(`.*\$\{.*\}.*`)

		if err != nil {
			logrus.Errorln(err)
			continue
		}

		// Test if current version contains a variable, and skip the depend if it's the case
		isContainsVariable := containsVariableRegex.Match([]byte(parentPom.Version))

		if err != nil {
			logrus.Errorln(err)
		}

		if isContainsVariable {
			logrus.Printf("Skipping parent pom as it relies on the property %q", parentPom.Version)
			continue
		}

		manifestName := fmt.Sprintf(
			"Bump Maven parent Pom %s/%s",
			parentPom.GroupID,
			parentPom.ArtifactID)

		artifactFullName := fmt.Sprintf("%s/%s", parentPom.GroupID, parentPom.ArtifactID)

		mavenSourceSpec := maven.Spec{
			GroupID:    parentPom.GroupID,
			ArtifactID: parentPom.ArtifactID,
		}

		for _, repo := range repositories {
			mavenSourceSpec.Repositories = append(mavenSourceSpec.Repositories, repo.URL)
		}

		manifest := config.Spec{
			Name: manifestName,
			Sources: map[string]source.Config{
				artifactFullName: {
					ResourceConfig: resource.ResourceConfig{
						Name: fmt.Sprintf("Get latest Parent Pom Artifact version: %q", artifactFullName),
						Kind: "maven",
						Spec: mavenSourceSpec,
					},
				},
			},
			Conditions: map[string]condition.Config{
				parentPom.GroupID: {
					DisableSourceInput: true,
					ResourceConfig: resource.ResourceConfig{
						Name: fmt.Sprintf("Ensure parent pom.xml groupId %q is specified", parentPom.GroupID),
						Kind: "xml",
						Spec: xml.Spec{
							File:  relativePomFile,
							Path:  "/project/parent/groupId",
							Value: parentPom.GroupID,
						},
					},
				},
				parentPom.ArtifactID: {
					DisableSourceInput: true,
					ResourceConfig: resource.ResourceConfig{
						Name: fmt.Sprintf("Ensure parent artifactId %q is specified", parentPom.ArtifactID),
						Kind: "xml",
						Spec: xml.Spec{
							File:  relativePomFile,
							Path:  "/project/parent/artifactId",
							Value: parentPom.ArtifactID,
						},
					},
				},
			},
			Targets: map[string]target.Config{
				artifactFullName: {
					SourceID: artifactFullName,
					ResourceConfig: resource.ResourceConfig{
						Name: fmt.Sprintf("Bump parent pom version for %q", artifactFullName),
						Kind: "xml",
						Spec: xml.Spec{
							File: relativePomFile,
							Path: "/project/parent/version",
						},
					},
				},
			},
		}
		manifests = append(manifests, manifest)

	}

	return manifests, nil
}
