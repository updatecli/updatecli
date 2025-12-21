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
		logrus.Debugf("parsing file %q", relativePomFile)
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

		mavenRepositories := getMavenRepositoriesURL(pomFile, doc)

		parentPom := getParentFromPom(doc)

		// No need to update Version if it's not specified
		if len(parentPom.Version) == 0 {
			continue
		}

		// Retrieve dependencies

		containsVariableRegex, err := regexp.Compile(`.*\$\{.*\}.*`)

		if err != nil {
			logrus.Debugln(err)
			continue
		}

		// Test if current version contains a variable, and skip the depend if it's the case
		isContainVariable := containsVariableRegex.Match([]byte(parentPom.Version))

		if isContainVariable {
			logrus.Printf("Skipping parent pom %q in %q as it relies on the property %q", parentPom.ArtifactID, relativePomFile, parentPom.Version)
			continue
		}

		artifactFullName := fmt.Sprintf("%s/%s", parentPom.GroupID, parentPom.ArtifactID)

		sourceVersionFilterKind := m.versionFilter.Kind
		sourceVersionFilterPattern := m.versionFilter.Pattern
		sourceVersionFilterRegex := m.versionFilter.Regex
		if !m.spec.VersionFilter.IsZero() {
			sourceVersionFilterKind = m.versionFilter.Kind
			sourceVersionFilterPattern, err = m.versionFilter.GreaterThanPattern(parentPom.Version)
			sourceVersionFilterRegex = m.versionFilter.Regex
			if err != nil {
				logrus.Debugf("building version filter pattern: %s", err)
				sourceVersionFilterPattern = "*"
			}
		}

		if len(m.spec.Ignore) > 0 {
			if m.spec.Ignore.isMatchingRules(m.rootDir, relativePomFile, parentPom.GroupID, parentPom.ArtifactID, parentPom.Version) {
				logrus.Debugf("Ignoring %s.%s from %q, as matching ignore rule(s)\n", parentPom.GroupID, parentPom.ArtifactID, relativePomFile)
				continue
			}
		}

		if len(m.spec.Only) > 0 {
			if !m.spec.Only.isMatchingRules(m.rootDir, relativePomFile, parentPom.GroupID, parentPom.ArtifactID, parentPom.Version) {
				logrus.Debugf("Ignoring package %s.%s from %q, as not matching only rule(s)\n", parentPom.GroupID, parentPom.ArtifactID, relativePomFile)
				continue
			}
		}

		tmpl, err := template.New("manifest").Parse(manifestTemplate)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		params := struct {
			ActionID                   string
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
			SourceVersionFilterRegex   string
			TargetID                   string
			TargetName                 string
			TargetXMLPath              string
			File                       string
			ScmID                      string
		}{
			ActionID:                   m.actionID,
			ManifestName:               fmt.Sprintf("Bump Maven parent Pom %s/%s", parentPom.GroupID, parentPom.ArtifactID),
			ConditionID:                artifactFullName,
			ConditionGroupID:           "groupid",
			ConditionGroupIDName:       fmt.Sprintf("Ensure parent pom.xml groupId %q is specified", parentPom.GroupID),
			ConditionGroupIDPath:       "/project/parent/groupId",
			ConditionGroupIDValue:      parentPom.GroupID,
			ConditionArtifactID:        "artifactid",
			ConditionArtifactIDName:    fmt.Sprintf("Ensure parent artifactId %q is specified", parentPom.ArtifactID),
			ConditionArtifactIDPath:    "/project/parent/artifactId",
			ConditionArtifactIDValue:   parentPom.ArtifactID,
			SourceID:                   artifactFullName,
			SourceName:                 fmt.Sprintf("Get latest Parent Pom Artifact version %q", artifactFullName),
			SourceKind:                 "maven",
			SourceGroupID:              parentPom.GroupID,
			SourceArtifactID:           parentPom.ArtifactID,
			SourceRepositories:         mavenRepositories,
			SourceVersionFilterKind:    sourceVersionFilterKind,
			SourceVersionFilterPattern: sourceVersionFilterPattern,
			SourceVersionFilterRegex:   sourceVersionFilterRegex,
			TargetID:                   artifactFullName,
			TargetName:                 fmt.Sprintf("deps(maven): update %q to {{ source %q }}", artifactFullName, artifactFullName),
			TargetXMLPath:              "/project/parent/version",
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

	return manifests, nil
}
