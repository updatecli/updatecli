package fleet

import (
	"bytes"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"text/template"

	"github.com/sirupsen/logrus"
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

// fleetMetadata is the information that we need to retrieve from Helm chart files.
type fleetMetadata struct {
	Helm fleetHelmData
}

func (f Fleet) discoverFleetDependenciesManifests() ([][]byte, error) {

	var manifests [][]byte

	searchFromDir := f.rootDir
	// If the spec.RootDir is an absolute path, then it as already been set
	// correctly in the New function.
	if f.spec.RootDir != "" && !path.IsAbs(f.spec.RootDir) {
		searchFromDir = filepath.Join(f.rootDir, f.spec.RootDir)
	}

	foundFleetBundleFiles, err := searchFleetBundleFiles(
		searchFromDir,
		FleetBundleFiles[:])

	if err != nil {
		return nil, err
	}

	for _, foundFleetBundleFile := range foundFleetBundleFiles {
		logrus.Debugf("parsing file %q", foundFleetBundleFile)

		relativeFoundChartFile, err := filepath.Rel(f.rootDir, foundFleetBundleFile)
		if err != nil {
			// Let's try the next chart if one fail
			logrus.Debugln(err)
			continue
		}

		chartRelativeMetadataPath := filepath.Dir(relativeFoundChartFile)
		chartName := filepath.Base(chartRelativeMetadataPath)

		// Retrieve chart dependencies for each chart

		data, err := getFleetBundleData(foundFleetBundleFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		if data == nil {
			continue
		}

		// Skip pipeline if at least of the helm chart or helm repository is not specified
		if len(data.Helm.Chart) == 0 || len(data.Helm.Repo) == 0 {
			continue
		}

		if len(f.spec.Ignore) > 0 {
			if f.spec.Ignore.isMatchingRules(f.rootDir, relativeFoundChartFile, data.Helm.Repo, data.Helm.Chart, data.Helm.Version) {
				logrus.Debugf("Ignoring Helm chart %q from %q, as matching ignore rule(s)\n", data.Helm.Chart, relativeFoundChartFile)
				continue
			}
		}

		if len(f.spec.Only) > 0 {
			if !f.spec.Only.isMatchingRules(f.rootDir, relativeFoundChartFile, data.Helm.Repo, data.Helm.Chart, data.Helm.Version) {
				logrus.Debugf("Ignoring Helm chart %q from %q, as not matching only rule(s)\n", data.Helm.Chart, relativeFoundChartFile)
				continue
			}
		}

		token := ""
		foundRegistry := false

		helmRepo, err := url.Parse(data.Helm.Repo) // to validate URL format
		if err != nil {
			logrus.Debugf("invalid URL: %s", err)
			return nil, nil
		}

		if _, foundRegistry = f.spec.Auths[helmRepo.Host]; foundRegistry {
			token = f.spec.Auths[helmRepo.Host].Token
			logrus.Debugf("found token for repository %q", data.Helm.Repo)
		}

		sourceVersionFilterKind := "semver"
		sourceVersionFilterPattern := "*"
		sourceVersionFilterRegex := "*"

		if !f.spec.VersionFilter.IsZero() {
			sourceVersionFilterKind = f.versionFilter.Kind
			sourceVersionFilterPattern, err = f.versionFilter.GreaterThanPattern(data.Helm.Version)
			sourceVersionFilterRegex = f.versionFilter.Regex
			if err != nil {
				logrus.Debugf("building version filter pattern: %s", err)
				sourceVersionFilterPattern = "*"
			}
		}

		tmpl, err := template.New("manifest").Parse(manifestTemplate)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		params := struct {
			ActionID                   string
			ManifestName               string
			ImageName                  string
			ChartName                  string
			ChartRepository            string
			ConditionID                string
			FleetBundle                string
			SourceID                   string
			SourceName                 string
			SourceKind                 string
			SourceVersionFilterKind    string
			SourceVersionFilterPattern string
			SourceVersionFilterRegex   string
			Token                      string
			TargetID                   string
			File                       string
			ScmID                      string
		}{
			ActionID:                   f.actionID,
			ManifestName:               fmt.Sprintf("deps(rancher/fleet): bump %q Fleet bundle for %q Helm chart", chartName, data.Helm.Chart),
			ChartName:                  data.Helm.Chart,
			ChartRepository:            data.Helm.Repo,
			ConditionID:                data.Helm.Chart,
			FleetBundle:                chartName,
			SourceID:                   data.Helm.Chart,
			SourceName:                 fmt.Sprintf("Get latest %q Helm chart version", data.Helm.Chart),
			SourceKind:                 "helmchart",
			SourceVersionFilterKind:    sourceVersionFilterKind,
			SourceVersionFilterPattern: sourceVersionFilterPattern,
			SourceVersionFilterRegex:   sourceVersionFilterRegex,
			Token:                      token,
			TargetID:                   data.Helm.Chart,
			File:                       relativeFoundChartFile,
			ScmID:                      f.scmID,
		}

		manifest := bytes.Buffer{}
		if err := tmpl.Execute(&manifest, params); err != nil {
			logrus.Debugln(err)
			continue
		}

		manifests = append(manifests, manifest.Bytes())
	}

	return manifests, nil
}
