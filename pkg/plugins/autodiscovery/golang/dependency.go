package golang

import (
	"bytes"
	"fmt"
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

		pwd, err := os.Getwd()
		if err != nil {
			continue
		}
		relativeFoundFile, err := filepath.Rel(pwd, foundFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		// If the Go binary is available then we can run `go mod tidy` in case of the go.mod modification
		goModTidyEnabled := false
		switch isGolangInstalled() {
		case true:
			goModTidyEnabled = true
		case false:
			logrus.Warning("Golang not detected so we can't run go mod tidy after go.mod file change")
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
		TargetName           string
		ScmID                string
	}{
		GoModFile:            filename,
		VersionFilterKind:    versionFilterKind,
		VersionFilterPattern: versionFilterPattern,
		TargetName:           `Bump Golang to {{ source "golangVersion" }}`,
		ScmID:                scmID,
	}

	manifest := bytes.Buffer{}
	if err := tmpl.Execute(&manifest, params); err != nil {
		logrus.Debugln(err)
		return nil, err
	}
	return manifest.Bytes(), nil
}

func getGolangModuleManifest(filename, module, versionFilterKind, versionFilterPattern, scmID string, goModTidy bool) ([]byte, error) {
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
		TargetName           string
	}{
		GoModFile:            filename,
		Module:               module,
		VersionFilterKind:    versionFilterKind,
		VersionFilterPattern: versionFilterPattern,
		GoModTidyEnabled:     goModTidy,
		ScmID:                scmID,
		TargetName:           fmt.Sprintf("Bump %s to {{ source \"golangModuleVersion\" }}", module),
	}

	manifest := bytes.Buffer{}
	if err := tmpl.Execute(&manifest, params); err != nil {
		logrus.Debugln(err)
		return nil, err
	}
	return manifest.Bytes(), nil
}
