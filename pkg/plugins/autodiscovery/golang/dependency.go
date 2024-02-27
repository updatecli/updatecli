package golang

import (
	"bytes"
	"os"
	"path/filepath"
	"text/template"

	"github.com/sirupsen/logrus"
)

// discoverDependencyManifests search for each go.mod file
// and then try to update both "direct" Go module and the Golang version
func (g Golang) discoverDependencyManifests() ([][]byte, error) {

	var manifests [][]byte

	foundFiles, err := searchGoModFiles(g.rootDir)

	if err != nil {
		return nil, err
	}

	for _, foundFile := range foundFiles {
		logrus.Debugf("parsing file %q", foundFile)

		relativeFoundFile, err := filepath.Rel(g.rootDir, foundFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		relativeWorkDir, err := filepath.Rel(g.rootDir, filepath.Dir(foundFile))
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		goSumFound := false
		goSumFilePath := filepath.Join(relativeWorkDir, "go.sum")
		if _, err := os.Stat(goSumFilePath); err == nil {
			goSumFound = true
		}

		// If the Go binary is available then we can run `go mod tidy` in case of the go.mod modification
		goModTidyEnabled := false
		switch isGolangInstalled() {
		case true:
			// If both go and go.sum are present, then we can run `go mod tidy` after go.mod file change
			if goSumFound {
				goModTidyEnabled = true
			}

		case false:
			if goSumFound {
				logrus.Warningf("File %q detected but not Golang so we can't run go mod tidy if %s is modified", goSumFilePath, foundFile)
			}
		}

		goVersion, goModules, err := getGoModContent(foundFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		for goModule, goModuleVersion := range goModules {
			// Test if the ignore rule based on path is respected
			if len(g.spec.Ignore) > 0 {
				if g.spec.Ignore.isMatchingRules(g.rootDir, relativeFoundFile, goVersion, goModule, goModuleVersion) {
					logrus.Debugf("Ignoring module %q from file %q, as matching ignore rule(s)\n", goModule, relativeFoundFile)
					continue
				}
			}

			// Test if the only rule based on path is respected
			if len(g.spec.Only) > 0 {
				if !g.spec.Only.isMatchingRules(g.rootDir, relativeFoundFile, goVersion, goModule, goModuleVersion) {
					logrus.Debugf("Ignoring module %q from %q, as not matching only rule(s)\n", goModule, relativeFoundFile)
					continue
				}
			}

			goModuleVersionPattern, err := g.versionFilter.GreaterThanPattern(goModuleVersion)
			if err != nil {
				logrus.Debugf("skipping golang module %q due to: %s", goModule, err)
				continue
			}

			moduleManifest, err := getGolangModuleManifest(
				relativeFoundFile,
				goModule,
				g.versionFilter.Kind,
				goModuleVersionPattern,
				g.scmID,
				relativeWorkDir,
				goModTidyEnabled)
			if err != nil {
				logrus.Debugf("skipping golang module %q module due to: %s", goModule, err)
				continue
			}

			manifests = append(manifests, moduleManifest)
		}
		// Test if the ignore rule based on path is respected
		if len(g.spec.Ignore) > 0 {
			if g.spec.Ignore.isMatchingRules(g.rootDir, relativeFoundFile, goVersion, "", "") {
				logrus.Debugf("Ignoring golang version update from file %q, as matching ignore rule(s)\n", relativeFoundFile)
				continue
			}
		}

		// Test if the only rule based on path is respected
		if len(g.spec.Only) > 0 {
			if !g.spec.Only.isMatchingRules(g.rootDir, relativeFoundFile, goVersion, "", "") {
				logrus.Debugf("Ignoring golang version update from %q, as not matching only rule(s)\n", relativeFoundFile)
				continue
			}
		}

		goVersionPattern, err := g.versionFilter.GreaterThanPattern(goVersion)
		golangVersionManifest := []byte{}
		if err != nil {
			logrus.Debugln(err)
		} else {
			golangVersionManifest, err = getGolangVersionManifest(
				relativeFoundFile,
				g.versionFilter.Kind,
				goVersionPattern, g.scmID)
			if err != nil {
				logrus.Debugln(err)
				logrus.Debugln("skipping golang version manifest due to previous error")
			}
		}
		manifests = append(manifests, golangVersionManifest)
	}

	logrus.Printf("%v manifests identified", len(manifests))

	return manifests, nil
}

func getGolangVersionManifest(filename, versionFilterKind, versionFilterPattern, scmID string) ([]byte, error) {
	tmpl, err := template.New("manifest").Parse(goManifestTemplate)
	if err != nil {
		logrus.Debugln(err)
		return nil, err
	}

	params := struct {
		GoModFile            string
		VersionFilterKind    string
		VersionFilterPattern string
		ScmID                string
	}{
		GoModFile:            filename,
		VersionFilterKind:    versionFilterKind,
		VersionFilterPattern: versionFilterPattern,
		ScmID:                scmID,
	}

	manifest := bytes.Buffer{}
	if err := tmpl.Execute(&manifest, params); err != nil {
		logrus.Debugln(err)
		return nil, err
	}
	return manifest.Bytes(), nil
}

func getGolangModuleManifest(filename, module, versionFilterKind, versionFilterPattern, scmID, workdir string, goModTidy bool) ([]byte, error) {

	tmpl, err := template.New("manifest").Parse(goModuleManifestTemplate)
	if err != nil {
		logrus.Debugln(err)
		return nil, err
	}

	params := struct {
		GoModFile            string
		Module               string
		VersionFilterKind    string
		VersionFilterPattern string
		GoModTidyEnabled     bool
		ScmID                string
		WorkDir              string
	}{
		GoModFile:            filename,
		Module:               module,
		VersionFilterKind:    versionFilterKind,
		VersionFilterPattern: versionFilterPattern,
		GoModTidyEnabled:     goModTidy,
		ScmID:                scmID,
		WorkDir:              workdir,
	}

	manifest := bytes.Buffer{}
	if err := tmpl.Execute(&manifest, params); err != nil {
		logrus.Debugln(err)
		return nil, err
	}
	return manifest.Bytes(), nil
}
