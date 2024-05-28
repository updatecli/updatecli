package argocd

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/sirupsen/logrus"
)

var (
	// ArgoCDFilePatterns specifies accepted Helm chart metadata file name
	ArgoCDFilePatterns [2]string = [2]string{"*.yaml", "*.yml"}
)

// ArgoCDApplicationSpec is the information that we need to retrieve from Helm chart files.
type ArgoCDApplicationSpec struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Spec       struct {
		Source  ApplicationSourceSpec
		Sources []ApplicationSourceSpec
	}
}

type ApplicationSourceSpec struct {
	RepoURL        string `yaml:"repoURL"`
	TargetRevision string `yaml:"targetRevision"`
	Chart          string `yaml:"chart"`
	Ref            string `yaml:"ref"`
}

// IsZero checks if the ApplicationSourceSpec is empty in the context of Updatecli
// We prefer ignoring when chart and targetRevision are empty
func (a ApplicationSourceSpec) IsZero() bool {
	return a.RepoURL == "" || a.TargetRevision == "" || a.Chart == ""
}

func (f ArgoCD) discoverArgoCDManifests() ([][]byte, error) {

	var manifests [][]byte

	foundFiles, err := searchArgoCDFiles(
		f.rootDir,
		ArgoCDFilePatterns[:])

	if err != nil {
		return nil, err
	}

	for _, foundFile := range foundFiles {
		logrus.Debugf("parsing file %q", foundFile)

		relativeFilepath, err := filepath.Rel(f.rootDir, foundFile)
		if err != nil {
			// Let's try the next chart if one fail
			logrus.Debugln(err)
			continue
		}

		// Retrieve chart dependencies for each chart
		data, err := readManifest(foundFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		if !data.Spec.Source.IsZero() {
			manifest, err := f.generateManifestBySource(
				data.Spec.Source,
				relativeFilepath,
				"$.spec.source",
			)
			if err != nil {
				logrus.Errorf("error discovering application source: %s", err)
			}

			manifests = append(manifests, manifest)
		}

		for i, source := range data.Spec.Sources {
			if source.IsZero() {
				continue
			}

			manifest, err := f.generateManifestBySource(
				source,
				relativeFilepath,
				fmt.Sprintf("$.spec.sources[%d]", i),
			)
			if err != nil {
				logrus.Errorf("error discovering application source: %s", err)
			}

			manifests = append(manifests, manifest)
		}
	}

	return manifests, nil
}

// generateManifestBySource generates an ArgoCD manifest for Updatecli based on the application source input
func (f ArgoCD) generateManifestBySource(data ApplicationSourceSpec, file string, targetKey string) ([]byte, error) {

	var err error

	if data.IsZero() {
		return nil, nil
	}

	if len(f.spec.Ignore) > 0 {
		if f.spec.Ignore.isMatchingRules(f.rootDir, file, data.RepoURL, data.Chart, data.TargetRevision) {
			logrus.Debugf("Ignoring Helm chart %q from %q, as matching ignore rule(s)\n", data.Chart, file)
			return nil, nil
		}
	}

	if len(f.spec.Only) > 0 {
		if !f.spec.Only.isMatchingRules(f.rootDir, file, data.RepoURL, data.Chart, data.TargetRevision) {
			logrus.Debugf("Ignoring Helm chart %q from %q, as not matching only rule(s)\n", data.Chart, file)
			return nil, nil
		}
	}

	sourceVersionFilterKind := "semver"
	sourceVersionFilterPattern := "*"

	if !f.spec.VersionFilter.IsZero() {
		sourceVersionFilterKind = f.versionFilter.Kind
		sourceVersionFilterPattern, err = f.versionFilter.GreaterThanPattern(data.TargetRevision)
		if err != nil {
			logrus.Debugf("building version filter pattern: %s", err)
			sourceVersionFilterPattern = "*"
		}
	}

	tmpl, err := template.New("manifest").Parse(manifestTemplate)
	if err != nil {
		logrus.Debugln(err)
		return nil, nil
	}

	params := struct {
		ManifestName               string
		ImageName                  string
		ChartName                  string
		ChartRepository            string
		ManifestFile               string
		ConditionID                string
		SourceID                   string
		SourceName                 string
		SourceKind                 string
		SourceVersionFilterKind    string
		SourceVersionFilterPattern string
		TargetID                   string
		TargetKey                  string
		File                       string
		ScmID                      string
	}{
		ManifestName:               fmt.Sprintf("deps(helm): bump Helm chart %q in ArgoCD manifest %q", data.Chart, file),
		ChartName:                  data.Chart,
		ChartRepository:            data.RepoURL,
		ConditionID:                data.Chart,
		ManifestFile:               file,
		SourceID:                   data.Chart,
		SourceName:                 fmt.Sprintf("Get latest %q Helm chart version", data.Chart),
		SourceKind:                 "helmchart",
		SourceVersionFilterKind:    sourceVersionFilterKind,
		SourceVersionFilterPattern: sourceVersionFilterPattern,
		TargetID:                   data.Chart,
		TargetKey:                  targetKey,
		File:                       file,
		ScmID:                      f.scmID,
	}

	manifest := bytes.Buffer{}
	if err = tmpl.Execute(&manifest, params); err != nil {
		return nil, fmt.Errorf("error executing template: %w", err)
	}

	return manifest.Bytes(), nil
}
