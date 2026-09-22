package npm

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/vulnerability"
)

type lockFileSupport struct {
	yarn bool
	pnpm bool
	npm  bool
}

// detectLockFileSupport checks which package managers are available for lock file updates.
// Returns the support config and whether to skip this package.json entirely.
func detectLockFileSupport(dir string) (lockFileSupport, bool) {
	// Returns config and whether to skip this package.json
	support := lockFileSupport{
		yarn: false,
		pnpm: false,
		npm:  false,
	}

	if isLockFileDetected(filepath.Join(dir, "yarn.lock")) {
		switch isYarnInstalled() {
		case true:
			support.yarn = true
		case false:
			logrus.Warning("skipping, Yarn lock file detected but Updatecli couldn't detect the yarn command to update it in case of a package.json update")
			return support, true
		}
	}

	// It doesn't make sense to update the package.json if Updatecli do not have access to the pnpm to update the lock file pnpm-lock.yaml.
	if isLockFileDetected(filepath.Join(dir, "pnpm-lock.yaml")) {
		switch isPnpmInstalled() {
		case true:
			support.pnpm = true
		case false:
			logrus.Warning("skipping, Pnpm lock file detected but Updatecli couldn't detect the pnpm command to update it in case of a package.json update")
			return support, true
		}
	}

	// It doesn't make sense to update the package.json if Updatecli do not have access to the npm command to update package-lock.json
	if isLockFileDetected(filepath.Join(dir, "package-lock.json")) {
		switch isNpmInstalled() {
		case true:
			support.npm = true
		case false:
			logrus.Warning("skipping, NPM lock file detected but Updatecli couldn't detect the npm command to update it in case of a package.json update")
			return support, true
		}
	}
	return support, false
}

// reportUnresolvedDependencies warns about the dependencies of a package.json skipped for lack of an
// installed version, so a scan examining nothing can't be mistaken for one finding nothing.
func (n Npm) reportUnresolvedDependencies(dependencies []string, packageJson string, locked lockedVersions) {
	if len(dependencies) == 0 {
		return
	}

	sort.Strings(dependencies)
	skipped := strings.Join(dependencies, ", ")

	if locked.lockFile == "" {
		logrus.Warningf("no lock file found for %q, skipping its dependencies using a version constraint: %s", packageJson, skipped)
		return
	}

	lockFile, err := filepath.Rel(n.rootDir, locked.lockFile)
	if err != nil {
		lockFile = locked.lockFile
	}

	logrus.Warningf("no version resolved in %q for some dependencies of %q, skipping: %s", lockFile, packageJson, skipped)
}

