package pyproject

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/semver/v3"
	pep440 "github.com/aquasecurity/go-pep440-version"
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
	// listed in several groups, or by several members of a workspace, at the same version gives
	// a single manifest.
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

// newVulnerabilityScan prepares the vulnerability scan of a pyproject.toml whose uv.lock, if
// any, is in lockDir: next to it, or at the root of its workspace.
// hasUvLock reports whether that uv.lock exists, as detected by detectLockFileSupport.
// trustedRegistries holds the normalized URLs of pypi.org and its trusted mirrors.
// handled is shared by the pyproject.toml files sharing lockDir.
func newVulnerabilityScan(lockDir string, hasUvLock bool, trustedRegistries map[string]bool, handled map[string]bool) *vulnerabilityScan {
	scan := &vulnerabilityScan{
		handled:          handled,
		resolved:         map[string]bool{},
		unresolved:       map[string]string{},
		forked:           map[string]bool{},
		skippedUntrusted: map[string]string{},
	}

	if !hasUvLock {
		return scan
	}

	lock, err := loadUvLockVersions(lockDir, trustedRegistries)
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

	// With a uv.lock, a package it doesn't resolve is unresolved, even when pinned with "==":
	// the lock may be stale, and the package come from a private index or a git repository.
	if s.lock.versions != nil {
		switch versions := s.lock.versions[name]; len(versions) {
		case 0:
			return "", nil
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

	// handledByLock holds, for each uv.lock directory, the packages turned into security manifests.
	handledByLock := map[string]map[string]bool{}

	// A first pass reads every pyproject.toml, so that the declarations sharing a uv.lock are all
	// known before the single manifest updating each of their packages is generated.
	var projects []foundProject
	locks := map[string]*lockState{}
	for _, foundFile := range foundFiles {
		log.Debugf("parsing file %q", foundFile)

		data, err := loadPyprojectData(foundFile)
		if err != nil {
			log.Debugln(err)
			continue
		}

		dir := filepath.Dir(foundFile)

		// A workspace member is locked by the uv.lock at the workspace root, where uv runs. The
		// root may sit above "rootdir", so the search goes up to the root directory.
		lockDir := searchUvLock(dir, p.rootDir)
		if lockDir == "" {
			lockDir = dir
		}

		if locks[lockDir] == nil {
			locks[lockDir] = newLockState()
		}
		locks[lockDir].addConstraints(data)

		projects = append(projects, foundProject{file: foundFile, lockDir: lockDir, data: data})
	}

	for _, project := range projects {
		foundFile, lockDir, data := project.file, project.lockDir, project.data

		lockSupport, skip := detectLockFileSupport(lockDir, p.uvAvailable)
		if skip {
			if p.spec.Vulnerability == nil {
				log.Warningf("skipping %q, uv.lock detected in %q but Updatecli couldn't detect the uv command to update it in case of a pyproject.toml update", foundFile, lockDir)
				continue
			}
			// uv.lock can still be read, so the manifests report vulnerable packages without updating them.
			log.Warningf("uv.lock detected in %q but Updatecli couldn't detect the uv command, so security manifests only report vulnerable packages", lockDir)
		}

		relativeFile, err := filepath.Rel(p.rootDir, foundFile)
		if err != nil {
			log.Debugln(err)
			continue
		}

		workdir, err := filepath.Rel(p.rootDir, lockDir)
		if err != nil {
			log.Debugln(err)
			continue
		}

		projectName := data.Project.Name

		var scan *vulnerabilityScan
		if p.spec.Vulnerability != nil {
			// The members of a workspace share its uv.lock, so a package is updated once for all of them.
			if handledByLock[lockDir] == nil {
				handledByLock[lockDir] = map[string]bool{}
			}
			scan = newVulnerabilityScan(lockDir, lockSupport.uvLock, p.trustedRegistries, handledByLock[lockDir])

			// Without readable uv.lock, neither the current versions nor where packages come
			// from are known, and an exact pin may belong to a private or git package.
			if scan.lockErr != nil {
				log.Warningf("skipping %q: uv.lock can't be read (%s), so the installed versions and their sources are unknown", relativeFile, scan.lockErr)
				continue
			}
		}

		// The installed versions tell the packages uv.lock resolves at several versions, and keep
		// an update held back by the age filter from falling back below the version in use.
		var installed lockedVersions
		if scan == nil && lockSupport.uvLock {
			lock, err := loadUvLockVersions(lockDir, p.trustedRegistries)
			switch {
			case err == nil:
				installed = lock.installed
			case p.ageFilter.NeverDowngrade():
				log.Warningf("skipping %q: uv.lock can't be read (%s), so an update could lower the installed versions", relativeFile, err)
				continue
			default:
				log.Debugf("reading the uv.lock of %q: %s", relativeFile, err)
			}
		}

		file := discoveredFile{
			relativeFile: relativeFile,
			projectName:  projectName,
			workdir:      workdir,
			lockSupport:  lockSupport,
			scan:         scan,
			installed:    installed,
			lock:         locks[lockDir],
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
	// workdir is the uv.lock directory, relative to the root directory, where uv runs: the
	// pyproject.toml directory or, for a workspace member, the workspace root.
	workdir string
	// lockSupport tells which lock files the generated targets update.
	lockSupport lockFileSupport
	// scan is only set in vulnerability mode.
	scan *vulnerabilityScan
	// installed holds the versions uv.lock resolves from a registry. It is set outside
	// vulnerability mode whenever uv.lock can be read: it tells the packages resolved at several
	// versions, and the versions an age filter not allowing a downgrade bounds the filter at.
	installed lockedVersions
	// lock holds what the pyproject.toml files sharing the uv.lock, such as workspace members,
	// share in routine mode.
	lock *lockState
}

// foundProject is a pyproject.toml read by the first discovery pass.
type foundProject struct {
	// file is the pyproject.toml path.
	file string
	// lockDir is the directory of its uv.lock, or its own directory when it has none.
	lockDir string
	// data is the pyproject.toml content.
	data pyprojectTOML
}

// lockState holds what the pyproject.toml files sharing one uv.lock share in routine mode.
type lockState struct {
	// constraints maps the normalized name of a package to the distinct constraints its
	// declarations set, in every table of every pyproject.toml sharing the uv.lock.
	constraints map[string][]string
	// bumped holds the normalized names of the packages a routine manifest already re-locks.
	bumped map[string]bool
}

// newLockState returns an empty lockState.
func newLockState() *lockState {
	return &lockState{constraints: map[string][]string{}, bumped: map[string]bool{}}
}

// addConstraints records the constraints of every dependency the pyproject.toml declares.
func (l *lockState) addConstraints(data pyprojectTOML) {
	requirements := append([]string{}, data.Project.Dependencies...)
	for _, group := range data.Project.OptionalDependencies {
		requirements = append(requirements, group...)
	}
	// The second pass reports unsupported group values and entries, so they are left out silently.
	for _, group := range data.DependencyGroups {
		entries, _ := dependencyGroupEntries(group)
		for _, entry := range entries {
			if requirement, ok := entry.(string); ok {
				requirements = append(requirements, requirement)
			}
		}
	}

	for _, requirement := range requirements {
		dep, err := parsePEP508(requirement)
		if err != nil || dep.Constraint == "" {
			continue
		}
		name := pep503.NormalizeName(dep.Name)
		if !slices.Contains(l.constraints[name], dep.Constraint) {
			l.constraints[name] = append(l.constraints[name], dep.Constraint)
		}
	}
}

// combinedConstraint returns the constraint accepting the versions every declaration of the
// package accepts, as uv.lock must satisfy them all: their PEP 440 specifiers joined by commas.
// Environment markers aren't read, so a declaration uv doesn't apply, such as one for a Python
// version the project doesn't support, would restrict it too. When uv.lock resolves the
// package at locked, a single version, only the declarations it satisfies are kept, as uv's
// resolution tells the ones that apply.
func (l *lockState) combinedConstraint(name, locked string) string {
	var lockedVersion pep440.Version
	checkLocked := false
	if v, err := pep440.Parse(locked); locked != "" && err == nil {
		lockedVersion, checkLocked = v, true
	}

	var constraints []string
	for _, constraint := range l.constraints[pep503.NormalizeName(name)] {
		if checkLocked {
			specifiers, err := pep440.NewSpecifiers(constraint, pep440.WithPreRelease(true))
			if err == nil && !specifiers.Check(lockedVersion) {
				log.Debugf("leaving out the constraint %q of %q, which the version %q uv.lock resolves doesn't satisfy", constraint, name, locked)
				continue
			}
		}
		constraints = append(constraints, constraint)
	}
	sort.Strings(constraints)
	return strings.Join(constraints, ",")
}

// installedVersion returns the version of a dependency an update must not lower: the highest
// version uv.lock resolves within the declared constraint or, when uv.lock resolves none, the
// version pinned with "==". It returns "" when neither is known, including when uv.lock only
// resolves versions outside the constraint, which then stays unbounded, as npm does.
func (f discoveredFile) installedVersion(dep pythonDependency) string {
	if _, locked := f.installed[pep503.NormalizeName(dep.Name)]; locked {
		return f.installed.highestSatisfying(dep.Name, dep.Constraint)
	}
	return exactPinnedVersion(dep.Constraint)
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

	// One uv.lock pins one version of a package, so a single manifest re-locks it, rather than
	// one per group or workspace member opening conflicting pull requests. Its version must
	// satisfy every declaration sharing the uv.lock.
	if file.lockSupport.uv {
		name := pep503.NormalizeName(dep.Name)
		if file.lock.bumped[name] {
			log.Debugf("skipping %q from %q, already updated in the uv.lock of %q", dep.Name, file.relativeFile, file.workdir)
			return nil
		}
		file.lock.bumped[name] = true

		// "uv lock --upgrade-package" applies one version to every fork, which fails when the
		// declarations, split by environment markers, require different ones.
		versions := file.installed[name]
		if len(versions) > 1 {
			log.Warningf("skipping %q from %q: uv.lock resolves it at several versions (%s), which a single upgrade can't target",
				dep.Name, file.relativeFile, strings.Join(versions, ", "))
			return nil
		}

		locked := ""
		if len(versions) == 1 {
			locked = versions[0]
		}
		dep.Constraint = file.lock.combinedConstraint(name, locked)
	}

	params, ok := p.buildTemplateParams(file, dep, deps.label())
	if !ok {
		return nil
	}

	return renderManifest(routineManifestTmpl, params)
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
// It returns false when no manifest must be generated for the dependency.
func (p Pyproject) buildTemplateParams(file discoveredFile, dep pythonDependency, label string) (manifestTemplateParams, bool) {
	// Build a human-readable manifest name that includes the group when set.
	var manifestName string
	if label == "" {
		manifestName = fmt.Sprintf("deps(pypi): bump %q for %q project", dep.Name, file.projectName)
	} else {
		manifestName = fmt.Sprintf("deps(pypi): bump %q [%s] for %q project", dep.Name, label, file.projectName)
	}

	sourceVersionFilterKind, sourceVersionFilterPattern, sourceVersionFilterRegex, ok := p.routineVersionFilter(file, dep)
	if !ok {
		return manifestTemplateParams{}, false
	}

	return manifestTemplateParams{
		ManifestName:               manifestName,
		ActionID:                   p.actionID,
		SourceID:                   dep.Name,
		SourceName:                 fmt.Sprintf("Get latest %q package version", dep.Name),
		SourceVersionFilterKind:    sourceVersionFilterKind,
		SourceVersionFilterPattern: sourceVersionFilterPattern,
		SourceVersionFilterRegex:   sourceVersionFilterRegex,
		SourceAge:                  p.ageFilter.Spec,
		DependencyName:             dep.Name,
		IndexURL:                   p.spec.IndexURL,
		TargetID:                   dep.Name,
		TargetName:                 bumpTargetName(dep.Name),
		ScmID:                      p.scmID,
		UvEnabled:                  file.lockSupport.uv,
		LockFile:                   uvLockFile,
		Workdir:                    file.workdir,
	}, true
}

// routineVersionFilter returns the version filter of the source looking for a newer version of a
// dependency, outside security updates.
//
// Priority:
//  1. User-specified VersionFilter from spec.
//  2. Constraint derived from the dependency itself (e.g. ">=2.28").
//  3. Wildcard "*" when no version information is present.
//
// Unless the age filter allows a downgrade, the filter never accepts a version lower than the
// installed one when it's known, so a source whose newer versions are all too recent finds
// nothing rather than a downgrade. When the filter can't be bounded at the installed version,
// it returns false, as the update could then lower it.
func (p Pyproject) routineVersionFilter(file discoveredFile, dep pythonDependency) (kind, pattern, regex string, ok bool) {
	installed := ""
	if p.ageFilter.NeverDowngrade() {
		installed = file.installedVersion(dep)
	}

	filter := p.versionFilter
	// A user versionfilter applies from the version the constraint starts from, and from none
	// when it has no lower bound, as in "foo<3.0". The declared constraint is used as is.
	current := declaredLowerBound(dep.Constraint)
	// A pep440 filter is only bounded at a version known to be installed.
	if filter.Kind == version.PEP440VERSIONKIND {
		current = ""
	}
	if p.spec.VersionFilter.IsZero() {
		filter = version.Filter{Kind: version.PEP440VERSIONKIND, Pattern: "*", Regex: p.versionFilter.Regex}
		if dep.Constraint != "" {
			filter.Pattern = dep.Constraint
		}
		current = ""
	}
	if installed != "" {
		current = installed
	}
	// A relative pattern, such as "minor", means nothing without a version to start from.
	isRelative := filter.Kind == version.SEMVERVERSIONKIND && relativeSemverPatterns[filter.Pattern]
	if current == "" {
		if isRelative {
			log.Warningf("skipping %q from %q: the %q versionfilter pattern needs a version to start from, which neither the constraint %q nor uv.lock gives",
				dep.Name, file.relativeFile, filter.Pattern, dep.Constraint)
			return "", "", "", false
		}
		return filter.Kind, filter.Pattern, filter.Regex, true
	}

	greaterThanPattern, err := p.ageFilter.GreaterThanPattern(&filter, current)
	// A semver pattern which can't parse the installed version is left unbounded, without error.
	if err == nil && installed != "" && filter.Kind == version.SEMVERVERSIONKIND {
		if _, semverErr := semver.NewVersion(installed); semverErr != nil {
			err = fmt.Errorf("%q is not a semantic version", installed)
		}
	}

	switch {
	case err == nil:
		return filter.Kind, greaterThanPattern, filter.Regex, true
	case installed != "":
		log.Warningf("skipping %q from %q: the %q versionfilter can't be bounded at the installed version %q (%s), so an update held back by the age filter could lower it. Set \"allowdowngrade\" to true to allow it",
			dep.Name, file.relativeFile, filter.Kind, installed, err)
		return "", "", "", false
	case isRelative:
		log.Warningf("skipping %q from %q: the %q versionfilter pattern can't start from %q (%s)",
			dep.Name, file.relativeFile, filter.Pattern, current, err)
		return "", "", "", false
	default:
		log.Debugf("building version filter pattern for %q: %s", dep.Name, err)
		return filter.Kind, filter.Pattern, filter.Regex, true
	}
}
