package pyproject

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// pyprojectTOML mirrors the subset of pyproject.toml we care about.
type pyprojectTOML struct {
	Project struct {
		Name                 string              `toml:"name"`
		Dependencies         []string            `toml:"dependencies"`
		OptionalDependencies map[string][]string `toml:"optional-dependencies"`
	} `toml:"project"`
	// DependencyGroups holds PEP 735 groups. Entries are either PEP 508 strings
	// or {include-group = "..."} tables. Values are decoded loosely, so a malformed group
	// is skipped on its own instead of making the whole pyproject.toml unreadable.
	DependencyGroups map[string]interface{} `toml:"dependency-groups"`
}

// dependencyGroupEntries returns the entries of a PEP 735 dependency group, and false when its
// value isn't an array. An array only made of tables, such as include-group entries written
// with the [[dependency-groups.<name>]] syntax, decodes to a different type than a mixed one.
func dependencyGroupEntries(group interface{}) ([]interface{}, bool) {
	switch g := group.(type) {
	case []interface{}:
		return g, true
	case []map[string]interface{}:
		entries := make([]interface{}, 0, len(g))
		for _, table := range g {
			entries = append(entries, table)
		}
		return entries, true
	default:
		return nil, false
	}
}

// dependencyGroupMemberships returns, for each PEP 735 dependency group, the groups its
// dependencies belong to: the group itself and every group including it, directly or not,
// through {include-group = "..."} entries. Names are compared normalized, as PEP 735 requires,
// and returned as declared, sorted. Include cycles, which PEP 735 forbids, don't loop.
func dependencyGroupMemberships(groups map[string]interface{}) map[string][]string {
	declared := map[string]string{}
	for name := range groups {
		declared[normalizeName(name)] = name
	}

	// includedBy maps a normalized group name to the normalized names of the groups including it.
	includedBy := map[string][]string{}
	for name, group := range groups {
		entries, ok := dependencyGroupEntries(group)
		if !ok {
			continue
		}
		for _, entry := range entries {
			table, ok := entry.(map[string]interface{})
			if !ok {
				continue
			}
			included, ok := table["include-group"].(string)
			if !ok {
				continue
			}
			child := normalizeName(included)
			includedBy[child] = append(includedBy[child], normalizeName(name))
		}
	}

	memberships := map[string][]string{}
	for name := range groups {
		seen := map[string]bool{}
		queue := []string{normalizeName(name)}
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			if seen[current] {
				continue
			}
			seen[current] = true
			queue = append(queue, includedBy[current]...)
		}

		var members []string
		for normalized := range seen {
			if d, found := declared[normalized]; found {
				members = append(members, d)
			}
		}
		sort.Strings(members)
		memberships[name] = members
	}

	return memberships
}

// dependencyGroupRequirements returns the PEP 508 strings of a PEP 735 dependency group.
// include-group tables are skipped: the included group is processed on its own,
// so following them would only generate duplicate manifests.
// A group whose value isn't an array, which PEP 735 requires, is skipped.
func dependencyGroupRequirements(group interface{}) []string {
	entries, ok := dependencyGroupEntries(group)
	if !ok {
		log.Warningf("skipping dependency group with unsupported value %v, PEP 735 expects an array", group)
		return nil
	}

	var requirements []string
	for _, entry := range entries {
		switch e := entry.(type) {
		case string:
			requirements = append(requirements, e)
		case map[string]interface{}:
			if _, ok := e["include-group"]; !ok {
				log.Debugf("skipping unsupported dependency group entry %v", e)
			}
		default:
			log.Debugf("skipping unsupported dependency group entry %v", e)
		}
	}
	return requirements
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
		log.Debugf("parsing file %q", foundFile)

		dir := filepath.Dir(foundFile)

		lockSupport, skip := detectLockFileSupport(dir, p.uvAvailable)
		if skip {
			continue
		}

		data, err := loadPyprojectData(foundFile)
		if err != nil {
			log.Debugln(err)
			continue
		}

		relativeFile, err := filepath.Rel(p.rootDir, foundFile)
		if err != nil {
			log.Debugln(err)
			continue
		}

		workdir, err := filepath.Rel(p.rootDir, dir)
		if err != nil {
			log.Debugln(err)
			continue
		}

		projectName := data.Project.Name

		// Process [project.dependencies] — the main dependency group.
		mainDeps := make([]string, len(data.Project.Dependencies))
		copy(mainDeps, data.Project.Dependencies)
		sort.Strings(mainDeps)

		manifests = append(manifests, p.processDependencies(mainDeps, "", "", nil, relativeFile, lockSupport, projectName, workdir)...)

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

			manifests = append(manifests, p.processDependencies(groupDeps, group, "", nil, relativeFile, lockSupport, projectName, workdir)...)
		}

		// Process each [dependency-groups] group (PEP 735) in deterministic order.
		depGroups := make([]string, 0, len(data.DependencyGroups))
		for g := range data.DependencyGroups {
			depGroups = append(depGroups, g)
		}
		sort.Strings(depGroups)
		memberships := dependencyGroupMemberships(data.DependencyGroups)

		for _, group := range depGroups {
			// The name ends up in the generated manifest, which a quote would break.
			if !dependencyGroupNameRegex.MatchString(group) {
				log.Warningf("skipping dependency group %q from %q, PEP 735 only allows letters, digits, \".\", \"_\" and \"-\" in group names", group, relativeFile)
				continue
			}

			groupDeps := dependencyGroupRequirements(data.DependencyGroups[group])
			sort.Strings(groupDeps)

			manifests = append(manifests, p.processDependencies(groupDeps, "", group, memberships[group], relativeFile, lockSupport, projectName, workdir)...)
		}
	}

	log.Printf("%v manifests identified", len(manifests))

	return manifests, nil
}

