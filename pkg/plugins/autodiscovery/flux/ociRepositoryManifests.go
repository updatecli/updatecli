package flux

import (
	"bytes"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
)

func (f Flux) discoverOCIRepositoryManifests() [][]byte {

	var manifests [][]byte

	for _, foundFluxFile := range f.ociRepositoryFiles {
		logrus.Debugf("parsing oci repository file %q", foundFluxFile)

		relateFoundFluxFile, err := filepath.Rel(f.rootDir, foundFluxFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		data, err := loadOCIRepository(foundFluxFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		if data == nil {
			continue
		}

		ociName := strings.TrimPrefix(data.Spec.URL, "oci://")
		ociVersion := data.Spec.Reference.Tag

		// Skip pipeline if at least of the oci url or oci tag is not specified
		if len(ociName) == 0 || len(ociVersion) == 0 {
			continue
		}

		if len(f.spec.Ignore) > 0 {
			if f.spec.Ignore.isMatchingRules(f.rootDir, relateFoundFluxFile, "", ociName, ociVersion) {
				logrus.Debugf("Ignoring OCI repository %q from %q, as matching ignore rule(s)\n", ociName, relateFoundFluxFile)
				continue
			}
		}

		if len(f.spec.Only) > 0 {
			if !f.spec.Only.isMatchingRules(f.rootDir, relateFoundFluxFile, "", ociName, ociVersion) {
				logrus.Debugf("Ignoring OCI repository %q from %q, as not matching only rule(s)\n", ociName, relateFoundFluxFile)
				continue
			}
		}

		versionFilterKind := defaultVersionFilterKind
		versionFilterPattern := defaultVersionFilterPattern
		versionFilterRegex := defaultVersionFilterRegex
		tagFilter := ""

		sourceSpec := dockerimage.NewDockerImageSpecFromImage(ociName, ociVersion, f.spec.Auths)
		if sourceSpec != nil {
			versionFilterKind = sourceSpec.VersionFilter.Kind
			versionFilterPattern = sourceSpec.VersionFilter.Pattern
			versionFilterRegex = sourceSpec.VersionFilter.Regex
			tagFilter = sourceSpec.TagFilter
		}

		if !f.spec.VersionFilter.IsZero() {
			versionFilterKind = f.versionFilter.Kind
			versionFilterPattern, err = f.versionFilter.GreaterThanPattern(ociVersion)
			versionFilterRegex = f.versionFilter.Regex
			if err != nil {
				logrus.Debugf("building version filter pattern: %s", err)
				versionFilterPattern = ociVersion
			}
		}

		var tmpl *template.Template
		if f.digest && sourceSpec != nil {
			tmpl, err = template.New("manifest").Parse(ociRepositoryManifestTemplateDigestAndLatest)
			if err != nil {
				logrus.Debugf("parsing oci repository file %q: %s", foundFluxFile, err)
				continue
			}
		} else if f.digest && sourceSpec == nil {
			tmpl, err = template.New("manifest").Parse(ociRepositoryManifestTemplateDigest)
			if err != nil {
				logrus.Debugf("parsing oci repository file %q: %s", foundFluxFile, err)
				continue
			}
		} else if !f.digest && sourceSpec != nil {
			tmpl, err = template.New("manifest").Parse(ociRepositoryManifestTemplateLatest)
			if err != nil {
				logrus.Debugf("parsing oci repository file %q: %s", foundFluxFile, err)
				continue
			}
		} else {
			logrus.Infoln("No source spec detected")
			continue
		}

		params := struct {
			ActionID             string
			OCIName              string
			OCIVersion           string
			File                 string
			ImageName            string
			VersionFilterKind    string
			VersionFilterPattern string
			VersionFilterRegex   string
			ScmID                string
			TagFilter            string
		}{
			ActionID:             f.actionID,
			OCIName:              ociName,
			OCIVersion:           ociVersion,
			File:                 relateFoundFluxFile,
			VersionFilterKind:    versionFilterKind,
			VersionFilterPattern: versionFilterPattern,
			VersionFilterRegex:   versionFilterRegex,
			ScmID:                f.scmID,
			TagFilter:            tagFilter,
		}

		manifest := bytes.Buffer{}
		if err := tmpl.Execute(&manifest, params); err != nil {
			logrus.Debugln(err)
			continue
		}

		manifests = append(manifests, manifest.Bytes())
	}

	return manifests
}
