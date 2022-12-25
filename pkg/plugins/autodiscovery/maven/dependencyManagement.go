package maven

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"regexp"

	"github.com/beevik/etree"
	"github.com/sirupsen/logrus"
)

func (m Maven) discoverDependencyManagementsManifests() ([][]byte, error) {

	var manifests [][]byte

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

		dependencies := getDependencyManagementsFromPom(doc)

		if len(dependencies) == 0 {
			logrus.Debugf("no Maven dependencyManagements found in %q\n", pomFile)
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
				logrus.Printf("Skipping dependencyManagement as it relies on property %q", dependency.Version)
				continue
			}

			// No need to update Version if it's not specified
			if len(dependencies[i].Version) == 0 {
				continue
			}

			//manifest := config.Spec{
			//	Name: manifestName,
			//	Sources: map[string]source.Config{
			//		artifactFullName: {
			//			ResourceConfig: resource.ResourceConfig{
			//				Name: fmt.Sprintf("Get latest Maven Artifact version: %q", artifactFullName),
			//				Kind: "maven",
			//				Spec: mavenSourceSpec,
			//			},
			//		},
			//	},
			//	Conditions: map[string]condition.Config{
			//		dependency.GroupID: {
			//			DisableSourceInput: true,
			//			ResourceConfig: resource.ResourceConfig{
			//				Name: fmt.Sprintf("Ensure dependencyManagement groupId %q is specified", dependency.GroupID),
			//				Kind: "xml",
			//				Spec: xml.Spec{
			//					File:  relativePomFile,
			//					Path:  fmt.Sprintf("/project/dependencyManagement/dependencies/dependency[%d]/groupId", i+1),
			//					Value: dependency.GroupID,
			//				},
			//			},
			//		},
			//		dependency.ArtifactID: {
			//			DisableSourceInput: true,
			//			ResourceConfig: resource.ResourceConfig{
			//				Name: fmt.Sprintf("Ensure dependencyManagement artifactId %q is specified", dependency.ArtifactID),
			//				Kind: "xml",
			//				Spec: xml.Spec{
			//					File:  relativePomFile,
			//					Path:  fmt.Sprintf("/project/dependencyManagement/dependencies/dependency[%d]/artifactId", i+1),
			//					Value: dependency.ArtifactID,
			//				},
			//			},
			//		},
			//	},
			//	Targets: map[string]target.Config{
			//		artifactFullName: {
			//			SourceID: artifactFullName,
			//			ResourceConfig: resource.ResourceConfig{
			//				Name: fmt.Sprintf("Bump dependencyManagement version for %q", artifactFullName),
			//				Kind: "xml",
			//				Spec: xml.Spec{
			//					File: relativePomFile,
			//					Path: fmt.Sprintf("/project/dependencyManagement/dependencies/dependency[%d]/version", i+1),
			//				},
			//			},
			//		},
			//	},
			//}
			// Set scmID if defined
			artifactFullName := fmt.Sprintf("%s/%s", dependency.GroupID, dependency.ArtifactID)

			repos := []string{}
			for _, repo := range repositories {
				repos = append(repos, repo.URL)
			}

			tmpl, err := template.New("manifest").Parse(manifestTemplate)
			if err != nil {
				logrus.Errorln(err)
				continue
			}

			params := struct {
				ManifestName             string
				ConditionID              string
				ConditionGroupID         string
				ConditionGroupIDName     string
				ConditionGroupIDPath     string
				ConditionGroupIDValue    string
				ConditionArtifactID      string
				ConditionArtifactIDName  string
				ConditionArtifactIDPath  string
				ConditionArtifactIDValue string
				SourceID                 string
				SourceName               string
				SourceKind               string
				SourceGroupID            string
				SourceArtifactID         string
				SourceRepositories       []string
				TargetID                 string
				TargetName               string
				TargetXMLPath            string
				File                     string
				ScmID                    string
			}{
				ManifestName:             fmt.Sprintf("Bump Maven dependencyManagement %s/%s", dependency.GroupID, dependency.ArtifactID),
				ConditionID:              artifactFullName,
				ConditionGroupID:         dependency.GroupID,
				ConditionGroupIDName:     fmt.Sprintf("Ensure dependencyManagement groupId %q is specified", dependency.GroupID),
				ConditionGroupIDValue:    dependency.GroupID,
				ConditionGroupIDPath:     fmt.Sprintf("/project/dependencyManagement/dependencies/dependency[%d]/groupId", i+1),
				ConditionArtifactIDName:  fmt.Sprintf("Ensure dependencyManagement artifactId %q is specified", dependency.ArtifactID),
				ConditionArtifactID:      dependency.ArtifactID,
				ConditionArtifactIDPath:  fmt.Sprintf("/project/dependencyManagement/dependencies/dependency[%d]/artifactId", i+1),
				ConditionArtifactIDValue: dependency.ArtifactID,
				SourceID:                 artifactFullName,
				SourceName:               fmt.Sprintf("Get latest Maven Artifact version %q", artifactFullName),
				SourceKind:               "maven",
				SourceGroupID:            dependency.GroupID,
				SourceArtifactID:         dependency.ArtifactID,
				SourceRepositories:       repos,
				TargetID:                 artifactFullName,
				TargetName:               fmt.Sprintf("Bump dependencyManagement version for %q", artifactFullName),
				TargetXMLPath:            fmt.Sprintf("/project/dependencyManagement/dependencies/dependency[%d]/version", i+1),
				File:                     relativePomFile,
				ScmID:                    m.scmID,
			}

			manifest := bytes.Buffer{}
			if err := tmpl.Execute(&manifest, params); err != nil {
				return nil, err
			}

			manifests = append(manifests, manifest.Bytes())
		}
	}

	return manifests, nil
}
