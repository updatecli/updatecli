package cargo

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/updatecli/updatecli/pkg/plugins/utils/cargo"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"

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
	Name                     string
	CargoFile                string
	CargoLockFile            string
	Workspace                bool
	WorkspaceMembers         []crateMetadata
	WorkspaceDependencies    []crateDependency
	WorkspaceDevDependencies []crateDependency
	Dependencies             []crateDependency
	DevDependencies          []crateDependency
}

func (c Cargo) generateManifest(
	crate *crateMetadata,
	dependency *crateDependency,
	dependencyType string,
	dependencyTypeLabel string,
) (bytes.Buffer, error) {
	manifest := bytes.Buffer{}

	tmpl, err := template.New("manifest").Parse(dependencyManifest)
	if err != nil {
		logrus.Errorln(err)
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

	filter := c.spec.VersionFilter
	if c.spec.VersionFilter.IsZero() {
		filter.Kind = version.SEMVERVERSIONKIND
		filter.Pattern = "*"
	}

	sourceVersionFilterKind := filter.Kind
	sourceVersionFilterPattern := filter.Pattern

	if filter.Kind == version.SEMVERVERSIONKIND && filter.Pattern != "*" {
		sourceVersionFilterPattern = dependency.Version

		if isStrictSemver(dependency.Version) {

			sourceVersionFilterPattern, err = c.versionFilter.GreaterThanPattern(dependency.Version)
			if err != nil {
				logrus.Debugf("building version filter pattern: %s", err)
				sourceVersionFilterPattern = "*"
			}
		}
	}

	cargoFile, err := filepath.Rel(c.rootDir, crate.CargoFile)
	if err != nil {
		return manifest, err
	}
	cargoLockFile := crate.CargoLockFile
	if cargoLockFile != "" {
		relCargoLockFile, err := filepath.Rel(c.rootDir, cargoLockFile)
		if err != nil {
			return manifest, err
		}
		cargoLockFile = relCargoLockFile
	}

	manifestName := fmt.Sprintf("deps(cargo): bump %s %q for %q crate", dependencyTypeLabel, dependency.Name, crate.Name)
	targetName := fmt.Sprintf("deps(cargo): bump crate %s %q to {{ source %q }}", dependencyTypeLabel, dependency.Name, dependency.Name)
	cargoLockTargetName := fmt.Sprintf("deps(cargo): update %s following bump crate %s %q to {{ source %q }}", cargoLockFile, dependencyTypeLabel, dependency.Name, dependency.Name)
	if crate.Workspace {
		manifestName = fmt.Sprintf("deps(cargo): bump %s %q", dependencyTypeLabel, dependency.Name)
		targetName = fmt.Sprintf("deps(cargo): bump %s %q to {{ source %q }}", dependencyTypeLabel, dependency.Name, dependency.Name)
		cargoLockTargetName = fmt.Sprintf("deps(cargo): update %s following bump %s %q to {{ source %q }}", cargoLockFile, dependencyTypeLabel, dependency.Name, dependency.Name)
	}

	params := struct {
		ActionID                   string
		ManifestName               string
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
		CargoUpgradeAvailable      bool
		CargoFile                  string
		CargoLockFile              string
		CargoLockTargetName        string
		TargetID                   string
		TargetName                 string
		TargetKey                  string
		ScmID                      string
		WithRegistry               bool
		RegistrySCMID              string
		RegistryRootDir            string
		RegistryURL                string
		RegistryAuthToken          string
		RegistryHeaderFormat       string
	}{
		ActionID:                   c.actionID,
		ManifestName:               manifestName,
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
		CargoUpgradeAvailable:      c.cargoUpgradeAvailable,
		CargoFile:                  cargoFile,
		CargoLockFile:              cargoLockFile,
		TargetID:                   dependency.Name,
		TargetName:                 targetName,
		CargoLockTargetName:        cargoLockTargetName,
		TargetKey:                  TargetKey,
		ScmID:                      c.scmID,
		WithRegistry:               dependency.Registry != "",
		RegistrySCMID:              Registry.SCMID,
		RegistryRootDir:            Registry.RootDir,
		RegistryURL:                Registry.URL,
		RegistryAuthToken:          Registry.Auth.Token,
		RegistryHeaderFormat:       Registry.Auth.HeaderFormat,
	}

	if err := tmpl.Execute(&manifest, params); err != nil {
		logrus.Errorln(err)
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

		// Retrieve Cargo dependencies for each crate
		crate, err := getCrateMetadata(filepath.Dir(foundCargoFile))
		if err != nil {
			logrus.Debugln(err)
			continue
		}
		c.processCrateMetadata(&manifests, crate)
		for _, member := range crate.WorkspaceMembers {
			c.processCrateMetadata(&manifests, member)
		}
	}

	return manifests, nil
}

func (c Cargo) processCrateMetadata(
	manifests *[][]byte,
	crate crateMetadata,
) {

	if crate.Name == "" && len(crate.WorkspaceMembers) == 0 {
		return
	}

	if len(crate.Dependencies) == 0 && len(crate.DevDependencies) == 0 && len(crate.WorkspaceDependencies) == 0 && len(crate.WorkspaceDevDependencies) == 0 {
		return
	}
	if crate.CargoLockFile != "" && !c.cargoAvailable && !c.cargoUpgradeAvailable {
		logrus.Warning("skipping, Cargo lock file detected but Updatecli couldn't detect nor the `cargo` command neither the `cargo upgrade` to update it in case of a Cargo.lock update")
		return
	}

	cr := crate

	dependencies := cr.Dependencies
	sort.Slice(dependencies, func(i, j int) bool {
		return dependencies[i].Name < dependencies[j].Name
	})
	devDependencies := cr.DevDependencies
	sort.Slice(devDependencies, func(i, j int) bool {
		return devDependencies[i].Name < devDependencies[j].Name
	})
	workspaceDependencies := cr.WorkspaceDependencies
	sort.Slice(dependencies, func(i, j int) bool {
		return dependencies[i].Name < dependencies[j].Name
	})
	workspaceDevDependencies := cr.WorkspaceDevDependencies
	sort.Slice(devDependencies, func(i, j int) bool {
		return devDependencies[i].Name < devDependencies[j].Name
	})

	c.processDependencies(manifests, cr.Name, &crate, dependencies, "dependencies", "dependency")
	c.processDependencies(manifests, cr.Name, &crate, devDependencies, "dev-dependencies", "dev dependency")
	c.processDependencies(manifests, cr.Name, &crate, workspaceDependencies, "workspace.dependencies", "workspace dependency")
	c.processDependencies(manifests, cr.Name, &crate, workspaceDevDependencies, "workspace.dev-dependencies", "workspace dev dependency")
}

func (c Cargo) processDependencies(
	manifests *[][]byte,
	name string,
	crate *crateMetadata,
	dependencies []crateDependency,
	depType string,
	depTypeLabel string,
) {
	for _, dependency := range dependencies {
		relativeFoundCargoFile, err := filepath.Rel(c.rootDir, crate.CargoFile)

		if err != nil {
			// Jump to the next Cargo if current failed
			logrus.Debugln(err)
			continue
		}

		if len(c.spec.Ignore) > 0 && c.spec.Ignore.isMatchingRules(c.rootDir, relativeFoundCargoFile, dependency.Registry, dependency.Name, dependency.Version) {
			logrus.Debugf("Ignoring %s.%s from %q, as matching ignore rule(s)\n", dependency.Registry, dependency.Name, crate.CargoFile)
			continue
		}
		if len(c.spec.Only) > 0 && !c.spec.Only.isMatchingRules(c.rootDir, relativeFoundCargoFile, dependency.Registry, dependency.Name, dependency.Version) {
			logrus.Debugf("Ignoring package %s.%s from %q, as not matching only rule(s)\n", dependency.Registry, dependency.Name, crate.CargoFile)
			continue
		}

		manifest, err := c.generateManifest(crate, &dependency, depType, depTypeLabel)
		if err != nil {
			logrus.Debugln(err)
			continue
		}
		*manifests = append(*manifests, manifest.Bytes())
	}
}
