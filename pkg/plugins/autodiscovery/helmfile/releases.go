package helmfile

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/helm"
)

var (
	// DefaultFilePattern specifies accepted Helm chart metadata filename
	DefaultFilePattern [2]string = [2]string{"*.yaml", "*.yml"}
)

// Release holds the Helmfile release information.
type release struct {
	Name    string
	Chart   string
	Version string
}

// Repository holds the Helmfile repository information
type repository struct {
	Name     string
	URL      string
	OCI      bool
	Username string
	Password string
}

// helmfileMetadata is the information retrieved from Helmfile files.
type helmfileMetadata struct {
	Name         string
	Repositories []repository
	Releases     []release
}

// discoverHelmfileReleaseManifests search recursively from a root directory for Helmfile file
func (h Helmfile) discoverHelmfileReleaseManifests() ([][]byte, error) {

	var manifests [][]byte

	foundHelmfileFiles, err := searchHelmfileFiles(
		h.rootDir,
		DefaultFilePattern[:])

	if err != nil {
		return nil, err
	}

	for _, foundHelmfile := range foundHelmfileFiles {

		relativeFoundChartFile, err := filepath.Rel(h.rootDir, foundHelmfile)
		if err != nil {
			// Jump to the next Helmfile if current failed
			logrus.Errorln(err)
			continue
		}

		helmfileRelativeMetadataPath := filepath.Dir(relativeFoundChartFile)
		helmfileFilename := filepath.Base(helmfileRelativeMetadataPath)

		// Test if the ignore rule based on path doesn't match
		if len(h.spec.Ignore) > 0 && h.spec.Ignore.isMatchingIgnoreRule(h.rootDir, relativeFoundChartFile) {
			logrus.Debugf("Ignoring Helmfile %q from %q, as not matching rule(s)\n",
				helmfileFilename,
				helmfileRelativeMetadataPath)
			continue
		}

		// Test if the only rule based on path doesn't match
		if len(h.spec.Only) > 0 && !h.spec.Only.isMatchingOnlyRule(h.rootDir, relativeFoundChartFile) {
			logrus.Debugf("Ignoring Helmfile %q from %q, as not matching rule(s)\n",
				helmfileFilename,
				helmfileRelativeMetadataPath)
			continue
		}

		// Retrieve chart dependencies for each chart

		metadata, err := getHelmfileMetadata(foundHelmfile)
		if err != nil {
			return nil, err
		}

		if metadata == nil {
			continue
		}

		if len(metadata.Releases) == 0 {
			continue
		}

		for i, release := range metadata.Releases {
			var chartName, chartURL, OCIUsername, OCIPassword string
			var isOCI bool

			for _, repository := range metadata.Repositories {
				if strings.HasPrefix(release.Chart, repository.Name+"/") {
					chartName = strings.TrimPrefix(release.Chart, repository.Name+"/")
					chartURL = repository.URL
					isOCI = repository.OCI
					OCIUsername = repository.Username
					OCIPassword = repository.Password
					break
				}
			}

			if chartName == "" || chartURL == "" {
				logrus.Debugf("repository not identified for release %q, skipping", release.Chart)
				continue
			}

			// Helmfile uses the repository flag 'oci'
			// to identify OCI Helm chart
			// Updatecli expects the scheme 'oci://'.
			// Therefor Updatecli removes any 'http://' or 'https://' schemes before adding 'oci://'
			if isOCI {
				for _, scheme := range []string{"https://", "http://"} {
					if strings.HasPrefix(chartURL, scheme) {
						chartURL = strings.TrimPrefix(chartURL, scheme)
						break
					}
				}
				chartURL = "oci://" + chartURL
			}

			if release.Version == "" {
				logrus.Debugf("no version specified for release %q, skipping", release.Chart)
				continue
			}

			helmSourcespec := helm.Spec{
				Name: chartName,
				URL:  chartURL,
			}
			if OCIUsername != "" && isOCI {
				helmSourcespec.InlineKeyChain.Username = OCIUsername
			}
			if OCIPassword != "" && isOCI {
				helmSourcespec.InlineKeyChain.Password = OCIPassword
			}

			//manifest := config.Spec{
			//	Name: manifestName,
			//	Sources: map[string]source.Config{
			//		sourceID: {
			//			ResourceConfig: resource.ResourceConfig{
			//				Name: fmt.Sprintf("Get latest %q Helm Chart Version", release.Name),
			//				Kind: "helmchart",
			//				Spec: helmSourcespec,
			//			},
			//		},
			//	},
			//	Conditions: map[string]condition.Config{
			//		conditionID: {
			//			DisableSourceInput: true,
			//			ResourceConfig: resource.ResourceConfig{
			//				Name: fmt.Sprintf("Ensure release %q is specified for Helmfile %q", release.Name, relativeFoundChartFile),
			//				Kind: "yaml",
			//				Spec: yaml.Spec{
			//					File:  foundHelmfile,
			//					Key:   fmt.Sprintf("releases[%d].chart", i),
			//					Value: release.Chart,
			//				},
			//			},
			//		},
			//	},
			//	Targets: map[string]target.Config{
			//		targetID: {
			//			SourceID: release.Name,
			//			ResourceConfig: resource.ResourceConfig{
			//				Name: fmt.Sprintf("Bump %q Helm Chart Version for Helmfile %q", release.Name, relativeFoundChartFile),
			//				Kind: "yaml",
			//				Spec: yaml.Spec{
			//					File: foundHelmfile,
			//					Key:  fmt.Sprintf("releases[%d].version", i),
			//				},
			//			},
			//		},
			//	},
			//}
			// Set scmID if defined
			tmpl, err := template.New("manifest").Parse(manifestTemplate)
			if err != nil {
				logrus.Errorln(err)
				continue
			}

			params := struct {
				ManifestName               string
				ChartName                  string
				ChartRepository            string
				ConditionID                string
				ConditionName              string
				ConditionKey               string
				ConditionValue             string
				SourceID                   string
				SourceName                 string
				SourceKind                 string
				SourceVersionFilterKind    string
				SourceVersionFilterPattern string
				TargetID                   string
				TargetName                 string
				TargetKey                  string
				File                       string
				ScmID                      string
			}{
				ManifestName:               fmt.Sprintf("Bump %q Helm Chart version for Helmfile %q", release.Name, relativeFoundChartFile),
				ChartName:                  chartName,
				ChartRepository:            chartURL,
				ConditionID:                release.Name,
				ConditionName:              fmt.Sprintf("Ensure release %q is specified for Helmfile %q", release.Name, relativeFoundChartFile),
				ConditionKey:               fmt.Sprintf("releases[%d].chart", i),
				ConditionValue:             release.Chart,
				SourceID:                   release.Name,
				SourceName:                 fmt.Sprintf("Get latest %q Helm Chart Version", release.Name),
				SourceKind:                 "helmchart",
				SourceVersionFilterKind:    "semver",
				SourceVersionFilterPattern: "'*'",
				TargetID:                   release.Name,
				TargetName:                 fmt.Sprintf("Bump %q Helm Chart Version for Helmfile %q", release.Name, relativeFoundChartFile),
				TargetKey:                  fmt.Sprintf("releases[%d].version", i),
				File:                       foundHelmfile,
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