// processDependencies generates manifests for a list of PEP 508 dependency strings.
// extra is the optional-dependency group name and dependencyGroup the PEP 735 group name;
// both are "" for main dependencies.
// memberGroups holds the dependency groups the dependencies belong to, which includes the
// groups including dependencyGroup, as matched by dependencygroups rules.
func (p Pyproject) processDependencies(
	deps []string,
	extra string,
	dependencyGroup string,
	memberGroups []string,
	relativeFile string,
	lockSupport lockFileSupport,
	projectName string,
	workdir string,
) [][]byte {
	var manifests [][]byte

	// The "group:" prefix keeps dependency groups apart from optional-dependency groups of the same name.
	group := extra
	if dependencyGroup != "" {
		group = "group:" + dependencyGroup
	}

	tmpl, err := template.New("manifest").Parse(age.ManifestTemplate + manifestTemplate)
	if err != nil {
		log.Errorln(err)
		return manifests
	}

	for _, depStr := range deps {
		dep, err := parsePEP508(depStr)
		if err != nil {
			log.Warningf("skipping dependency %q from %q: %s", depStr, relativeFile, err)
			continue
		}

		if len(p.spec.Ignore) > 0 && p.spec.Ignore.isMatchingRules(p.rootDir, relativeFile, dep.Name, dep.Version, memberGroups) {
			log.Debugf("ignoring %q from %q as matching ignore rule(s)", dep.Name, relativeFile)
			continue
		}

		if len(p.spec.Only) > 0 && !p.spec.Only.isMatchingRules(p.rootDir, relativeFile, dep.Name, dep.Version, memberGroups) {
			log.Debugf("ignoring %q from %q as not matching only rule(s)", dep.Name, relativeFile)
			continue
		}

		params := p.buildTemplateParams(dep, group, lockSupport, projectName, workdir)

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, params); err != nil {
			log.Debugln(err)
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
			log.Debugf("building version filter pattern for %q: %s", dep.Name, err)
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

	return manifestTemplateParams{
		ManifestName:               manifestName,
		ActionID:                   p.actionID,
		SourceID:                   dep.Name,
		SourceName:                 fmt.Sprintf("Get latest %q package version", dep.Name),
		SourceVersionFilterKind:    sourceVersionFilterKind,
		SourceVersionFilterPattern: sourceVersionFilterPattern,
		SourceVersionFilterRegex:   sourceVersionFilterRegex,
		SourceAge:                  p.releaseAge,
		DependencyName:             dep.Name,
		IndexURL:                   p.spec.IndexURL,
		TargetID:                   dep.Name,
		TargetName:                 fmt.Sprintf("deps(pypi): bump %q to {{ source %q }}", dep.Name, dep.Name),
		ScmID:                      p.scmID,
		UvEnabled:                  lockSupport.uv,
		LockFile:                   "uv.lock",
		Workdir:                    workdir,
	}
}
