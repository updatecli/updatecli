package npm

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
)

func (n Npm) discoverDependencyManifests() ([][]byte, error) {

	var manifests [][]byte

	foundFiles, err := searchPackageJsonFiles(n.rootDir)

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

		// Test if the ignore rule based on path is respected
		if len(n.spec.Ignore) > 0 && n.spec.Ignore.isMatchingIgnoreRule(n.rootDir, relativeFoundFile) {
			logrus.Debugf("Ignoring %q as not matching rule(s)\n", foundFile)
			continue
		}

		// Test if the only rule based on path is respected
		if len(n.spec.Only) > 0 && !n.spec.Only.isMatchingOnlyRule(n.rootDir, relativeFoundFile) {
			logrus.Debugf("Ignoring %q as not matching rule(s)\n", foundFile)
			continue
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
				// If a dependency already contains a version constraint then we ignore it
				if isVersionConstraintSpecified(dependencyName, dependencyVersion, n.spec.StrictSemver) {
					continue
				}

				if err != nil {
					logrus.Debugln(err)
					continue
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
					SourceVersionFilterKind:    "semver",
					SourceVersionFilterPattern: dependencyVersion,
					TargetID:                   "npm",
					TargetName:                 fmt.Sprintf("Bump %q package version", dependencyName),
					// NPM package allows dot in package name which has a different meaning in Dasel query
					// Therefor we must escape it for Dasel query to work
					TargetKey:                fmt.Sprintf("%s.%s", dependencyType, strings.ReplaceAll(dependencyName, ".", `\.`)),
					TargetPackageJsonEnabled: false,
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
