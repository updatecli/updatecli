package helm

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
)

var (
	// ChartValidFiles specifies accepted Helm chart metadata filename
	ChartValidFiles [2]string = [2]string{"Chart.yaml", "Chart.yml"}
)

// Dependency specify the dependency information.
type dependency struct {
	Name       string
	Repository string
	Version    string
}

// chartMetadata is the information fetches from a Chart.yaml that Updatecli needs to identify update.
type chartMetadata struct {
	Name         string
	Dependencies []dependency
}

func (h Helm) discoverHelmDependenciesManifests() ([][]byte, error) {

	var manifests [][]byte

	searchFromDir := h.rootDir
	// If the spec.RootDir is an absolute path, then it as already been set
	// correctly in the New function.
	if h.spec.RootDir != "" && !path.IsAbs(h.spec.RootDir) {
		searchFromDir = filepath.Join(h.rootDir, h.spec.RootDir)
	}

	foundChartFiles, err := searchChartFiles(
		searchFromDir,
		ChartValidFiles[:])

	if err != nil {
		return nil, err
	}

	for _, foundChartFile := range foundChartFiles {
		logrus.Debugf("parsing file %q", foundChartFile)

		relativeFoundChartFile, err := filepath.Rel(h.rootDir, foundChartFile)
		if err != nil {
			// Jump to the next Helm chart if current failed
			logrus.Debugln(err)
			continue
		}

		chartRelativeMetadataPath := filepath.Dir(relativeFoundChartFile)
		chartName := filepath.Base(chartRelativeMetadataPath)

		// Retrieve chart dependencies for each chart
		dependencies, err := getChartMetadata(foundChartFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		if dependencies == nil {
			continue
		}

		if len(dependencies.Dependencies) == 0 {
			continue
		}

		deps := *dependencies
		for i, dependency := range deps.Dependencies {

			sourceVersionFilterKind := "semver"
			sourceVersionFilterPattern := "*"
			sourceVersionFilterRegex := "*"

			if strings.HasPrefix(dependency.Repository, "file://") || dependency.Repository == "" {
				logrus.Debugf("Ignoring dependency %q for chart %q as it is a local dependency\n", chartName, dependency.Name)
				continue
			}

			if !h.spec.VersionFilter.IsZero() {
				sourceVersionFilterKind = h.versionFilter.Kind
				sourceVersionFilterPattern, err = h.versionFilter.GreaterThanPattern(dependency.Version)
				sourceVersionFilterRegex = h.versionFilter.Regex
				if err != nil {
					logrus.Debugf("building version filter pattern: %s", err)
					sourceVersionFilterPattern = "*"
				}
			}

			// Test if the ignore rule based on path is respected
			if len(h.spec.Ignore) > 0 {
				if h.spec.Ignore.isMatchingRules(h.rootDir, chartRelativeMetadataPath, deps.Dependencies[i].Name, deps.Dependencies[i].Version, "", "") {
					logrus.Debugf("Ignoring Dependency version update from file %q, as matching ignore rule(s)\n", relativeFoundChartFile)
					continue
				}
			}

			// Test if the only rule based on path is respected
			if len(h.spec.Only) > 0 {
				if !h.spec.Only.isMatchingRules(h.rootDir, chartRelativeMetadataPath, deps.Dependencies[i].Name, deps.Dependencies[i].Version, "", "") {
					logrus.Debugf("Ignoring Dependency version update from %q, as not matching only rule(s)\n", relativeFoundChartFile)
					continue
				}
			}

			tmpl, err := template.New("manifest").Parse(dependencyManifest)
			if err != nil {
				logrus.Debugln(err)
				continue
			}

			dependencyNameSlug := strings.ReplaceAll(dependency.Name, "/", "_")

			params := struct {
				ActionID                    string
				ManifestName                string
				ImageName                   string
				ChartName                   string
				DependencyName              string
				DependencyRepository        string
				ConditionID                 string
				ConditionKey                string
				FleetBundle                 string
				SourceName                  string
				SourceVersionFilterKind     string
				SourceVersionFilterPattern  string
				SourceVersionFilterRegex    string
				TargetID                    string
				TargetKey                   string
				TargetChartName             string
				TargetChartSkipPackaging    bool
				TargetChartVersionIncrement string
				TargetFile                  string
				File                        string
				ScmID                       string
			}{
				ActionID:                    h.actionID,
				ManifestName:                fmt.Sprintf("Bump dependency %q for Helm chart %q", dependency.Name, chartName),
				ChartName:                   chartName,
				DependencyName:              dependency.Name,
				DependencyRepository:        dependency.Repository,
				ConditionID:                 dependencyNameSlug,
				ConditionKey:                fmt.Sprintf("$.dependencies[%d].name", i),
				FleetBundle:                 chartName,
				SourceName:                  fmt.Sprintf("Get latest %q Helm chart version", dependency.Name),
				SourceVersionFilterKind:     sourceVersionFilterKind,
				SourceVersionFilterPattern:  sourceVersionFilterPattern,
				SourceVersionFilterRegex:    sourceVersionFilterRegex,
				TargetID:                    dependencyNameSlug,
				TargetKey:                   fmt.Sprintf("$.dependencies[%d].version", i),
				TargetChartName:             chartRelativeMetadataPath,
				TargetChartSkipPackaging:    h.spec.SkipPackaging,
				TargetChartVersionIncrement: h.spec.VersionIncrement,
				TargetFile:                  filepath.Base(foundChartFile),
				File:                        relativeFoundChartFile,
				ScmID:                       h.scmID,
			}

			manifest := bytes.Buffer{}
			if err := tmpl.Execute(&manifest, params); err != nil {
				logrus.Debugln(err)
				continue
			}

			manifests = append(manifests, manifest.Bytes())
		}
	}

	return manifests, nil
}
