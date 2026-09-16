package golang

import (
	"bytes"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/vulnerability"
)

// discoverDependencyManifests search for each go.mod file
// and then try to update both "direct" Go module and the Golang version
//
//nolint:funlen
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

		goVersion, goModules, goModulesToReplace, appliedReplaces, err := getGoModContent(foundFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		// newSecurityManifestParams returns the security manifest parameters shared by every module of the go.mod file
		newSecurityManifestParams := func() securityManifestParams {
			return securityManifestParams{
				ActionID:         g.actionID,
				GoModFile:        relativeFoundFile,
				Vulnerability:    g.spec.Vulnerability,
				GoModTidyEnabled: goModTidyEnabled,
				ScmID:            g.scmID,
				WorkDir:          relativeWorkDir,
			}
		}

		generateModuleManifests := func(modules map[string]string) {

			for goModule, goModuleVersion := range modules {
				// Skip golang module manifest if there is only one rule on the go version
				if g.spec.Only.isGoVersionOnly() || g.onlygoVersion {
					break
				}
				// Test if the ignore rule based on path is respected
				if len(g.spec.Ignore) > 0 {
					if g.spec.Ignore.isMatchingRules(g.rootDir, relativeFoundFile, goVersion, goModule, goModuleVersion, false) {
						logrus.Debugf("Ignoring module %q from file %q, as matching ignore rule(s)\n", goModule, relativeFoundFile)
						continue
					}
				}

				// Test if the only rule based on path is respected
				if len(g.spec.Only) > 0 {
					if !g.spec.Only.isMatchingRules(g.rootDir, relativeFoundFile, goVersion, goModule, goModuleVersion, false) {
						logrus.Debugf("Ignoring module %q from %q, as not matching only rule(s)\n", goModule, relativeFoundFile)
						continue
					}
				}

				goModuleVersionPattern := g.versionFilter.Pattern
				goModuleVersionKind := g.versionFilter.Kind
				switch isPseudoVersion(goModuleVersion) {
				case false:
					goModuleVersionPattern, err = g.versionFilter.GreaterThanPattern(goModuleVersion)
					if err != nil {
						logrus.Debugf("skipping golang module %q due to: %s", goModule, err)
						continue
					}

				case true:
					logrus.Debugf("Module %q uses a pseudo-version %q, so ignoring version filter pattern for this module as the registry will only return one version", goModule, goModuleVersion)
					// If the new version is a pseudo-version,
					// we cannot apply a version filter pattern as
					// golang registry will only return one version for this module.
					goModuleVersionKind = "latest"
					goModuleVersionPattern = ""
				}

				moduleManifest, err := getGolangModuleManifest(
					relativeFoundFile,
					goModule,
					goModuleVersionKind,
					goModuleVersionPattern,
					g.versionFilter.Regex,
					g.scmID,
					g.actionID,
					relativeWorkDir,
					goModTidyEnabled,
					g.spec.Age,
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
					if g.spec.Ignore.isMatchingRules(g.rootDir, relativeFoundFile, goVersion, replace.NewPath, replace.NewVersion, true) {
						logrus.Debugf("Ignoring module %q from file %q, as matching ignore rule(s)\n", replace.NewPath, relativeFoundFile)
						continue
					}
				}

				// Test if the only rule based on path is respected
				if len(g.spec.Only) > 0 {
					if !g.spec.Only.isMatchingRules(g.rootDir, relativeFoundFile, goVersion, replace.NewPath, replace.NewVersion, true) {
						logrus.Debugf("Ignoring module %q from %q, as not matching only rule(s)\n", replace.NewPath, relativeFoundFile)
						continue
					}
				}

				goModuleVersionPattern := g.versionFilter.Pattern
				goModuleVersionKind := g.versionFilter.Kind
				switch isPseudoVersion(replace.NewVersion) {
				case false:
					goModuleVersionPattern, err = g.versionFilter.GreaterThanPattern(replace.NewVersion)
					if err != nil {
						logrus.Debugf("skipping golang module %q due to: %s", replace.NewPath, err)
						continue
					}

				case true:
					// If the new version is a pseudo-version,
					// we cannot apply a version filter pattern as
					// golang registry will only return one version for this module.

					logrus.Debugf("Module %q uses a pseudo-version %q, so ignoring version filter for this module as the registry will only return one version", replace.NewPath, replace.NewVersion)

					goModuleVersionKind = "latest"
					goModuleVersionPattern = ""
				}

				moduleManifest, err := getGolangReplaceModuleManifest(
					relativeFoundFile,
					replace.OldPath,
					replace.OldVersion,
					replace.NewPath,
					goModuleVersionKind,
					goModuleVersionPattern,
					g.versionFilter.Regex,
					g.scmID,
					g.actionID,
					relativeWorkDir,
					goModTidyEnabled,
					g.spec.Age,
				)
				if err != nil {
					logrus.Debugf("skipping golang module %q module due to: %s", replace.NewPath, err)
					continue
				}

				manifests = append(manifests, moduleManifest)
			}
		}

		// generateSecurityManifests checks every direct module against the OSV database.
		// Only and ignore rules are evaluated against the direct module, even when it's replaced.
		generateSecurityManifests := func() {

			for goModule, goModuleVersion := range goModules {
				replace, replaced := appliedReplaces[goModule]

				// Test if the ignore rule based on path is respected
				if len(g.spec.Ignore) > 0 {
					if g.spec.Ignore.isMatchingRules(g.rootDir, relativeFoundFile, goVersion, goModule, goModuleVersion, replaced) {
						logrus.Debugf("Ignoring module %q from file %q, as matching ignore rule(s)\n", goModule, relativeFoundFile)
						continue
					}
				}

				// Test if the only rule based on path is respected
				if len(g.spec.Only) > 0 {
					if !g.spec.Only.isMatchingRules(g.rootDir, relativeFoundFile, goVersion, goModule, goModuleVersion, replaced) {
						logrus.Debugf("Ignoring module %q from %q, as not matching only rule(s)\n", goModule, relativeFoundFile)
						continue
					}
				}

				params := newSecurityManifestParams()
				params.Module = goModule
				params.Version = goModuleVersion
				params.TargetModule = goModule

				if replaced {
					switch {
					case replace.isLocal():
						logrus.Debugf("skipping golang module %q security manifest as it's replaced by the local path %q in %q", goModule, replace.NewPath, relativeFoundFile)
						continue
					case replace.Ambiguous:
						logrus.Warningf("skipping golang module %q security manifest as its replace directive without version can't be updated while another one with a version exists in %q", goModule, relativeFoundFile)
						continue
					}

					// The replacement module is the one built, so it's the one checked against the OSV database
					params.Module = replace.NewPath
					params.Version = replace.NewVersion
					params.Replace = true
					params.ReplaceVersion = replace.OldVersion
				}

				moduleManifest, err := getGolangModuleSecurityManifest(params)
				if err != nil {
					logrus.Debugf("skipping golang module %q security manifest due to: %s", params.Module, err)
					continue
				}

				manifests = append(manifests, moduleManifest)
			}
		}

		if g.spec.Vulnerability != nil {
			generateSecurityManifests()
		} else {
			generateModuleManifests(goModules)
			generateReplaceModuleManifests(goModulesToReplace)
		}

		if g.spec.Only.isGoModuleOnly() || g.onlyGoModule {
			continue
		}

		if g.spec.Vulnerability != nil {
			logrus.Debugf("Ignoring golang version update from file %q, as Go version security updates are not supported", relativeFoundFile)
			continue
		}

		// Test if the ignore rule based on path is respected
		if len(g.spec.Ignore) > 0 {
			if g.spec.Ignore.isMatchingRules(g.rootDir, relativeFoundFile, goVersion, "", "", false) {
				logrus.Debugf("Ignoring golang version update from file %q, as matching ignore rule(s)\n", relativeFoundFile)
				continue
			}
		}

		// Test if the only rule based on path is respected
		if len(g.spec.Only) > 0 {
			if !g.spec.Only.isMatchingRules(g.rootDir, relativeFoundFile, goVersion, "", "", false) {
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
				g.versionFilter.Regex,
				goVersionPattern,
				g.scmID,
				g.actionID,
				g.spec.Age)
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

// parseManifestTemplate parses a manifest template along with the templates it can reference, "tidy" and "vulnerability".
func parseManifestTemplate(text string) (*template.Template, error) {
	return template.New("manifest").Parse(vulnerability.ManifestTemplate + goTidyTemplate + text)
}

func getGolangVersionManifest(
	filename,
	versionFilterKind,
	versionFilterRegex,
	versionFilterPattern,
	scmID,
	actionID string,
	a age.Spec) ([]byte, error) {
	tmpl, err := parseManifestTemplate(goManifestTemplate)
	if err != nil {
		logrus.Debugln(err)
		return nil, err
	}

	params := struct {
		ActionID             string
		GoModFile            string
		VersionFilterKind    string
		VersionFilterPattern string
		VersionFilterRegex   string
		ScmID                string
		Age                  age.Spec
	}{
		ActionID:             actionID,
		GoModFile:            filename,
		VersionFilterKind:    versionFilterKind,
		VersionFilterPattern: versionFilterPattern,
		VersionFilterRegex:   versionFilterRegex,
		ScmID:                scmID,
		Age:                  a,
	}

	manifest := bytes.Buffer{}
	if err := tmpl.Execute(&manifest, params); err != nil {
		logrus.Debugln(err)
		return nil, err
	}
	return manifest.Bytes(), nil
}

func getGolangModuleManifest(
	filename,
	module,
	versionFilterKind,
	versionFilterPattern,
	versionFilterRegex,
	scmID,
	actionID,
	workdir string,
	goModTidy bool,
	a age.Spec) ([]byte, error) {

	tmpl, err := parseManifestTemplate(goModuleManifestTemplate)
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
		VersionFilterRegex   string
		GoModTidyEnabled     bool
		ScmID                string
		WorkDir              string
		Age                  age.Spec
	}{
		ActionID:             actionID,
		GoModFile:            filename,
		Module:               module,
		VersionFilterKind:    versionFilterKind,
		VersionFilterPattern: versionFilterPattern,
		VersionFilterRegex:   versionFilterRegex,
		GoModTidyEnabled:     goModTidy,
		ScmID:                scmID,
		WorkDir:              workdir,
		Age:                  a,
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
	versionFilterRegex,
	scmID,
	actionID,
	workdir string,
	goModTidy bool,
	a age.Spec) ([]byte, error) {

	tmpl, err := parseManifestTemplate(goReplaceModuleManifestTemplate)
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
		VersionFilterRegex   string
		GoModTidyEnabled     bool
		ScmID                string
		WorkDir              string
		Age                  age.Spec
	}{
		ActionID:             actionID,
		GoModFile:            filename,
		OldPathModule:        oldPathModule,
		OldVersionModule:     oldVersionModule,
		NewPathModule:        newPathModule,
		VersionFilterKind:    versionFilterKind,
		VersionFilterPattern: versionFilterPattern,
		VersionFilterRegex:   versionFilterRegex,
		GoModTidyEnabled:     goModTidy,
		ScmID:                scmID,
		WorkDir:              workdir,
		Age:                  a,
	}

	manifest := bytes.Buffer{}
	if err := tmpl.Execute(&manifest, params); err != nil {
		logrus.Debugln(err)
		return nil, err
	}
	return manifest.Bytes(), nil
}

// securityManifestParams holds the values rendered by goModuleSecurityManifestTemplate.
type securityManifestParams struct {
	ActionID  string
	GoModFile string
	// Module is the module checked against the OSV database, the replacement module for a replace directive.
	Module string
	// Version is the version of Module.
	Version string
	// TargetModule is the module updated in the go.mod file, the replaced module for a replace directive.
	TargetModule string
	// Replace indicates that the manifest updates a replace directive.
	Replace bool
	// ReplaceVersion is the replaced module version of the replace directive, if any.
	ReplaceVersion string
	// Vulnerability holds the settings shared by generated "vulnerability/osv" specs.
	Vulnerability    *vulnerability.Spec
	GoModTidyEnabled bool
	ScmID            string
	WorkDir          string
}

func getGolangModuleSecurityManifest(params securityManifestParams) ([]byte, error) {
	tmpl, err := parseManifestTemplate(goModuleSecurityManifestTemplate)
	if err != nil {
		logrus.Debugln(err)
		return nil, err
	}

	manifest := bytes.Buffer{}
	if err := tmpl.Execute(&manifest, params); err != nil {
		logrus.Debugln(err)
		return nil, err
	}
	return manifest.Bytes(), nil
}
