package bazel

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Bazel struct holds all information needed to generate Bazel module manifests.
type Bazel struct {
	// actionID holds the actionID used by the newly generated manifest
	actionID string
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootdir defines the root directory from where looking for MODULE.bazel files
	rootDir string
	// scmID holds the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New returns a new valid Bazel object.
func New(spec interface{}, rootDir, scmID, actionID string) (Bazel, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Bazel{}, err
	}

	// Validate ignore rules
	if err := s.Ignore.Validate(); err != nil {
		return Bazel{}, fmt.Errorf("invalid ignore spec: %w", err)
	}

	// Validate only rules
	if err := s.Only.Validate(); err != nil {
		return Bazel{}, fmt.Errorf("invalid only spec: %w", err)
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		logrus.Debugln("no versioning filtering specified, fallback to semantic versioning")
		// By default, Bazel versioning uses semantic versioning
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	dir := rootDir
	if path.IsAbs(s.RootDir) {
		if scmID != "" {
			logrus.Warningf("rootdir %q is an absolute path, scmID %q will be ignored", s.RootDir, scmID)
		}
		dir = s.RootDir
	}

	// Fallback to the current process path if not rootdir specified.
	if len(dir) == 0 {
		logrus.Errorln("no working directory defined")
		return Bazel{}, fmt.Errorf("no working directory defined")
	}

	return Bazel{
		actionID:      actionID,
		spec:          s,
		rootDir:       dir,
		scmID:         scmID,
		versionFilter: newFilter,
	}, nil
}

// DiscoverManifests discovers Bazel module dependencies and generates Updatecli manifests.
func (b Bazel) DiscoverManifests() ([][]byte, error) {
	logrus.Infof("\n\n%s\n", strings.ToTitle("Bazel"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Bazel")+1))

	manifests, err := b.discoverBazelModuleManifests()
	if err != nil {
		return nil, err
	}

	return manifests, nil
}

// discoverBazelModuleManifests discovers all Bazel module dependencies and generates manifests
func (b Bazel) discoverBazelModuleManifests() ([][]byte, error) {
	var manifests [][]byte

	searchFromDir := b.rootDir
	// If the spec.RootDir is an absolute path, then it has already been set correctly in the New function.
	if b.spec.RootDir != "" && !path.IsAbs(b.spec.RootDir) {
		searchFromDir = filepath.Join(b.rootDir, b.spec.RootDir)
	}

	foundFiles, err := findModuleFiles(searchFromDir)
	if err != nil {
		return nil, err
	}

	for _, foundFile := range foundFiles {
		logrus.Debugf("parsing file %q", foundFile)

		relativeFoundFile, err := filepath.Rel(b.rootDir, foundFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		dependencies, err := parseModuleDependencies(foundFile)
		if err != nil {
			logrus.Debugf("skipping file %q due to: %s", foundFile, err)
			continue
		}

		for _, dep := range dependencies {
			// Test if the ignore rule based on path is respected
			if len(b.spec.Ignore) > 0 {
				if shouldIgnore(dep.Name, dep.Version, foundFile, b.rootDir, b.spec.Ignore) {
					logrus.Debugf("Ignoring module %q from file %q, as matching ignore rule(s)\n", dep.Name, relativeFoundFile)
					continue
				}
			}

			// Test if the only rule based on path is respected
			if len(b.spec.Only) > 0 {
				if !shouldInclude(dep.Name, dep.Version, foundFile, b.rootDir, b.spec.Only) {
					logrus.Debugf("Ignoring module %q from %q, as not matching only rule(s)\n", dep.Name, relativeFoundFile)
					continue
				}
			}

			versionPattern, err := b.versionFilter.GreaterThanPattern(dep.Version)
			if err != nil {
				logrus.Debugf("skipping module %q due to: %s", dep.Name, err)
				continue
			}

			moduleManifest, err := b.getBazelModuleManifest(
				relativeFoundFile,
				dep.Name,
				versionPattern,
			)
			if err != nil {
				logrus.Debugf("skipping module %q due to: %s", dep.Name, err)
				continue
			}

			manifests = append(manifests, moduleManifest)
		}
	}

	logrus.Printf("%v manifests identified", len(manifests))

	return manifests, nil
}

// getBazelModuleManifest generates a manifest for a single Bazel module dependency
func (b Bazel) getBazelModuleManifest(filename, moduleName, versionFilterPattern string) ([]byte, error) {
	tmpl, err := template.New("manifest").Parse(manifestTemplate)
	if err != nil {
		return nil, err
	}

	// Sanitize module name for use in IDs (replace special characters)
	sourceID := sanitizeID(moduleName)
	conditionID := sanitizeID(moduleName)
	targetID := sanitizeID(moduleName)

	params := struct {
		ActionID             string
		ModuleName           string
		ModuleFile           string
		SourceID             string
		ConditionID          string
		TargetID             string
		VersionFilterKind    string
		VersionFilterPattern string
		VersionFilterRegex   string
		ScmID                string
		TargetName           string
	}{
		ActionID:             b.actionID,
		ModuleName:           moduleName,
		ModuleFile:           filename,
		SourceID:             sourceID,
		ConditionID:          conditionID,
		TargetID:             targetID,
		VersionFilterKind:    b.versionFilter.Kind,
		VersionFilterPattern: versionFilterPattern,
		VersionFilterRegex:   b.versionFilter.Regex,
		ScmID:                b.scmID,
		TargetName:           fmt.Sprintf("Bump Bazel module %s to {{ source \"%s\" }}", moduleName, sourceID),
	}

	manifest := bytes.Buffer{}
	if err := tmpl.Execute(&manifest, params); err != nil {
		logrus.Debugln(err)
		return nil, err
	}
	return manifest.Bytes(), nil
}

// sanitizeID sanitizes a string for use as an ID in YAML (replaces special characters)
func sanitizeID(s string) string {
	// Replace common special characters with underscores
	replacer := strings.NewReplacer(
		"-", "_",
		".", "_",
		"/", "_",
		"\\", "_",
	)
	return replacer.Replace(s)
}
