package helm

import (
	"bytes"
	"fmt"
	"path/filepath"
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

	foundChartFiles, err := searchChartFiles(
		h.rootDir,
		ChartValidFiles[:])

	if err != nil {
		return nil, err
	}

	for _, foundChartFile := range foundChartFiles {

		relativeFoundChartFile, err := filepath.Rel(h.rootDir, foundChartFile)
		if err != nil {
			// Jump to the next Helm chart if current failed
			logrus.Errorln(err)
			continue
		}

		chartRelativeMetadataPath := filepath.Dir(relativeFoundChartFile)
		chartName := filepath.Base(chartRelativeMetadataPath)

		// Test if the ignore rule based on path doesn't match
		if len(h.spec.Ignore) > 0 && h.spec.Ignore.isMatchingIgnoreRule(h.rootDir, relativeFoundChartFile) {
			logrus.Debugf("Ignoring Helm Chart %q from %q, as not matching rule(s)\n",
				chartName,
				chartRelativeMetadataPath)
			continue
		}

		// Test if the only rule based on path match
		if len(h.spec.Only) > 0 && !h.spec.Only.isMatchingOnlyRule(h.rootDir, relativeFoundChartFile) {
			logrus.Debugf("Ignoring Helm Chart %q from %q, as not matching rule(s)\n",
				chartName,
				chartRelativeMetadataPath)
			continue
		}

		// Retrieve chart dependencies for each chart
		dependencies, err := getChartMetadata(foundChartFile)
		if err != nil {
			return nil, err
		}

		if dependencies == nil {
			continue
		}

		if len(dependencies.Dependencies) == 0 {
			continue
		}

		deps := *dependencies
		for i, dependency := range deps.Dependencies {

			tmpl, err := template.New("manifest").Parse(dependencyManifest)
			if err != nil {
				logrus.Errorln(err)
				continue
			}

			params := struct {
				ManifestName               string
				ImageName                  string
				ChartName                  string
				DependencyName             string
				DependencyRepository       string
				ConditionID                string
				ConditionKey               string
				FleetBundle                string
				SourceID                   string
				SourceName                 string
				SourceVersionFilterKind    string
				SourceVersionFilterPattern string
				TargetID                   string
				TargetKey                  string
				TargetChartName            string
				TargetFile                 string
				File                       string
				ScmID                      string
			}{
				ManifestName:               fmt.Sprintf("Bump dependency %q for Helm Chart %q", dependency.Name, chartName),
				ChartName:                  chartName,
				DependencyName:             dependency.Name,
				DependencyRepository:       dependency.Repository,
				ConditionID:                dependency.Name,
				ConditionKey:               fmt.Sprintf("dependencies[%d].name", i),
				FleetBundle:                chartName,
				SourceID:                   dependency.Name,
				SourceName:                 fmt.Sprintf("Get latest %q Helm Chart Version", dependency.Name),
				SourceVersionFilterKind:    "semver",
				SourceVersionFilterPattern: "*",
				TargetID:                   dependency.Name,
				TargetKey:                  fmt.Sprintf("dependencies[%d].version", i),
				TargetChartName:            chartRelativeMetadataPath,
				TargetFile:                 filepath.Base(foundChartFile),
				File:                       relativeFoundChartFile,
				ScmID:                      h.scmID,
			}

			manifest := bytes.Buffer{}
			if err := tmpl.Execute(&manifest, params); err != nil {
				return nil, err
			}

			manifests = append(manifests, manifest.Bytes())
		}
	}

	return manifests, nil
}
