package cargo

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/updatecli/updatecli/pkg/plugins/utils/cargo"

	"github.com/sirupsen/logrus"
)

var (
	// ValidFiles specifies accepted Cargo files.
	ValidFiles [1]string = [1]string{"Cargo.toml"}
)

type crateDependency struct {
	Name     string
	Registry string
	Version  string
	Inlined  bool
}

type crateMetadata struct {
	Name            string
	Dependencies    []crateDependency
	DevDependencies []crateDependency
}

func (c Cargo) generateManifest(crateName string, dependency crateDependency, relativeFile string, foundFile string, dependencyType string, targetCargoCleanupEnabled bool) (bytes.Buffer, error) {
	manifest := bytes.Buffer{}

	// No need to continue if both Cargo.lock and Cargo.toml do not need to be updated
	if !targetCargoCleanupEnabled && !isStrictSemver(dependency.Version) {
		return manifest, nil
	}

	tmpl, err := template.New("manifest").Parse(dependencyManifest)
	if err != nil {
		logrus.Debugln(err)
		return manifest, err
	}
	var existingSourceKey string
	if dependency.Inlined {
		existingSourceKey = fmt.Sprintf("%s.%s", dependencyType, dependency.Name)
	} else {
		existingSourceKey = fmt.Sprintf("%s.%s.version", dependencyType, dependency.Name)
	}
	var ConditionQuery string
	if dependency.Inlined {
		ConditionQuery = fmt.Sprintf("%s.(?:-=%s)", dependencyType, dependency.Name)
	} else {
		ConditionQuery = fmt.Sprintf("%s.(?:-=%s).version", dependencyType, dependency.Name)
	}
	var TargetKey string
	if dependency.Inlined {
		TargetKey = fmt.Sprintf("%s.%s", dependencyType, dependency.Name)
	} else {
		TargetKey = fmt.Sprintf("%s.%s.version", dependencyType, dependency.Name)
	}
	var Registry cargo.Registry
	if dependency.Registry != "" {
		if registry, found := c.spec.Registries[dependency.Registry]; found {
			Registry = registry
		}
	}

	sourceVersionFilterKind := "semver"
	sourceVersionFilterPattern := dependency.Version

	if isStrictSemver(dependency.Version) {
		sourceVersionFilterPattern = ">=" + dependency.Version

		if !c.spec.VersionFilter.IsZero() {
			sourceVersionFilterKind = c.versionFilter.Kind
			sourceVersionFilterPattern, err = c.versionFilter.GreaterThanPattern(dependency.Version)
			if err != nil {
				logrus.Debugf("building version filter pattern: %s", err)
				sourceVersionFilterPattern = "*"
			}
		}
	}

	params := struct {
		ActionID                   string
		ManifestName               string
		CrateName                  string
		DependencyName             string
		SourceID                   string
		SourceName                 string
		SourceVersionFilterKind    string
		SourceVersionFilterPattern string
		ExistingSourceID           string
		ExistingSourceName         string
		ExistingSourceKey          string
		ConditionID                string
		ConditionQuery             string
		File                       string
		TargetIDEnable             bool
		TargetID                   string
		TargetName                 string
		TargetFile                 string
		TargetKey                  string
		TargetCargoCleanupEnabled  bool
		TargetWorkdir              string
		ScmID                      string
		WithRegistry               bool
		RegistrySCMID              string
		RegistryRootDir            string
		RegistryURL                string
		RegistryAuthToken          string
		RegistryHeaderFormat       string
	}{
		ActionID:                   c.actionID,
		ManifestName:               fmt.Sprintf("deps(cargo): bump %s %q for %q crate", dependencyType, dependency.Name, crateName),
		CrateName:                  crateName,
		DependencyName:             dependency.Name,
		SourceID:                   dependency.Name,
		SourceName:                 fmt.Sprintf("Get latest %q crate version", dependency.Name),
		SourceVersionFilterKind:    sourceVersionFilterKind,
		SourceVersionFilterPattern: sourceVersionFilterPattern,
		ExistingSourceID:           fmt.Sprintf("%s-current-version", dependency.Name),
		ExistingSourceKey:          existingSourceKey,
		ExistingSourceName:         fmt.Sprintf("Get current %q crate version", dependency.Name),
		ConditionID:                dependency.Name,
		ConditionQuery:             ConditionQuery,
		File:                       relativeFile,
		TargetIDEnable:             isStrictSemver(dependency.Version),
		TargetID:                   dependency.Name,
		TargetName:                 fmt.Sprintf("deps(cargo): bump crate dependency %q to {{ source %q }}", dependency.Name, dependency.Name),
		TargetFile:                 filepath.Base(foundFile),
		TargetKey:                  TargetKey,
		TargetCargoCleanupEnabled:  targetCargoCleanupEnabled,
		TargetWorkdir:              filepath.Dir(foundFile),
		ScmID:                      c.scmID,
		WithRegistry:               dependency.Registry != "",
		RegistrySCMID:              Registry.SCMID,
		RegistryRootDir:            Registry.RootDir,
		RegistryURL:                Registry.URL,
		RegistryAuthToken:          Registry.Auth.Token,
		RegistryHeaderFormat:       Registry.Auth.HeaderFormat,
	}

	if err := tmpl.Execute(&manifest, params); err != nil {
		logrus.Debugln(err)
		return manifest, err
	}
	return manifest, nil
}

