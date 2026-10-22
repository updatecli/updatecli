package pyproject

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pep503"
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
// through {include-group = "..."} entries. Names are compared and returned normalized, as
// PEP 735 requires, sorted. Include cycles, which PEP 735 forbids, don't loop.
func dependencyGroupMemberships(groups map[string]interface{}) map[string][]string {
	declared := map[string]bool{}
	for name := range groups {
		declared[pep503.NormalizeName(name)] = true
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
			child := pep503.NormalizeName(included)
			includedBy[child] = append(includedBy[child], pep503.NormalizeName(name))
		}
	}

	memberships := map[string][]string{}
	for name := range groups {
		seen := map[string]bool{}
		queue := []string{pep503.NormalizeName(name)}
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
			if declared[normalized] {
				members = append(members, normalized)
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

// vulnerabilityScan tracks the dependencies of one pyproject.toml in vulnerability mode.
type vulnerabilityScan struct {
	// lock holds what uv.lock resolved, its zero value without a readable uv.lock.
	// Packages not resolved from pypi.org or a trusted mirror are never scanned.
	lock lockedPackages
	// skippedUntrusted maps the normalized names skipped for coming from an untrusted registry
	// to their declared name, so they are reported at once.
	skippedUntrusted map[string]string
	// lockErr reports a uv.lock present but unreadable.
	lockErr error
	// handled holds the "name==version" pairs already turned into a manifest, so a package
	// listed in several groups at the same version gives a single manifest.
	handled map[string]bool
	// resolved holds the normalized names for which a current version was found at least once.
	resolved map[string]bool
	// unresolved maps the normalized names dropped for lack of a current version to their
	// declared name, so the user is told about them at once rather than getting a silently
	// incomplete report.
	unresolved map[string]string
	// forked holds the normalized names already reported as locked at several versions.
	forked map[string]bool
}

// newVulnerabilityScan prepares the vulnerability scan of the pyproject.toml in dir.
// hasUvLock reports whether a uv.lock sits next to it, as detected by detectLockFileSupport.
// trustedRegistries holds the normalized URLs of pypi.org and its trusted mirrors.
func newVulnerabilityScan(dir string, hasUvLock bool, trustedRegistries map[string]bool) *vulnerabilityScan {
	scan := &vulnerabilityScan{
		handled:          map[string]bool{},
		resolved:         map[string]bool{},
		unresolved:       map[string]string{},
		forked:           map[string]bool{},
		skippedUntrusted: map[string]string{},
	}

	if !hasUvLock {
		return scan
	}

	lock, err := loadUvLockVersions(dir, trustedRegistries)
	if err != nil {
		scan.lockErr = err
		return scan
	}
	scan.lock = lock

	return scan
}

// currentVersion returns the version of a dependency to check against the OSV database:
// the version resolved in uv.lock or, when uv.lock has none, the version pinned in pyproject.toml.
// It returns "" when no current version can be identified, along with the versions uv.lock
// resolves when there are several of them. It doesn't record nor report anything, so that
// only and ignore rules can be evaluated against the version before it is accounted for.
func (s *vulnerabilityScan) currentVersion(dep pythonDependency) (version string, forkedVersions []string) {
	name := pep503.NormalizeName(dep.Name)
	if _, untrusted := s.lock.untrustedRegistry[name]; untrusted || s.lock.notFromRegistry[name] {
		// Not a PyPI release, even when pinned with "==" in pyproject.toml.
		return "", nil
	}

	if s.lock.versions != nil {
		switch versions := s.lock.versions[name]; len(versions) {
		case 0:
		case 1:
			return versions[0], nil
		default:
			return "", versions
		}
	}

	return exactPinnedVersion(dep.Constraint), nil
}

// markUnresolved records a dependency dropped for lack of a current version.
// A package uv.lock resolves at several versions gets its own warning, once.
func (s *vulnerabilityScan) markUnresolved(dep pythonDependency, relativeFile string, forkedVersions []string) {
	name := pep503.NormalizeName(dep.Name)

	if s.lock.notFromRegistry[name] {
		log.Debugf("ignoring %q from %q, as uv.lock doesn't resolve it from a registry", dep.Name, relativeFile)
		return
	}

	if _, untrusted := s.lock.untrustedRegistry[name]; untrusted {
		s.skippedUntrusted[name] = dep.Name
		return
	}

	if len(forkedVersions) > 0 {
		if !s.forked[name] {
			s.forked[name] = true
			log.Warningf("skipping %q from %q: uv.lock resolves it at several versions (%s), which a single upgrade can't target",
				dep.Name, relativeFile, strings.Join(forkedVersions, ", "))
		}
		return
	}

	if _, found := s.unresolved[name]; !found {
		s.unresolved[name] = dep.Name
	}
}

// claim records a dependency at its current version and reports whether a manifest is still
// needed for it, which is false when the same package at the same version was already handled.
func (s *vulnerabilityScan) claim(dep pythonDependency, currentVersion string) bool {
	name := pep503.NormalizeName(dep.Name)
	s.resolved[name] = true

	key := name + "==" + currentVersion
	if s.handled[key] {
		return false
	}
	s.handled[key] = true
	return true
}

// unresolvedMessage describes the dependencies skipped for lack of a current version,
// or returns "" when there are none. A dependency resolved elsewhere in the file, such as
// pinned in another group, isn't reported.
func (s *vulnerabilityScan) unresolvedMessage(relativeFile string) string {
	var names []string
	for name, declared := range s.unresolved {
		if !s.resolved[name] {
			names = append(names, declared)
		}
	}
	if len(names) == 0 {
		return ""
	}

	sort.Strings(names)
	skipped := strings.Join(names, ", ")

	if s.lock.versions == nil {
		return fmt.Sprintf("no uv.lock found for %q, skipping its dependencies not pinned to an exact version: %s", relativeFile, skipped)
	}

	return fmt.Sprintf("no version resolved in uv.lock for some dependencies of %q, skipping: %s", relativeFile, skipped)
}

// untrustedMessage describes the dependencies skipped for being resolved from a registry other
// than pypi.org or a trusted mirror, or returns "" when there are none.
func (s *vulnerabilityScan) untrustedMessage(relativeFile string) string {
	if len(s.skippedUntrusted) == 0 {
		return ""
	}

	var skipped []string
	for name, declared := range s.skippedUntrusted {
		skipped = append(skipped, fmt.Sprintf("%s (%s)", declared, s.lock.untrustedRegistry[name]))
	}
	sort.Strings(skipped)

	return fmt.Sprintf("skipping dependencies of %q resolved from a registry other than pypi.org, as the OSV database only covers pypi.org. List the registry under \"pypimirrors\" if it mirrors pypi.org: %s",
		relativeFile, strings.Join(skipped, ", "))
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
			if p.spec.Vulnerability == nil {
				log.Warning("skipping, uv.lock detected but Updatecli couldn't detect the uv command to update it in case of a pyproject.toml update")
				continue
			}
			// uv.lock can still be read, so the manifests report vulnerable packages without updating them.
			log.Warningf("uv.lock detected in %q but Updatecli couldn't detect the uv command, so security manifests only report vulnerable packages", dir)
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

		var scan *vulnerabilityScan
		if p.spec.Vulnerability != nil {
			scan = newVulnerabilityScan(dir, lockSupport.uvLock, p.trustedRegistries)

			// Without readable uv.lock, neither the current versions nor where packages come
			// from are known, and an exact pin may belong to a private or git package.
			if scan.lockErr != nil {
				log.Warningf("skipping %q: uv.lock can't be read (%s), so the installed versions and their sources are unknown", relativeFile, scan.lockErr)
				continue
			}
		}

		file := discoveredFile{
			relativeFile: relativeFile,
			projectName:  projectName,
			workdir:      workdir,
			lockSupport:  lockSupport,
			scan:         scan,
		}

		// Process [project.dependencies] — the main dependency group.
		mainDeps := make([]string, len(data.Project.Dependencies))
		copy(mainDeps, data.Project.Dependencies)
		sort.Strings(mainDeps)

		manifests = append(manifests, p.processDependencies(file, dependencyList{requirements: mainDeps})...)

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

			manifests = append(manifests, p.processDependencies(file, dependencyList{requirements: groupDeps, extra: group})...)
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

			manifests = append(manifests, p.processDependencies(file, dependencyList{
				requirements: groupDeps,
				group:        group,
				memberGroups: memberships[group],
			})...)
		}

		if scan != nil {
			if msg := scan.untrustedMessage(relativeFile); msg != "" {
				log.Warningln(msg)
			}
			if msg := scan.unresolvedMessage(relativeFile); msg != "" {
				log.Warningln(msg)
			}
		}
	}

	log.Printf("%v manifests identified", len(manifests))

	return manifests, nil
}

