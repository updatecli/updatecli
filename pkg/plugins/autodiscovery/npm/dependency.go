package npm

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
)

func (n Npm) discoverDependencyManifests() ([][]byte, error) {

	var manifests [][]byte

	searchFromDir := n.rootDir
	// If the spec.RootDir is an absolute path, then it as already been set
	// correctly in the New function.
	if n.spec.RootDir != "" && !path.IsAbs(n.spec.RootDir) {
		searchFromDir = filepath.Join(n.rootDir, n.spec.RootDir)
	}

	foundFiles, err := searchPackageJsonFiles(searchFromDir)

	if err != nil {
		return nil, err
	}

	for _, foundFile := range foundFiles {

		logrus.Debugf("parsing file %q", foundFile)

		relativeFoundFile, err := filepath.Rel(n.rootDir, foundFile)
		if err != nil {
			// Let's try the next pom.xml if one fail
			logrus.Debugln(err)
			continue
		}

		// It doesn't make sense to update the package.json if Updatecli do not have access to the yarn to update the lock file yarn.lock
		yarnTargetCleanManifestEnabled := false
		if isLockFileDetected(filepath.Join(filepath.Dir(foundFile), "yarn.lock")) {
			switch isYarnInstalled() {
			case true:
				yarnTargetCleanManifestEnabled = true
			case false:
				logrus.Warning("skipping, Yarn lock file detected but Updatecli couldn't detect the yarn command to update it in case of a package.json update")
				continue
			}
		}

		// It doesn't make sense to update the package.json if Updatecli do not have access to the npm command to update package-lock.json
		npmTargetCleanupManifestEnabled := false
		if isLockFileDetected(filepath.Join(filepath.Dir(foundFile), "package-lock.json")) {
			switch isNpmInstalled() {
			case true:
				npmTargetCleanupManifestEnabled = true
			case false:
				logrus.Warning("skipping, NPM lock file detected but Updatecli couldn't detect the npm command to update it in case of a package.json update")
				continue
			}
		}

		data, err := loadPackageJsonData(foundFile)

		if err != nil {
			logrus.Debugln(err)
			continue
		}

		getManifest := func(dependencies map[string]string, dependencyType string) {
			if len(dependencies) == 0 {
				logrus.Debugf("no NPM %s found in %q\n", dependencyType, foundFile)
				return
			}
			for dependencyName, dependencyVersion := range dependencies {
				if !isVersionConstraintSupported(dependencyName, dependencyVersion) {
					continue
				}

				if len(n.spec.Ignore) > 0 {
					if n.spec.Ignore.isMatchingRules(n.rootDir, relativeFoundFile, dependencyName, dependencyVersion) {
						logrus.Debugf("Ignoring NPM package %q from %q, as matching ignore rule(s)\n", dependencyName, relativeFoundFile)
						continue
					}
				}

				if len(n.spec.Only) > 0 {
					if !n.spec.Only.isMatchingRules(n.rootDir, relativeFoundFile, dependencyName, dependencyVersion) {
						logrus.Debugf("Ignoring NPM package %q from %q, as not matching only rule(s)\n", dependencyName, relativeFoundFile)
						continue
					}
				}

				isVersionConstraint := isVersionConstraintSpecified(
					dependencyName,
					dependencyVersion)

				// If a version constraint is specified such as "~4.0.0" then package.json shouldn't be updated
				// And if no lock file exist then we can skip this dependency
				if yarnTargetCleanManifestEnabled &&
					npmTargetCleanupManifestEnabled &&
					isVersionConstraint {
					continue
				}

				sourceVersionFilterKind := "semver"
				sourceVersionFilterPattern := dependencyVersion

				/*
					Pattern order
						1. Reuse version constraint defined from cargo.toml
						2. If no version constraint defined then convert the version to ">=x.y.z"
						3. If no version constraint defined but versionfilter defined in the manifest
						   then we use that version filter kind and pattern
				*/

				if !isVersionConstraint {
					sourceVersionFilterPattern = ">=" + dependencyVersion

					if !n.spec.VersionFilter.IsZero() {
						sourceVersionFilterKind = n.versionFilter.Kind
						sourceVersionFilterPattern, err = n.versionFilter.GreaterThanPattern(dependencyVersion)
						if err != nil {
							logrus.Debugf("building version filter pattern: %s", err)
							sourceVersionFilterPattern = "*"
						}
					}
				}

				tmpl, err := template.New("manifest").Parse(manifestTemplate)
				if err != nil {
					logrus.Debugln(err)
					continue
				}

				params := struct {
					ManifestName               string
					SourceID                   string
					SourceName                 string
					SourceKind                 string
					SourceNPMName              string
					SourceVersionFilterKind    string
					SourceVersionFilterPattern string
					TargetID                   string
					TargetName                 string
					TargetKey                  string
					TargetPackageJsonEnabled   bool
					TargetYarnCleanupEnabled   bool
					TargetNPMCleanupEnabled    bool
					TargetWorkdir              string
					TargetNPMCommand           string
					TargetYarnCommand          string
					File                       string
					ScmID                      string
				}{
					ManifestName:               fmt.Sprintf("Bump %q package version", dependencyName),
					SourceID:                   "npm",
					SourceName:                 fmt.Sprintf("Get %q package version", dependencyName),
					SourceKind:                 "npm",
					SourceNPMName:              dependencyName,
					SourceVersionFilterKind:    sourceVersionFilterKind,
					SourceVersionFilterPattern: sourceVersionFilterPattern,
					TargetID:                   "npm",
					TargetName:                 fmt.Sprintf("Bump %q package version to {{ source \"npm\" }}", dependencyName),
					// NPM package allows dot in package name which has a different meaning in Dasel query
					// Therefor we must escape it for Dasel query to work
					TargetKey:                fmt.Sprintf("%s.%s", dependencyType, strings.ReplaceAll(dependencyName, ".", `\.`)),
					TargetPackageJsonEnabled: !yarnTargetCleanManifestEnabled && !npmTargetCleanupManifestEnabled,
					TargetYarnCleanupEnabled: yarnTargetCleanManifestEnabled,
					TargetNPMCleanupEnabled:  npmTargetCleanupManifestEnabled,
					TargetWorkdir:            filepath.Dir(relativeFoundFile),
					TargetNPMCommand:         getTargetCommand("npm", dependencyName),
					TargetYarnCommand:        getTargetCommand("yarn", dependencyName),
					File:                     relativeFoundFile,
					ScmID:                    n.scmID,
				}

				manifest := bytes.Buffer{}
				if err := tmpl.Execute(&manifest, params); err != nil {
					logrus.Debugln(err)
					continue
				}

				manifests = append(manifests, manifest.Bytes())
			}
		}

		getManifest(data.Dependencies, "dependencies")
		getManifest(data.DevDependencies, "devDependencies")
	}

	logrus.Printf("%v manifests identified", len(manifests))

	return manifests, nil
}