func (c Cargo) discoverCargoDependenciesManifests() ([][]byte, error) {
	var manifests [][]byte

	searchFromDir := c.rootDir
	// If the spec.RootDir is an absolute path, then it as already been set
	// correctly in the New function.
	if c.spec.RootDir != "" && !path.IsAbs(c.spec.RootDir) {
		searchFromDir = filepath.Join(c.rootDir, c.spec.RootDir)
	}

	foundCargoFiles, err := findCargoFiles(
		searchFromDir,
		ValidFiles[:],
	)

	if err != nil {
		return nil, err
	}

	for _, foundCargoFile := range foundCargoFiles {
		logrus.Debugf("parsing file %q", foundCargoFile)

		relativeFoundCargoFile, err := filepath.Rel(c.rootDir, foundCargoFile)
		if err != nil {
			// Jump to the next Cargo if current failed
			logrus.Debugln(err)
			continue
		}

		cargoTargetCleanManifestEnabled := false
		if cargo.IsLockFileDetected(filepath.Join(filepath.Dir(foundCargoFile), "Cargo.lock")) {
			switch cargo.IsCargoInstalled() {
			case true:
				cargoTargetCleanManifestEnabled = true
			case false:
				logrus.Warning("skipping, Cargo lock file detected but Updatecli couldn't detect the cargo command to update it in case of a Cargo.lock update")
				continue
			}
		}

		// Retrieve Cargo dependencies for each crate
		crate, err := getCrateMetadata(foundCargoFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		if crate == nil {
			continue
		}

		if len(crate.Dependencies) == 0 && len(crate.DevDependencies) == 0 {
			continue
		}

		cr := *crate
		dependencies := cr.Dependencies
		sort.Slice(dependencies, func(i, j int) bool {
			return dependencies[i].Name < dependencies[j].Name
		})
		devDependencies := cr.DevDependencies
		sort.Slice(devDependencies, func(i, j int) bool {
			return devDependencies[i].Name < devDependencies[j].Name
		})

		for _, dependency := range dependencies {

			if len(c.spec.Ignore) > 0 {
				if c.spec.Ignore.isMatchingRules(c.rootDir, relativeFoundCargoFile, dependency.Registry, dependency.Name, dependency.Version) {
					logrus.Debugf("Ignoring %s.%s from %q, as matching ignore rule(s)\n", dependency.Registry, dependency.Name, relativeFoundCargoFile)
					continue
				}
			}

			if len(c.spec.Only) > 0 {
				if !c.spec.Only.isMatchingRules(c.rootDir, relativeFoundCargoFile, dependency.Registry, dependency.Name, dependency.Version) {
					logrus.Debugf("Ignoring package %s.%s from %q, as not matching only rule(s)\n", dependency.Registry, dependency.Name, relativeFoundCargoFile)
					continue
				}
			}

			manifest, err := c.generateManifest(cr.Name, dependency, relativeFoundCargoFile, foundCargoFile, "dependencies", cargoTargetCleanManifestEnabled)
			if err != nil {
				logrus.Debugln(err)
				continue
			}
			manifests = append(manifests, manifest.Bytes())
		}
		for _, dependency := range devDependencies {
			if len(c.spec.Ignore) > 0 {
				if c.spec.Ignore.isMatchingRules(c.rootDir, relativeFoundCargoFile, dependency.Registry, dependency.Name, dependency.Version) {
					logrus.Debugf("Ignoring %s.%s from %q, as matching ignore rule(s)\n", dependency.Registry, dependency.Name, relativeFoundCargoFile)
					continue
				}
			}

			if len(c.spec.Only) > 0 {
				if !c.spec.Only.isMatchingRules(c.rootDir, relativeFoundCargoFile, dependency.Registry, dependency.Name, dependency.Version) {
					logrus.Debugf("Ignoring package %s.%s from %q, as not matching only rule(s)\n", dependency.Registry, dependency.Name, relativeFoundCargoFile)
					continue
				}
			}

			manifest, err := c.generateManifest(cr.Name, dependency, relativeFoundCargoFile, foundCargoFile, "dev-dependencies", cargoTargetCleanManifestEnabled)
			if err != nil {
				logrus.Debugln(err)
				continue
			}
			manifests = append(manifests, manifest.Bytes())
		}
	}

	logrus.Debugf("found manifests: %s", manifests)
	return manifests, nil
}
