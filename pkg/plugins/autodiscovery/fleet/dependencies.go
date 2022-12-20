package fleet

import (
	"fmt"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/plugins/resources/helm"
	"github.com/updatecli/updatecli/pkg/plugins/resources/yaml"
)

var (
	// FleetBundleFiles specifies accepted Helm chart metadata file name
	FleetBundleFiles [2]string = [2]string{"fleet.yaml", "fleet.yml"}
)

// Dependency specify the fleetHelmData information that we are looking for in Fleet bundle
type fleetHelmData struct {
	Chart   string
	Repo    string
	Version string
}

// fleetMetada is the information that we need to retrieve from Helm chart files.
type fleetMetada struct {
	Helm fleetHelmData
}

func (f Fleet) discoverFleetDependenciesManifests() ([]config.Spec, error) {

	var manifests []config.Spec

	foundFleetBundleFiles, err := searchFleetBundleFiles(
		f.rootDir,
		FleetBundleFiles[:])

	if err != nil {
		return nil, err
	}

	for _, foundFleetBundleFile := range foundFleetBundleFiles {

		relativeFoundChartFile, err := filepath.Rel(f.rootDir, foundFleetBundleFile)
		if err != nil {
			// Let's try the next chart if one fail
			logrus.Errorln(err)
			continue
		}

		chartRelativeMetadataPath := filepath.Dir(relativeFoundChartFile)
		chartName := filepath.Base(chartRelativeMetadataPath)

		// Test if the ignore rule based on path is respected
		if len(f.spec.Ignore) > 0 && f.spec.Ignore.isMatchingIgnoreRule(f.rootDir, relativeFoundChartFile) {
			logrus.Debugf("Ignoring Helm Chart %q from %q, as not matching rule(s)\n",
				chartName,
				chartRelativeMetadataPath)
			continue
		}

		// Test if the only rule based on path is respected
		if len(f.spec.Only) > 0 && !f.spec.Only.isMatchingOnlyRule(f.rootDir, relativeFoundChartFile) {
			logrus.Debugf("Ignoring Helm Chart %q from %q, as not matching rule(s)\n",
				chartName,
				chartRelativeMetadataPath)
			continue
		}

		// Retrieve chart dependencies for each chart

		data, err := getFleetBundleData(foundFleetBundleFile)
		if err != nil {
			return nil, err
		}

		if data == nil {
			continue
		}

		// Skip pipeline if at least of the helm chart or helm repository is not specified
		if len(data.Helm.Chart) == 0 || len(data.Helm.Repo) == 0 {
			continue
		}

		sourceID := data.Helm.Chart
		conditionID := data.Helm.Chart
		targetID := data.Helm.Chart

		manifestName := fmt.Sprintf("Bump Fleet Bundle %q for Helm Chart %q", chartName, data.Helm.Chart)

		manifest := config.Spec{
			Name: manifestName,
			Sources: map[string]source.Config{
				sourceID: {
					ResourceConfig: resource.ResourceConfig{
						Name: fmt.Sprintf("Get latest %q Helm Chart Version", data.Helm.Chart),
						Kind: "helmchart",
						Spec: helm.Spec{
							Name: data.Helm.Chart,
							URL:  data.Helm.Repo,
						},
					},
				},
			},
			Conditions: map[string]condition.Config{
				conditionID + "-name": {
					DisableSourceInput: true,
					ResourceConfig: resource.ResourceConfig{
						Name: fmt.Sprintf("Ensure Helm chart name %q is specified", data.Helm.Chart),
						Kind: "yaml",
						Spec: yaml.Spec{
							File:  relativeFoundChartFile,
							Key:   "helm.chart",
							Value: data.Helm.Chart,
						},
					},
				},
				conditionID + "-repository": {
					DisableSourceInput: true,
					ResourceConfig: resource.ResourceConfig{
						Name: fmt.Sprintf("Ensure Helm chart repository %q is specified", data.Helm.Repo),
						Kind: "yaml",
						Spec: yaml.Spec{
							File:  relativeFoundChartFile,
							Key:   "helm.repo",
							Value: data.Helm.Repo,
						},
					},
				},
			},
			Targets: map[string]target.Config{
				targetID: {
					SourceID: sourceID,
					ResourceConfig: resource.ResourceConfig{
						Name: fmt.Sprintf("Bump chart %q from Fleet bundle %q", data.Helm.Chart, chartName),
						Kind: "yaml",
						Spec: helm.Spec{
							File: relativeFoundChartFile,
							Key:  "helm.version",
						},
					},
				},
			},
		}
		// Set scmID if defined
		if f.scmID != "" {
			t := manifest.Targets[targetID]
			t.SCMID = f.scmID
			manifest.Targets[targetID] = t

			for _, id := range []string{conditionID + "-name", conditionID + "-repository"} {
				c := manifest.Conditions[id]
				c.SCMID = f.scmID
				manifest.Conditions[id] = c
			}
		}
		manifests = append(manifests, manifest)

	}

	return manifests, nil
}