// discoveredFile holds what the manifests generated for one pyproject.toml share.
type discoveredFile struct {
	// relativeFile is the pyproject.toml path, relative to the root directory.
	relativeFile string
	// projectName is the [project] name, used in manifest names.
	projectName string
	// workdir is the pyproject.toml directory, relative to the root directory, where uv runs.
	workdir string
	// lockSupport tells which lock files the generated targets update.
	lockSupport lockFileSupport
	// scan is only set in vulnerability mode.
	scan *vulnerabilityScan
}

// dependencyList is a list of dependencies declared by one table of a pyproject.toml.
type dependencyList struct {
	// requirements holds the PEP 508 dependency strings.
	requirements []string
	// extra is the [project.optional-dependencies] group name, or "".
	extra string
	// group is the [dependency-groups] (PEP 735) group name, or "".
	group string
	// memberGroups holds the normalized names of the dependency groups the dependencies belong
	// to, including the groups including group, as matched by dependencygroups rules.
	memberGroups []string
}

// label returns the name of the list as shown in manifest names, or "" for the main dependencies.
// The "group:" prefix keeps dependency groups apart from optional-dependency groups of the same name.
func (l dependencyList) label() string {
	if l.group != "" {
		return "group:" + l.group
	}
	return l.extra
}

// processDependencies generates manifests for a list of PEP 508 dependency strings:
// security manifests in vulnerability mode, routine ones otherwise.
func (p Pyproject) processDependencies(file discoveredFile, deps dependencyList) [][]byte {
	var manifests [][]byte

	for _, depStr := range deps.requirements {
		dep, err := parsePEP508(depStr)
		if err != nil {
			log.Warningf("skipping dependency %q from %q: %s", depStr, file.relativeFile, err)
			continue
		}

		var manifest []byte
		if file.scan != nil {
			manifest = p.securityManifest(file, deps, dep)
		} else {
			manifest = p.routineManifest(file, deps, dep)
		}

		if manifest != nil {
			manifests = append(manifests, manifest)
		}
	}

	return manifests
}

