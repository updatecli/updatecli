package cargo

import (
	"bytes"
	"fmt"
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
	params := struct {
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
		TargetID                   string
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
		ManifestName: fmt.Sprintf("Bump %s %q for %q crate", dependencyType, dependency.Name, crateName),
		CrateName:    crateName, DependencyName: dependency.Name,
		SourceID:                   dependency.Name,
		SourceName:                 fmt.Sprintf("Get latest %q crate version", dependency.Name),
		SourceVersionFilterKind:    "semver",
		SourceVersionFilterPattern: "*",
		ExistingSourceID:           fmt.Sprintf("%s-current-version", dependency.Name),
		ExistingSourceKey:          existingSourceKey,
		ExistingSourceName:         fmt.Sprintf("Get current %q crate version", dependency.Name),
		ConditionID:                dependency.Name,
		ConditionQuery:             ConditionQuery,
		File:                       relativeFile,
		TargetID:                   dependency.Name,
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

	foundCargoFiles, err := findCargoFiles(
		c.rootDir,
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

		cargoRelativePath := filepath.Dir(relativeFoundCargoFile)
		cargoCrateName := filepath.Base(cargoRelativePath)

		// Test if the ignore rule based on path doesn't match
		if len(c.spec.Ignore) > 0 && c.spec.Ignore.isMatchingIgnoreRule(c.rootDir, relativeFoundCargoFile) {
			logrus.Debugf("Ignoring Cargo Crate %q from %q, as not matching rule(s)\n",
				cargoCrateName,
				cargoRelativePath)
			continue
		}

		// Test if the only rule based on path match
		if len(c.spec.Only) > 0 && !c.spec.Only.isMatchingOnlyRule(c.rootDir, relativeFoundCargoFile) {
			logrus.Debugf("Ignoring Cargo Crate %q from %q, as not matching rule(s)\n",
				cargoCrateName,
				cargoRelativePath)
			continue
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
			manifest, err := c.generateManifest(cr.Name, dependency, relativeFoundCargoFile, foundCargoFile, "dependencies", cargoTargetCleanManifestEnabled)
			if err != nil {
				logrus.Debugln(err)
				continue
			}
			manifests = append(manifests, manifest.Bytes())
		}
		for _, dependency := range devDependencies {
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
