package maven

import (
	"fmt"
	"path"
	"path/filepath"

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

		properties := getPropertiesFromPom(doc)

		for i, dependency := range dependencies {

			targetXMLPath := fmt.Sprintf("%s/dependencies/dependency[%d]/version", dependencyPathPrefix, i+1)

			if m.spec.Vulnerability != nil {
				// A security update bumps the property, defined in the same pom.xml, holding the version
				if name, ok := propertyName(dependency.Version); ok {
					if value, found := properties[name]; found {
						dependency.Version = value
						targetXMLPath = "/project/properties/" + name
					}
				}
			}

			if isPropertyVersion(dependency.Version) {
				if m.spec.Vulnerability != nil {
					// A user expects every dependency to be checked for vulnerabilities
					log.Warningf("Skipping dependency %q in %q as its vulnerabilities can't be checked: property %q isn't defined in this pom.xml", dependency.ArtifactID, relativePomFile, dependency.Version)
				} else {
					log.Printf("Skipping dependency %q in %q as it relies on property %q", dependency.ArtifactID, relativePomFile, dependency.Version)
				}
				continue
			}

			// A version range would be replaced by a single version
			if isVersionRange(dependency.Version) {
				log.Printf("Skipping dependency %q in %q as it relies on the version range %q", dependency.ArtifactID, relativePomFile, dependency.Version)
				continue
			}

			// No need to update Version if it's not specified
			if len(dependency.Version) == 0 {
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

			// OSV indexes the artifacts a BOM manages, not the BOM itself
			if m.spec.Vulnerability != nil && dependency.isBOM() {
				log.Warningf("Skipping BOM %q in %q as OSV doesn't know about the versions it manages", dependency.ArtifactID, relativePomFile)
				continue
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
				TargetXMLPath:            targetXMLPath,
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
