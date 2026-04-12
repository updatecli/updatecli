package pyproject

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// pyprojectTOML mirrors the subset of pyproject.toml we care about.
type pyprojectTOML struct {
	Project struct {
		Name                 string              `toml:"name"`
		Dependencies         []string            `toml:"dependencies"`
		OptionalDependencies map[string][]string `toml:"optional-dependencies"`
	} `toml:"project"`
}

// loadPyprojectData reads and unmarshals a pyproject.toml file.
func loadPyprojectData(filePath string) (pyprojectTOML, error) {
	var data pyprojectTOML
	if _, err := toml.DecodeFile(filePath, &data); err != nil {
		return data, fmt.Errorf("parsing %q: %w", filePath, err)
	}
	return data, nil
}

// discoverDependencyManifests is the main entry point called by DiscoverManifests.
func (p Pyproject) discoverDependencyManifests() ([][]byte, error) {
	var manifests [][]byte

	searchFromDir := p.rootDir
	// spec.RootDir relative paths are joined onto rootDir; absolute ones were resolved in New().
	if p.spec.RootDir != "" && !path.IsAbs(p.spec.RootDir) {
		searchFromDir = filepath.Join(p.rootDir, p.spec.RootDir)
	}

	foundFiles, err := findPyprojectFiles(searchFromDir)
	if err != nil {
		return nil, err
	}

	for _, foundFile := range foundFiles {
		logrus.Debugf("parsing file %q", foundFile)

		dir := filepath.Dir(foundFile)

		lockSupport, skip := detectLockFileSupport(dir, p.uvAvailable)
		if skip {
			continue
		}

		data, err := loadPyprojectData(foundFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		relativeFile, err := filepath.Rel(p.rootDir, foundFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		workdir, err := filepath.Rel(p.rootDir, dir)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		projectName := data.Project.Name

		// Process [project.dependencies] — the main dependency group.
		mainDeps := make([]string, len(data.Project.Dependencies))
		copy(mainDeps, data.Project.Dependencies)
		sort.Strings(mainDeps)

		manifests = append(manifests, p.processDependencies(mainDeps, "", relativeFile, lockSupport, projectName, workdir)...)

		// Process each [project.optional-dependencies] group in deterministic order.
		groups := make([]string, 0, len(data.Project.OptionalDependencies))
		for g := range data.Project.OptionalDependencies {
			groups = append(groups, g)
		}
		sort.Strings(groups)

		for _, group := range groups {
			groupDeps := make([]string, len(data.Project.OptionalDependencies[group]))
			copy(groupDeps, data.Project.OptionalDependencies[group])
			sort.Strings(groupDeps)

			manifests = append(manifests, p.processDependencies(groupDeps, group, relativeFile, lockSupport, projectName, workdir)...)
		}
	}

	logrus.Printf("%v manifests identified", len(manifests))

	return manifests, nil
}

// processDependencies generates manifests for a list of PEP 508 dependency strings.
// group is the optional-dependency group name, or "" for main dependencies.
func (p Pyproject) processDependencies(
	deps []string,
	group string,
	relativeFile string,
	lockSupport lockFileSupport,
	projectName string,
	workdir string,
) [][]byte {
	var manifests [][]byte

	tmpl, err := template.New("manifest").Parse(manifestTemplate)
	if err != nil {
		logrus.Errorln(err)
		return manifests
	}

	for _, depStr := range deps {
		dep, err := parsePEP508(depStr)
		if err != nil {
			logrus.Warningf("skipping dependency %q from %q: %s", depStr, relativeFile, err)
			continue
		}

		if len(p.spec.Ignore) > 0 && p.spec.Ignore.isMatchingRules(p.rootDir, relativeFile, dep.Name, dep.Version) {
			logrus.Debugf("ignoring %q from %q as matching ignore rule(s)", dep.Name, relativeFile)
			continue
		}

		if len(p.spec.Only) > 0 && !p.spec.Only.isMatchingRules(p.rootDir, relativeFile, dep.Name, dep.Version) {
			logrus.Debugf("ignoring %q from %q as not matching only rule(s)", dep.Name, relativeFile)
			continue
		}

		params := p.buildTemplateParams(dep, group, lockSupport, projectName, workdir)

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, params); err != nil {
			logrus.Debugln(err)
			continue
		}

		manifests = append(manifests, buf.Bytes())
	}

	return manifests
}

// buildTemplateParams constructs the manifestTemplateParams for a single dependency.
func (p Pyproject) buildTemplateParams(
	dep pythonDependency,
	group string,
	lockSupport lockFileSupport,
	projectName string,
	workdir string,
) manifestTemplateParams {
	// Build a human-readable manifest name that includes the optional group when set.
	var manifestName string
	if group == "" {
		manifestName = fmt.Sprintf("deps(pypi): bump %q for %q project", dep.Name, projectName)
	} else {
		manifestName = fmt.Sprintf("deps(pypi): bump %q [%s] for %q project", dep.Name, group, projectName)
	}

	// Determine version filter.
	//
	// Priority:
	//   1. User-specified VersionFilter from spec.
	//   2. Constraint derived from the dependency itself (e.g. ">=2.28").
	//   3. Wildcard "*" when no version information is present.
	sourceVersionFilterKind := p.versionFilter.Kind
	sourceVersionFilterPattern := p.versionFilter.Pattern
	sourceVersionFilterRegex := p.versionFilter.Regex

	if !p.spec.VersionFilter.IsZero() && dep.Version != "" {
		var err error
		sourceVersionFilterPattern, err = p.versionFilter.GreaterThanPattern(dep.Version)
		if err != nil {
			logrus.Debugf("building version filter pattern for %q: %s", dep.Name, err)
			sourceVersionFilterPattern = p.versionFilter.Pattern
		}
	}

	if p.spec.VersionFilter.IsZero() {
		if dep.Constraint != "" {
			sourceVersionFilterKind = version.PEP440VERSIONKIND
			sourceVersionFilterPattern = dep.Constraint
		} else {
			sourceVersionFilterKind = version.PEP440VERSIONKIND
			sourceVersionFilterPattern = "*"
		}
	}

	relLockFile := filepath.Join(workdir, "uv.lock")

	return manifestTemplateParams{
		ManifestName:               manifestName,
		ActionID:                   p.actionID,
		SourceID:                   dep.Name,
		SourceName:                 fmt.Sprintf("Get latest %q package version", dep.Name),
		SourceVersionFilterKind:    sourceVersionFilterKind,
		SourceVersionFilterPattern: sourceVersionFilterPattern,
		SourceVersionFilterRegex:   sourceVersionFilterRegex,
		DependencyName:             dep.Name,
		IndexURL:                   p.spec.IndexURL,
		TargetID:                   dep.Name,
		TargetName:                 fmt.Sprintf("deps(pypi): bump %q to {{ source %q }}", dep.Name, dep.Name),
		ScmID:                      p.scmID,
		UvEnabled:                  lockSupport.uv,
		LockFile:                   relLockFile,
		Workdir:                    workdir,
	}
}
