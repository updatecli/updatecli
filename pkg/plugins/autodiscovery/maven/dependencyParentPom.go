package maven

import (
	"fmt"
	"path"
	"path/filepath"

	"regexp"

	"github.com/beevik/etree"
	"github.com/updatecli/updatecli/pkg/core/log"
)

func (m Maven) discoverParentPomDependencyManifests() ([][]byte, error) {

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
		log.Debugf("parsing file %q", relativePomFile)
		if err != nil {
			// Let's try the next pom.xml if one fail
			log.Debugln(err)
			continue
		}

		doc := etree.NewDocument()
		if err := doc.ReadFromFile(pomFile); err != nil {
			log.Debugln(err)
			continue
		}

		mavenRepositories := getMavenRepositoriesURL(pomFile, doc)

		parentPom := getParentFromPom(doc)

		// No need to update Version if it's not specified
		if len(parentPom.Version) == 0 {
			continue
		}

		// Retrieve dependencies

		containsVariableRegex, err := regexp.Compile(`.*\$\{.*\}.*`)

		if err != nil {
			log.Debugln(err)
			continue
		}

		// Test if current version contains a variable, and skip the depend if it's the case
		isContainVariable := containsVariableRegex.Match([]byte(parentPom.Version))

		if isContainVariable {
			log.Printf("Skipping parent pom %q in %q as it relies on the property %q", parentPom.ArtifactID, relativePomFile, parentPom.Version)
			continue
		}

		// No current version can be checked against the OSV database
		if m.spec.Vulnerability != nil && isVersionRange(parentPom.Version) {
			log.Printf("Skipping parent pom %q in %q as it relies on the version range %q", parentPom.ArtifactID, relativePomFile, parentPom.Version)
			continue
		}

		artifactFullName := fmt.Sprintf("%s/%s", parentPom.GroupID, parentPom.ArtifactID)

		if len(m.spec.Ignore) > 0 {
			if m.spec.Ignore.isMatchingRules(m.rootDir, relativePomFile, parentPom.GroupID, parentPom.ArtifactID, parentPom.Version) {
				log.Debugf("Ignoring %s.%s from %q, as matching ignore rule(s)\n", parentPom.GroupID, parentPom.ArtifactID, relativePomFile)
				continue
			}
		}

		if len(m.spec.Only) > 0 {
			if !m.spec.Only.isMatchingRules(m.rootDir, relativePomFile, parentPom.GroupID, parentPom.ArtifactID, parentPom.Version) {
				log.Debugf("Ignoring package %s.%s from %q, as not matching only rule(s)\n", parentPom.GroupID, parentPom.ArtifactID, relativePomFile)
				continue
			}
		}

		params := manifestTemplateParams{
			ActionID:                 m.actionID,
			ManifestName:             fmt.Sprintf("Bump Maven parent Pom %s/%s", parentPom.GroupID, parentPom.ArtifactID),
			ConditionID:              artifactFullName,
			ConditionGroupID:         "groupid",
			ConditionGroupIDName:     fmt.Sprintf("Ensure parent pom.xml groupId %q is specified", parentPom.GroupID),
			ConditionGroupIDPath:     "/project/parent/groupId",
			ConditionGroupIDValue:    parentPom.GroupID,
			ConditionArtifactID:      "artifactid",
			ConditionArtifactIDName:  fmt.Sprintf("Ensure parent artifactId %q is specified", parentPom.ArtifactID),
			ConditionArtifactIDPath:  "/project/parent/artifactId",
			ConditionArtifactIDValue: parentPom.ArtifactID,
			SourceID:                 artifactFullName,
			SourceName:               fmt.Sprintf("Get latest Parent Pom Artifact version %q", artifactFullName),
			SourceKind:               "maven",
			SourceGroupID:            parentPom.GroupID,
			SourceArtifactID:         parentPom.ArtifactID,
			SourceRepositories:       mavenRepositories,
			TargetID:                 artifactFullName,
			TargetName:               fmt.Sprintf("deps(maven): update %q to {{ source %q }}", artifactFullName, artifactFullName),
			TargetXMLPath:            "/project/parent/version",
			File:                     relativePomFile,
			ScmID:                    m.scmID,
		}

		manifest, err := m.renderManifest(params, parentPom.Version)
		if err != nil {
			log.Errorln(err)
			continue
		}

		manifests = append(manifests, manifest)

	}

	return manifests, nil
}
