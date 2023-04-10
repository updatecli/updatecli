package golang

import (
	"bytes"
	"path/filepath"
	"text/template"

	"github.com/sirupsen/logrus"
)

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

		goModTidyEnabled := false
		switch isGolangInstalled() {
		case true:
			goModTidyEnabled = true
		case false:
			logrus.Warning("Golang not detected so we won't run go mod tidy after go.mod file change")
		}

		// Test if the ignore rule based on path is respected
		if len(g.spec.Ignore) > 0 && g.spec.Ignore.isMatchingIgnoreRule(g.rootDir, relativeFoundFile) {
			logrus.Debugf("Ignoring %q as not matching rule(s)\n", foundFile)
			continue
		}

		// Test if the only rule based on path is respected
		if len(g.spec.Only) > 0 && !g.spec.Only.isMatchingOnlyRule(g.rootDir, relativeFoundFile) {
			logrus.Debugf("Ignoring %q as not matching rule(s)\n", foundFile)
			continue
		}

		goVersion, goModules, err := getGoModContent(foundFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		goVersionPattern, err := g.versionFilter.GreaterThanPattern(goVersion)
		golangVersionManifest := []byte{}
		if err != nil {
			logrus.Debugln(err)
		} else {
			golangVersionManifest, err = getGolangVersionManifest(relativeFoundFile, goVersionPattern, g.scmID)
			if err != nil {
				logrus.Debugln(err)
				logrus.Debugln("skipping golang version manifest due to previous error")
			}
		}

		manifests = append(manifests, golangVersionManifest)

		for goModule, goModuleVersion := range goModules {

			goModuleVersionPattern, err := g.versionFilter.GreaterThanPattern(goModuleVersion)
			if err != nil {
				logrus.Debugln(err)
				logrus.Debugf("skipping golang module %q module due to previous error", goModule)
				continue
			}

			moduleManifest, err := getGolangModuleManifest(
				relativeFoundFile,
				goModule,
				goModuleVersionPattern,
				g.scmID,
				goModTidyEnabled)
			if err != nil {
				logrus.Debugln(err)
				logrus.Debugf("skipping golang module %q module due to previous error", goModule)
				continue
			}

			manifests = append(manifests, moduleManifest)
		}
	}

	logrus.Printf("%v manifests identified", len(manifests))

	return manifests, nil
}

func getGolangVersionManifest(filename, versionPattern, scmID string) ([]byte, error) {
	tmpl, err := template.New("manifest").Parse(goManifestTemplate)
	if err != nil {
		logrus.Debugln(err)
		return nil, err
	}

	params := struct {
		GoModFile            string
		VersionFilterPattern string
		ScmID                string
	}{
		GoModFile:            filename,
		VersionFilterPattern: versionPattern,
		ScmID:                scmID,
	}

	manifest := bytes.Buffer{}
	if err := tmpl.Execute(&manifest, params); err != nil {
		logrus.Debugln(err)
		return nil, err
	}
	return manifest.Bytes(), nil
}

func getGolangModuleManifest(filename, module, versionPattern, scmID string, goModTidy bool) ([]byte, error) {
	tmpl, err := template.New("manifest").Parse(goModuleManifestTemplate)
	if err != nil {
		logrus.Debugln(err)
		return nil, err
	}

	params := struct {
		GoModFile            string
		Module               string
		VersionFilterPattern string
		GoModTidyEnabled     bool
		ScmID                string
	}{
		GoModFile:            filename,
		Module:               module,
		VersionFilterPattern: versionPattern,
		GoModTidyEnabled:     goModTidy,
		ScmID:                scmID,
	}

	manifest := bytes.Buffer{}
	if err := tmpl.Execute(&manifest, params); err != nil {
		logrus.Debugln(err)
		return nil, err
	}
	return manifest.Bytes(), nil
}
