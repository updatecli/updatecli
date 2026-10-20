package flux

import (
	"bytes"
	"net/url"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/semver/v3"
	"github.com/updatecli/updatecli/pkg/core/log"
)

func (f Flux) discoverHelmreleaseManifests() [][]byte {

	var manifests [][]byte

	for _, foundFluxFile := range f.helmReleaseFiles {

		relativeFoundFluxFile, err := filepath.Rel(f.rootDir, foundFluxFile)
		if err != nil {
			log.Debugln(err)
			continue
		}

		// Retrieve chart dependencies for each chart
		dataDocuments, err := loadHelmRelease(foundFluxFile)
		if err != nil {
			log.Debugln(err)
			continue
		}

		for index, d := range dataDocuments {
			if d.Spec.Chart == nil {
				log.Debugf("Ignoring Helmrelease from %q, as it does not define a chart template", relativeFoundFluxFile)
				continue
			}
			helmChartName := d.Spec.Chart.Spec.Chart
			helmChartVersion := d.Spec.Chart.Spec.Version

			sourceRef := d.Spec.Chart.Spec.SourceRef
			if sourceRef.Namespace == "" {
				sourceRef.Namespace = d.GetNamespace()
			}

			helmRepositoryURL := ""
			for _, helmRepository := range f.helmRepositories {
				if helmRepository.GetName() == sourceRef.Name && helmRepository.GetNamespace() == sourceRef.Namespace {
					helmRepositoryURL = helmRepository.Spec.URL
					break
				}
			}

			// Skip pipeline if at least of the helm chart or helm repository is not specified
			if len(helmChartName) == 0 || len(helmChartVersion) == 0 || len(helmRepositoryURL) == 0 {
				continue
			}

			// If the helmrelease version is not a valid semver, we skip the pipeline
			_, err = semver.NewVersion(helmChartVersion)
			if err != nil {
				if semver.ErrInvalidSemVer == err {
					log.Debugf("Ignoring Helm chart %q from %q, as %q not a valid semver version\n", helmChartName, relativeFoundFluxFile, helmChartVersion)
					continue
				}
				log.Debugf("parsing Helm chart version %q: %s", helmChartVersion, err)
			}

			if len(f.spec.Ignore) > 0 {
				if f.spec.Ignore.isMatchingRules(f.rootDir, relativeFoundFluxFile, helmRepositoryURL, helmChartName, helmChartVersion) {
					log.Debugf("Ignoring Helm chart %q from %q, as matching ignore rule(s)\n", helmChartName, relativeFoundFluxFile)
					continue
				}
			}

			if len(f.spec.Only) > 0 {
				if !f.spec.Only.isMatchingRules(f.rootDir, relativeFoundFluxFile, helmRepositoryURL, helmChartName, helmChartVersion) {
					log.Debugf("Ignoring Helm chart %q from %q, as not matching only rule(s)\n", helmChartName, relativeFoundFluxFile)
					continue
				}
			}

			token := ""
			repoURL, err := url.Parse(helmRepositoryURL)
			switch err {
			case nil:
				if _, ok := f.spec.Auths[repoURL.Host]; ok {
					token = f.spec.Auths[repoURL.Host].Token
					log.Debugf("found token for repository %q", repoURL.Host)
				}
			default:
				log.Debugf("Ignoring auth configuration due to invalid Helm repository URL: %s", err)
			}

			sourceVersionFilterKind := defaultVersionFilterKind
			sourceVersionFilterPattern := defaultVersionFilterPattern
			sourceVersionFilterRegex := defaultVersionFilterRegex

			if !f.spec.VersionFilter.IsZero() {
				sourceVersionFilterKind = f.versionFilter.Kind
				sourceVersionFilterPattern, err = f.versionFilter.GreaterThanPattern(helmChartVersion)
				sourceVersionFilterRegex = f.versionFilter.Regex
				if err != nil {
					log.Debugf("building version filter pattern: %s", err)
					sourceVersionFilterPattern = helmChartVersion
				}
			}

			tmpl, err := template.New("manifest").Parse(helmreleaseManifestTemplate)
			if err != nil {
				log.Debugln(err)
				continue
			}

			params := struct {
				ActionID                   string
				ChartName                  string
				ChartRepository            string
				File                       string
				ImageName                  string
				SourceVersionFilterKind    string
				SourceVersionFilterPattern string
				SourceVersionFilterRegex   string
				ScmID                      string
				TargetYAMLDocument         int
				Token                      string
			}{
				ActionID:                   f.actionID,
				ChartName:                  helmChartName,
				ChartRepository:            helmRepositoryURL,
				File:                       relativeFoundFluxFile,
				SourceVersionFilterKind:    sourceVersionFilterKind,
				SourceVersionFilterPattern: sourceVersionFilterPattern,
				SourceVersionFilterRegex:   sourceVersionFilterRegex,
				ScmID:                      f.scmID,
				TargetYAMLDocument:         index,
				Token:                      token,
			}

			manifest := bytes.Buffer{}
			if err := tmpl.Execute(&manifest, params); err != nil {
				log.Debugln(err)
				continue
			}

			manifests = append(manifests, manifest.Bytes())
		}
	}

	return manifests
}
