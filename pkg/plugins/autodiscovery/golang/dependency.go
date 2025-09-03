package golang

import (
	"bytes"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"github.com/sirupsen/logrus"
)

// discoverDependencyManifests search for each go.mod file
// and then try to update both "direct" Go module and the Golang version
func (g Golang) discoverDependencyManifests() ([][]byte, error) {

	var manifests [][]byte

	searchFromDir := g.rootDir
	// If the spec.RootDir is an absolute path, then it as already been set
	// correctly in the New function.
	if g.spec.RootDir != "" && !path.IsAbs(g.spec.RootDir) {
		searchFromDir = filepath.Join(g.rootDir, g.spec.RootDir)
	}

	foundFiles, err := searchGoModFiles(searchFromDir)

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
		goSumFilePath := filepath.Join(filepath.Dir(foundFile), "go.sum")
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

		goVersion, goModules, goModulesToReplace, err := getGoModContent(foundFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		generateModuleManifests := func(modules map[string]string) {

			for goModule, goModuleVersion := range modules {
				// Skip golang module manifest if there is only one rule on the go version
				if g.spec.Only.isGoVersionOnly() || g.onlygoVersion {
					break
				}
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
					g.actionID,
					relativeWorkDir,
					goModTidyEnabled,
				)
				if err != nil {
					logrus.Debugf("skipping golang module %q module due to: %s", goModule, err)
					continue
				}

				manifests = append(manifests, moduleManifest)
			}
		}

		generateReplaceModuleManifests := func(modules []Replace) {

			for _, replace := range modules {
				// Skip golang module manifest if there is only one rule on the go version
				if g.spec.Only.isGoVersionOnly() || g.onlygoVersion {
					break
				}
				// Test if the ignore rule based on path is respected
				if len(g.spec.Ignore) > 0 {
					if g.spec.Ignore.isMatchingRules(g.rootDir, relativeFoundFile, goVersion, replace.NewPath, replace.NewVersion) {
						logrus.Debugf("Ignoring module %q from file %q, as matching ignore rule(s)\n", replace.NewPath, relativeFoundFile)
						continue
					}
				}

				// Test if the only rule based on path is respected
				if len(g.spec.Only) > 0 {
					if !g.spec.Only.isMatchingRules(g.rootDir, relativeFoundFile, goVersion, replace.NewPath, replace.NewVersion) {
						logrus.Debugf("Ignoring module %q from %q, as not matching only rule(s)\n", replace.NewPath, relativeFoundFile)
						continue
					}
				}

				goModuleVersionPattern, err := g.versionFilter.GreaterThanPattern(replace.NewVersion)
				if err != nil {
					logrus.Debugf("skipping golang module %q due to: %s", replace.NewPath, err)
					continue
				}

				moduleManifest, err := getGolangReplaceModuleManifest(
					relativeFoundFile,
					replace.OldPath,
					replace.OldVersion,
					replace.NewPath,
					g.versionFilter.Kind,
					goModuleVersionPattern,
					g.scmID,
					g.actionID,
					relativeWorkDir,
					goModTidyEnabled,
				)
				if err != nil {
					logrus.Debugf("skipping golang module %q module due to: %s", replace.NewPath, err)
					continue
				}

				manifests = append(manifests, moduleManifest)
			}
		}

		generateModuleManifests(goModules)
		generateReplaceModuleManifests(goModulesToReplace)

		if g.spec.Only.isGoModuleOnly() || g.onlyGoModule {
			continue
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
				goVersionPattern,
				g.scmID,
				g.actionID)
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

func getGolangVersionManifest(filename, versionFilterKind, versionFilterPattern, scmID, actionID string) ([]byte, error) {
	tmpl, err := template.New("manifest").Parse(goManifestTemplate)
	if err != nil {
		logrus.Debugln(err)
		return nil, err
	}

	params := struct {
		ActionID             string
		GoModFile            string
		VersionFilterKind    string
		VersionFilterPattern string
		ScmID                string
	}{
		ActionID:             actionID,
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

func getGolangModuleManifest(filename, module, versionFilterKind, versionFilterPattern, scmID, actionID, workdir string, goModTidy bool) ([]byte, error) {

	tmpl, err := template.New("manifest").Parse(goModuleManifestTemplate)
	if err != nil {
		logrus.Debugln(err)
		return nil, err
	}

	params := struct {
		ActionID             string
		GoModFile            string
		Module               string
		VersionFilterKind    string
		VersionFilterPattern string
		GoModTidyEnabled     bool
		ScmID                string
		WorkDir              string
		Replace              bool
		ReplaceVersion       string
	}{
		ActionID:             actionID,
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

func getGolangReplaceModuleManifest(filename,
	oldPathModule,
	oldVersionModule,
	newPathModule,
	versionFilterKind,
	versionFilterPattern,
	scmID,
	actionID,
	workdir string,
	goModTidy bool,
) ([]byte, error) {

	tmpl, err := template.New("manifest").Parse(goReplaceModuleManifestTemplate)
	if err != nil {
		logrus.Debugln(err)
		return nil, err
	}

	params := struct {
		ActionID             string
		GoModFile            string
		OldPathModule        string
		OldVersionModule     string
		NewPathModule        string
		VersionFilterKind    string
		VersionFilterPattern string
		GoModTidyEnabled     bool
		ScmID                string
		WorkDir              string
	}{
		ActionID:             actionID,
		GoModFile:            filename,
		OldPathModule:        oldPathModule,
		OldVersionModule:     oldVersionModule,
		NewPathModule:        newPathModule,
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