// lockFileLocation returns the directory of the lock file of a project, and the path to it from the project directory
// where the package manager command runs, such as "../../" for a workspace project whose lock file is at the root.
// The package manager updates the workspace lock file along with the project package.json.
// Without a lock file resolved, the lock file is looked up next to the package.json.
func lockFileLocation(projectDir string, locked lockedVersions) (string, string) {
	if locked.lockFile == "" {
		return projectDir, ""
	}

	lockDir := filepath.Dir(locked.lockFile)
	relativeLockDir, err := filepath.Rel(projectDir, lockDir)
	if err != nil || relativeLockDir == "." {
		return lockDir, ""
	}

	return lockDir, filepath.ToSlash(relativeLockDir) + "/"
}

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

		locked, lockErr := loadLockedVersions(filepath.Dir(foundFile), searchFromDir)
		if lockErr != nil && n.spec.Vulnerability != nil {
			// Without the lock file, no dependency using a version constraint can be scanned
			logrus.Warningf("%s: skipping the dependencies of %q using a version constraint", lockErr, relativeFoundFile)
		}

		lockDir, lockFilePrefix := lockFileLocation(filepath.Dir(foundFile), locked)

		lockSupport, skip := detectLockFileSupport(lockDir)
		if skip {
			continue
		}

		data, err := loadPackageJsonData(foundFile)

		if err != nil {
			logrus.Debugln(err)
			continue
		}

		// unresolvedDependencies collects the dependencies dropped for lack of an installed version, so the
		// user is told about them at once rather than getting a silently incomplete vulnerability report.
		var unresolvedDependencies []string

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

				params := manifestTemplateParams{
					ManifestName:    fmt.Sprintf("Bump %q package version", dependencyName),
					SourceID:        npmIdentifier,
					SourceName:      fmt.Sprintf("Get %q package version", dependencyName),
					SourceKind:      npmIdentifier,
					SourceNPMName:   dependencyName,
					SourceNpmrcPath: n.npmrcPath,
					TargetID:        npmIdentifier,
					TargetName:      fmt.Sprintf("Bump %q package version to {{ source \"npm\" }}", dependencyName),
					// NPM package allows dot in package name which has a different meaning in Dasel query
					// Therefor we must escape it for Dasel query to work
					TargetKey:                fmt.Sprintf("%s.%s", dependencyType, strings.ReplaceAll(dependencyName, ".", `\.`)),
					TargetPackageJsonEnabled: !lockSupport.yarn && !lockSupport.pnpm && !lockSupport.npm,
					TargetYarnCleanupEnabled: lockSupport.yarn,
					TargetPnpmCleanupEnabled: lockSupport.pnpm,
					TargetNPMCleanupEnabled:  lockSupport.npm,
					TargetWorkdir:            filepath.Dir(relativeFoundFile),
					TargetLockFilePrefix:     lockFilePrefix,
					TargetNPMCommand:         getTargetCommand(npmIdentifier, dependencyName),
					TargetYarnCommand:        getTargetCommand("yarn", dependencyName),
					TargetPnpmCommand:        getTargetCommand("pnpm", dependencyName),
					File:                     relativeFoundFile,
					ScmID:                    n.scmID,
				}

				if n.spec.Vulnerability != nil {
					currentVersion := dependencyVersion
					if isVersionConstraint {
						currentVersion = locked.version(dependencyName, dependencyVersion)
					}

					if _, err := semver.StrictNewVersion(currentVersion); err != nil {
						logrus.Debugf("Ignoring NPM package %q from %q, as no installed version could be identified for %q\n", dependencyName, relativeFoundFile, dependencyVersion)
						unresolvedDependencies = append(unresolvedDependencies, dependencyName)
						continue
					}

					// The pipeline name contains the fixed version, so it can be used as pullrequest title
					params.ManifestName = params.TargetName
					params.SourceName = fmt.Sprintf("Get lowest %q package version without known vulnerabilities", dependencyName)
					params.SourceKind = "vulnerability/osv"
					params.Vulnerability = n.spec.Vulnerability
					params.CurrentVersion = currentVersion
				} else {
					sourceVersionFilterKind := "semver"
					sourceVersionFilterPattern := dependencyVersion
					sourceVersionFilterRegex := "*"

					if isVersionConstraint && n.ignoreVersionConstraint {
						sourceVersionFilterPattern = "*"

						if !n.spec.VersionFilter.IsZero() && dependencyVersion != "latest" {
							guessedVersion, err := convertSemverVersionConstraintToVersion(dependencyVersion)
							if err != nil {
								logrus.Debugf("converting version constraint to version: %s", err)
								guessedVersion = ""
							}
							sourceVersionFilterKind = n.versionFilter.Kind
							sourceVersionFilterPattern, err = n.versionFilter.GreaterThanPattern(guessedVersion)
							sourceVersionFilterRegex = n.versionFilter.Regex
							if err != nil {
								logrus.Debugf("building version filter pattern: %s", err)
								sourceVersionFilterPattern = "*"
							}
						}
					}

					if isVersionConstraint && !n.ignoreVersionConstraint && !n.spec.VersionFilter.IsZero() {
						// User want to respect the version constraint defined in package.json
						// but also want to apply a version filter on top of it.
						// This is not supported at this time as we don't have a clear use case.
						logrus.Warningf("NPM package %q from %q: Ignoring version filter as version constraint %q is defined and ignoreVersionConstraints is set to false", dependencyName, relativeFoundFile, dependencyVersion)
						logrus.Warningf("NPM package %q from %q: If you want to apply a version filter, please set ignoreVersionConstraints to true", dependencyName, relativeFoundFile)
					}

					/*
						Pattern order
							1. Reuse version constraint defined from package.json
							2. If no version constraint defined then convert the version to ">=x.y.z"
							3. If no version constraint defined but versionfilter defined in the manifest
							   then we use that version filter kind and pattern
					*/

					if !isVersionConstraint {
						sourceVersionFilterPattern = ">=" + dependencyVersion

						if !n.spec.VersionFilter.IsZero() {
							sourceVersionFilterKind = n.versionFilter.Kind
							sourceVersionFilterPattern, err = n.versionFilter.GreaterThanPattern(dependencyVersion)
							sourceVersionFilterRegex = n.versionFilter.Regex
							if err != nil {
								logrus.Debugf("building version filter pattern: %s", err)
								sourceVersionFilterPattern = "*"
							}
						}
					}

					params.SourceVersionFilterKind = sourceVersionFilterKind
					params.SourceVersionFilterPattern = sourceVersionFilterPattern
					params.SourceVersionFilterRegex = sourceVersionFilterRegex
					params.SourceURL = n.url
					params.SourceRegistryToken = n.registryToken
					params.SourceAge = n.releaseAge
				}

				tmpl, err := template.New("manifest").Parse(age.ManifestTemplate + vulnerability.ManifestTemplate + manifestTemplate)
				if err != nil {
					logrus.Debugln(err)
					continue
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

		// The lock file error already reported the dependencies it drops
		if lockErr == nil && n.spec.Vulnerability != nil {
			n.reportUnresolvedDependencies(unresolvedDependencies, relativeFoundFile, locked)
		}
	}

	logrus.Printf("%v manifests identified", len(manifests))

	return manifests, nil
}
