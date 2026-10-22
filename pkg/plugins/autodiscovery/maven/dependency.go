package maven

import (
	"fmt"
	"path"
	"path/filepath"

	"regexp"

	"github.com/beevik/etree"
	"github.com/updatecli/updatecli/pkg/core/log"
)

// discoverDependencyManifests discovers manifests for Maven dependencies
func (m Maven) discoverDependencyManifests(kind string) ([][]byte, error) {

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
		if err != nil {
			// Let's try the next pom.xml if one fail
			log.Debugln(err)
			continue
		}

		log.Debugf("parsing file %q", relativePomFile)

		doc := etree.NewDocument()
		if err := doc.ReadFromFile(pomFile); err != nil {
			log.Debugln(err)
			continue
		}

		mavenRepositories := getMavenRepositoriesURL(pomFile, doc)

		// Retrieve dependencies
		var dependencies []dependency
		var dependencyKind string
		var dependencyPathPrefix string

		switch kind {
		case "dependency":
			dependencies = getDependenciesFromPom(doc)
			dependencyKind = "dependencies"
			dependencyPathPrefix = "/project"
		case "dependencyManagement":
			dependencies = getDependencyManagementsFromPom(doc)
			dependencyKind = "dependencyManagements"
			dependencyPathPrefix = "/project/dependencyManagement"
		}

		if len(dependencies) == 0 {
			log.Debugf("no maven %s found in %q\n", dependencyKind, relativePomFile)
			continue
		}

		containsVariableRegex, err := regexp.Compile(`.*\$\{.*\}.*`)

		if err != nil {
			log.Errorln(err)
			continue
		}

		for i, dependency := range dependencies {

			// Test if current version contains a variable, and skip the depend if it's the case
			isContainVariable := containsVariableRegex.Match([]byte(dependency.Version))

			if err != nil {
				log.Debugln(err)
				continue
			}

			if isContainVariable {
				log.Printf("Skipping dependency %q in %q as it relies on property %q", dependency.ArtifactID, relativePomFile, dependency.Version)
				continue
			}

			// No current version can be checked against the OSV database
			if m.spec.Vulnerability != nil && isVersionRange(dependency.Version) {
				log.Printf("Skipping dependency %q in %q as it relies on the version range %q", dependency.ArtifactID, relativePomFile, dependency.Version)
				continue
			}

			// No need to update Version if it's not specified
			if len(dependencies[i].Version) == 0 {
				continue
			}

			artifactFullName := fmt.Sprintf("%s/%s", dependency.GroupID, dependency.ArtifactID)

			if len(m.spec.Ignore) > 0 {
				if m.spec.Ignore.isMatchingRules(m.rootDir, relativePomFile, dependency.GroupID, dependency.ArtifactID, dependency.Version) {
					log.Debugf("Ignoring %s.%s from %q, as matching ignore rule(s)\n", dependency.GroupID, dependency.ArtifactID, relativePomFile)
					continue
				}
			}

			if len(m.spec.Only) > 0 {
				if !m.spec.Only.isMatchingRules(m.rootDir, relativePomFile, dependency.GroupID, dependency.ArtifactID, dependency.Version) {
					log.Debugf("Ignoring package %s.%s from %q, as not matching only rule(s)\n", dependency.GroupID, dependency.ArtifactID, relativePomFile)
					continue
				}
			}

			params := manifestTemplateParams{
				ActionID:                 m.actionID,
				ManifestName:             fmt.Sprintf("Bump Maven %s %s", kind, artifactFullName),
				ConditionID:              artifactFullName,
				ConditionGroupID:         "groupid",
				ConditionGroupIDName:     fmt.Sprintf("Ensure %s groupId %q is specified", kind, dependency.GroupID),
				ConditionGroupIDPath:     fmt.Sprintf("%s/dependencies/dependency[%d]/groupId", dependencyPathPrefix, i+1),
				ConditionGroupIDValue:    dependency.GroupID,
				ConditionArtifactID:      "artifactid",
				ConditionArtifactIDName:  fmt.Sprintf("Ensure %s artifactId %q is specified", kind, dependency.ArtifactID),
				ConditionArtifactIDPath:  fmt.Sprintf("%s/dependencies/dependency[%d]/artifactId", dependencyPathPrefix, i+1),
				ConditionArtifactIDValue: dependency.ArtifactID,
				SourceID:                 artifactFullName,
				SourceName:               fmt.Sprintf("Get latest Maven Artifact version %q", artifactFullName),
				SourceKind:               "maven",
				SourceGroupID:            dependency.GroupID,
				SourceArtifactID:         dependency.ArtifactID,
				SourceRepositories:       mavenRepositories,
				TargetID:                 artifactFullName,
				TargetName:               fmt.Sprintf("deps(maven): update %q to {{ source %q }}", artifactFullName, artifactFullName),
				TargetXMLPath:            fmt.Sprintf("%s/dependencies/dependency[%d]/version", dependencyPathPrefix, i+1),
				File:                     relativePomFile,
				ScmID:                    m.scmID,
			}

			manifest, err := m.renderManifest(params, dependency.Version)
			if err != nil {
				log.Errorln(err)
				continue
			}

			manifests = append(manifests, manifest)
		}
	}

	return manifests, nil
}
