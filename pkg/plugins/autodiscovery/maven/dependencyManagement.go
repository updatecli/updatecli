package maven

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"text/template"

	"regexp"

	"github.com/beevik/etree"
	"github.com/sirupsen/logrus"
)

func (m Maven) discoverDependencyManagementsManifests() ([][]byte, error) {

	var manifests [][]byte

	searchFromDir := m.rootDir
	// If the spec.RootDir is an absolute path, then it as already been set
	// correctly in the New function.
	if m.spec.RootDir != "" && !path.IsAbs(m.spec.RootDir) {
		searchFromDir = filepath.Join(m.rootDir, m.spec.RootDir)
	}

	foundPomFiles, err := searchPomFiles(
		searchFromDir,
		pomFileName)

	if err != nil {
		return nil, err
	}

	for _, pomFile := range foundPomFiles {
		relativePomFile, err := filepath.Rel(m.rootDir, pomFile)
		logrus.Debugf("parsing file %q", pomFile)
		if err != nil {
			// Let's try the next pom.xml if one fail
			logrus.Debugln(err)
			continue
		}

		doc := etree.NewDocument()
		if err := doc.ReadFromFile(pomFile); err != nil {
			logrus.Debugln(err)
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
			logrus.Debugln(err)
			continue
		}

		for i, dependency := range dependencies {

			// Test if current version contains a variable, and skip the depend if it's the case
			isContainsVariable := containsVariableRegex.Match([]byte(dependency.Version))

			if err != nil {
				logrus.Debugln(err)
				continue
			}

			if isContainsVariable {
				logrus.Printf("Skipping dependencyManagement as it relies on property %q", dependency.Version)
				continue
			}

			// No need to update Version if it's not specified
			if len(dependencies[i].Version) == 0 {
				continue
			}

			artifactFullName := fmt.Sprintf("%s/%s", dependency.GroupID, dependency.ArtifactID)

			repos := []string{}
			for _, repo := range repositories {
				repos = append(repos, repo.URL)
			}

			sourceVersionFilterKind := m.versionFilter.Kind
			sourceVersionFilterPattern := m.versionFilter.Pattern
			if !m.spec.VersionFilter.IsZero() {
				sourceVersionFilterKind = m.versionFilter.Kind
				sourceVersionFilterPattern, err = m.versionFilter.GreaterThanPattern(dependency.Version)
				if err != nil {
					logrus.Debugf("building version filter pattern: %s", err)
					sourceVersionFilterPattern = "*"
				}
			}

			if len(m.spec.Ignore) > 0 {
				if m.spec.Ignore.isMatchingRules(m.rootDir, relativePomFile, dependency.GroupID, dependency.ArtifactID, dependency.Version) {
					logrus.Debugf("Ignoring %s.%s from %q, as matching ignore rule(s)\n", dependency.GroupID, dependency.ArtifactID, relativePomFile)
					continue
				}
			}

			if len(m.spec.Only) > 0 {
				if !m.spec.Only.isMatchingRules(m.rootDir, relativePomFile, dependency.GroupID, dependency.ArtifactID, dependency.Version) {
					logrus.Debugf("Ignoring package %s.%s from %q, as not matching only rule(s)\n", dependency.GroupID, dependency.ArtifactID, relativePomFile)
					continue
				}
			}

			tmpl, err := template.New("manifest").Parse(manifestTemplate)
			if err != nil {
				logrus.Debugln(err)
				continue
			}

			params := struct {
				ManifestName               string
				ConditionID                string
				ConditionGroupID           string
				ConditionGroupIDName       string
				ConditionGroupIDPath       string
				ConditionGroupIDValue      string
				ConditionArtifactID        string
				ConditionArtifactIDName    string
				ConditionArtifactIDPath    string
				ConditionArtifactIDValue   string
				SourceID                   string
				SourceName                 string
				SourceKind                 string
				SourceGroupID              string
				SourceArtifactID           string
				SourceRepositories         []string
				SourceVersionFilterKind    string
				SourceVersionFilterPattern string
				TargetID                   string
				TargetName                 string
				TargetXMLPath              string
				File                       string
				ScmID                      string
			}{
				ManifestName:               fmt.Sprintf("Bump Maven dependencyManagement %s/%s", dependency.GroupID, dependency.ArtifactID),
				ConditionID:                artifactFullName,
				ConditionGroupID:           "groupid",
				ConditionGroupIDName:       fmt.Sprintf("Ensure dependencyManagement groupId %q is specified", dependency.GroupID),
				ConditionGroupIDValue:      dependency.GroupID,
				ConditionGroupIDPath:       fmt.Sprintf("/project/dependencyManagement/dependencies/dependency[%d]/groupId", i+1),
				ConditionArtifactIDName:    fmt.Sprintf("Ensure dependencyManagement artifactId %q is specified", dependency.ArtifactID),
				ConditionArtifactID:        "artifactid",
				ConditionArtifactIDPath:    fmt.Sprintf("/project/dependencyManagement/dependencies/dependency[%d]/artifactId", i+1),
				ConditionArtifactIDValue:   dependency.ArtifactID,
				SourceID:                   artifactFullName,
				SourceName:                 fmt.Sprintf("Get latest Maven Artifact version %q", artifactFullName),
				SourceKind:                 "maven",
				SourceGroupID:              dependency.GroupID,
				SourceArtifactID:           dependency.ArtifactID,
				SourceRepositories:         repos,
				SourceVersionFilterKind:    sourceVersionFilterKind,
				SourceVersionFilterPattern: sourceVersionFilterPattern,
				TargetID:                   artifactFullName,
				TargetName:                 fmt.Sprintf("Bump dependencyManagement version for %q", artifactFullName),
				TargetXMLPath:              fmt.Sprintf("/project/dependencyManagement/dependencies/dependency[%d]/version", i+1),
				File:                       relativePomFile,
				ScmID:                      m.scmID,
			}

			manifest := bytes.Buffer{}
			if err := tmpl.Execute(&manifest, params); err != nil {
				logrus.Debugln(err)
				continue
			}

			manifests = append(manifests, manifest.Bytes())
		}
	}

	return manifests, nil
}