// routineManifest returns the manifest bumping a dependency to its latest version,
// or nil when only and ignore rules leave it out.
func (p Pyproject) routineManifest(file discoveredFile, deps dependencyList, dep pythonDependency) []byte {
	if !p.isSelected(file, deps, dep, dep.Version) {
		return nil
	}

	return renderManifest(routineManifestTmpl, p.buildTemplateParams(file, dep, deps.label()))
}

// securityManifest returns the manifest updating a dependency whose current version has known
// vulnerabilities, or nil when it has no identifiable current version, was already handled
// at that version, or is left out by only and ignore rules.
func (p Pyproject) securityManifest(file discoveredFile, deps dependencyList, dep pythonDependency) []byte {
	currentVersion, forkedVersions := file.scan.currentVersion(dep)

	// Rules match the version checked against the OSV database, the one from uv.lock or the
	// exact pin, rather than the declared lower bound.
	ruleVersion := dep.Version
	if currentVersion != "" {
		ruleVersion = currentVersion
	}
	if !p.isSelected(file, deps, dep, ruleVersion) {
		return nil
	}

	if currentVersion == "" {
		file.scan.markUnresolved(dep, file.relativeFile, forkedVersions)
		return nil
	}
	if !file.scan.claim(dep, currentVersion) {
		return nil
	}

	return renderManifest(securityManifestTmpl, p.securityTemplateParams(file, dep, currentVersion))
}

// isSelected reports whether only and ignore rules keep a dependency at the given version.
func (p Pyproject) isSelected(file discoveredFile, deps dependencyList, dep pythonDependency, ver string) bool {
	if len(p.spec.Ignore) > 0 && p.spec.Ignore.isMatchingRules(p.rootDir, file.relativeFile, dep.Name, ver, deps.memberGroups) {
		log.Debugf("ignoring %q from %q as matching ignore rule(s)", dep.Name, file.relativeFile)
		return false
	}

	if len(p.spec.Only) > 0 && !p.spec.Only.isMatchingRules(p.rootDir, file.relativeFile, dep.Name, ver, deps.memberGroups) {
		log.Debugf("ignoring %q from %q as not matching only rule(s)", dep.Name, file.relativeFile)
		return false
	}

	return true
}

// renderManifest executes a manifest template, returning nil when it fails.
func renderManifest(tmpl *template.Template, params manifestTemplateParams) []byte {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		log.Debugln(err)
		return nil
	}
	return buf.Bytes()
}

// bumpTargetName returns the name of the target bumping a dependency to the version of its source.
func bumpTargetName(dependency string) string {
	return fmt.Sprintf("deps(pypi): bump %q to {{ source %q }}", dependency, dependency)
}

// securityTemplateParams returns the parameters of the security manifest updating a dependency
// whose currentVersion has known vulnerabilities.
func (p Pyproject) securityTemplateParams(file discoveredFile, dep pythonDependency, currentVersion string) manifestTemplateParams {
	return manifestTemplateParams{
		// The pipeline name contains the fixed version, so it can be used as pullrequest title
		ManifestName:   bumpTargetName(dep.Name),
		ActionID:       p.actionID,
		SourceID:       dep.Name,
		SourceName:     fmt.Sprintf("Get lowest %q package version without known vulnerabilities", dep.Name),
		DependencyName: dep.Name,
		TargetID:       dep.Name,
		TargetName:     bumpTargetName(dep.Name),
		ScmID:          p.scmID,
		UvEnabled:      file.lockSupport.uv,
		LockFile:       uvLockFile,
		Workdir:        file.workdir,
		Vulnerability:  p.spec.Vulnerability,
		CurrentVersion: currentVersion,
	}
}

// buildTemplateParams constructs the manifestTemplateParams for a single dependency.
// label is the dependency list label shown in the manifest name, "" for the main dependencies.
func (p Pyproject) buildTemplateParams(file discoveredFile, dep pythonDependency, label string) manifestTemplateParams {
	// Build a human-readable manifest name that includes the group when set.
	var manifestName string
	if label == "" {
		manifestName = fmt.Sprintf("deps(pypi): bump %q for %q project", dep.Name, file.projectName)
	} else {
		manifestName = fmt.Sprintf("deps(pypi): bump %q [%s] for %q project", dep.Name, label, file.projectName)
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
		TargetName:                 bumpTargetName(dep.Name),
		ScmID:                      p.scmID,
		UvEnabled:                  file.lockSupport.uv,
		LockFile:                   uvLockFile,
		Workdir:                    file.workdir,
	}
}
