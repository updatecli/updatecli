package helm

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/pullrequest"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/plugins/resources/helm"
	"github.com/updatecli/updatecli/pkg/plugins/resources/yaml"
)

var (
	// ChartValidFiles specifies accepted Helm chart metadata file name
	ChartValidFiles [2]string = [2]string{"Chart.yaml", "Chart.yml"}
)

// Dependency specify the dependency information that we are looking for in Helm chart
type dependency struct {
	Name       string
	Repository string
	Version    string
}

// chartMetadata is the information that we need to retrieve from Helm chart files.
type chartMetadata struct {
	Name         string
	Dependencies []dependency
}

// Spec defines the parameters which can be provided to the Helm builder.
type Spec struct {
	// RootDir defines the root directory used to recursively search for Helm Chart
	RootDir string `yaml:",omitempty"`
	// Disable allows to disable the helm chart crawler
	Disable bool `yaml:",omitempty"`
	// Ignore allows to specify rule to ignore autodiscovery a specific Helm based on a rule
	Ignore MatchingRules
	// Only allows to specify rule to only autodiscover manifest for a specific Helm based on a rule
	Only MatchingRules
}

// Helm hold all information needed to generate helm manifest.
type Helm struct {
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for Helm Chart
	rootDir string
}

// New return a new valid Helm object.
func New(spec interface{}, rootDir string) (Helm, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Helm{}, err
	}

	dir := rootDir
	if len(s.RootDir) > 0 {
		dir = s.RootDir
	}

	// If no RootDir have been provided via settings,
	// then fallback to the current process path.
	if len(dir) == 0 {
		logrus.Errorln("no working directrory defined")
		return Helm{}, err
	}

	return Helm{
		spec:    s,
		rootDir: dir,
	}, nil

}

func (h Helm) DiscoverManifests(
	scmSpec *scm.Config,
	scmID string,
	pullrequestSpec *pullrequest.Config,
	pullrequestID string) ([]config.Spec, error) {

	var manifests []config.Spec

	foundChartFiles, err := searchChartMetadataFiles(
		h.rootDir,
		[]string{"Chart.yaml", "Chart.yml"})

	if err != nil {
		return nil, err
	}

	for _, foundChartFile := range foundChartFiles {

		relativeFoundChartFile, err := filepath.Rel(h.rootDir, foundChartFile)
		if err != nil {
			// Let's try the next chart if one fail
			logrus.Errorln(err)
			continue
		}

		chartRelativeMetadataPath := filepath.Dir(relativeFoundChartFile)
		metadataFilename := filepath.Base(foundChartFile)
		chartName := filepath.Base(chartRelativeMetadataPath)

		// Test if the ignore rule based on path is respected
		if len(h.spec.Ignore) > 0 && h.spec.Ignore.isMatchingIgnoreRule(h.rootDir, relativeFoundChartFile) {
			logrus.Debugf("Ignoring Helm Chart %q from %q, as not matching rule(s)\n",
				chartName,
				chartRelativeMetadataPath)
			continue
		}

		// Test if the only rule based on path is respected
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
			//pipelines = append(pipelines, config.Spec{
			manifest := config.Spec{
				Name: strings.Join([]string{
					chartName,
					dependency.Name,
				}, "-"),
				Sources: map[string]source.Config{
					dependency.Name: {
						ResourceConfig: resource.ResourceConfig{
							Name: fmt.Sprintf("Get latest %q Helm Chart Version", dependency.Name),
							Kind: "helmchart",
							Spec: helm.Spec{
								Name: dependency.Name,
								URL:  dependency.Repository,
							},
						},
					},
				},
				Conditions: map[string]condition.Config{
					dependency.Name: {
						DisableSourceInput: true,
						ResourceConfig: resource.ResourceConfig{
							Name: fmt.Sprintf("Ensure dependency %q is specified", dependency.Name),
							Kind: "yaml",
							Spec: yaml.Spec{
								File:  relativeFoundChartFile,
								Key:   fmt.Sprintf("dependencies[%d].name", i),
								Value: dependency.Name,
							},
						},
					},
				},
				Targets: map[string]target.Config{
					dependency.Name: {
						SourceID: dependency.Name,
						ResourceConfig: resource.ResourceConfig{
							Name: fmt.Sprintf("Bump chart dependency %q in Chart %q", dependency.Name, chartName),
							Kind: "helmchart",
							Spec: helm.Spec{
								File:             metadataFilename,
								Name:             chartRelativeMetadataPath,
								Key:              fmt.Sprintf("dependencies[%d].version", i),
								VersionIncrement: "minor",
							},
						},
					},
				},
			}

			//// Set scmID configuration
			if scmSpec != nil {
				manifest.SCMs = make(map[string]scm.Config)
				manifest.SCMs[scmID] = *scmSpec

				s := manifest.Sources[dependency.Name]
				s.SCMID = scmID
				manifest.Sources[dependency.Name] = s

				c := manifest.Conditions[dependency.Name]
				c.SCMID = scmID
				manifest.Conditions[dependency.Name] = c

				t := manifest.Targets[dependency.Name]
				t.SCMID = scmID
				manifest.Targets[dependency.Name] = t
			}

			if pullrequestSpec != nil {
				manifest.PullRequests = make(map[string]pullrequest.Config)
				manifest.PullRequests[pullrequestID] = *pullrequestSpec
			}

			manifests = append(manifests, manifest)

		}

	}

	return manifests, nil
}

// RunDisabled returns a bool saying if a run could done
func (h Helm) Enabled() bool {
	return !h.spec.Disable
}
