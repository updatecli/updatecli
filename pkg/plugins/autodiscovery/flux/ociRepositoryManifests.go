package flux

import (
	"bytes"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
)

func (f Flux) discoverOCIRepositoryManifests() [][]byte {

	var manifests [][]byte

	for _, foundFluxFile := range f.ociRepositoryFiles {
		relativeFoundFluxFile, err := filepath.Rel(f.rootDir, foundFluxFile)
		if err != nil {
			log.Debugln(err)
			continue
		}

		dataDocuments, err := loadOCIRepository(foundFluxFile)
		if err != nil {
			log.Debugln(err)
			continue
		}

		for index, d := range dataDocuments {
			ociName := strings.TrimPrefix(d.Spec.URL, "oci://")
			ociVersion := d.Spec.Reference.Tag

			// Skip pipeline if at least of the oci url or oci tag is not specified
			if len(ociName) == 0 || len(ociVersion) == 0 {
				continue
			}

			if len(f.spec.Ignore) > 0 {
				if f.spec.Ignore.isMatchingRules(f.rootDir, relativeFoundFluxFile, "", ociName, ociVersion) {
					log.Debugf("Ignoring OCI repository %q from %q, as matching ignore rule(s)\n", ociName, relativeFoundFluxFile)
					continue
				}
			}

			if len(f.spec.Only) > 0 {
				if !f.spec.Only.isMatchingRules(f.rootDir, relativeFoundFluxFile, "", ociName, ociVersion) {
					log.Debugf("Ignoring OCI repository %q from %q, as not matching only rule(s)\n", ociName, relativeFoundFluxFile)
					continue
				}
			}

			versionFilterKind := defaultVersionFilterKind
			versionFilterPattern := defaultVersionFilterPattern
			versionFilterRegex := defaultVersionFilterRegex
			tagFilter := ""

			registryUsername := ""
			registryPassword := ""
			registryToken := ""

			// New versions are looked up under the upstream name, while the targets keep
			// the URL written in the file.
			sourceOCIName := f.spec.RegistryMappings.Resolve(ociName)

			sourceSpec := dockerimage.NewDockerImageSpecFromImage(sourceOCIName, ociVersion, f.spec.Auths)
			if sourceSpec != nil {
				versionFilterKind = sourceSpec.VersionFilter.Kind
				versionFilterPattern = sourceSpec.VersionFilter.Pattern
				versionFilterRegex = sourceSpec.VersionFilter.Regex
				tagFilter = sourceSpec.TagFilter

				registryPassword = sourceSpec.Password
				registryUsername = sourceSpec.Username
				registryToken = sourceSpec.Token
			}

			if !f.spec.VersionFilter.IsZero() {
				versionFilterKind = f.versionFilter.Kind
				versionFilterPattern, err = f.versionFilter.GreaterThanPattern(ociVersion)
				versionFilterRegex = f.versionFilter.Regex
				if err != nil {
					log.Debugf("building version filter pattern: %s", err)
					versionFilterPattern = ociVersion
				}
			}

			// A mirror may not serve a version found upstream yet, or may serve it under
			// another digest, so the targets only run once the mirror has what they write.
			// With digest enabled, the targets write the tag and the digest to separate keys,
			// so the condition checks "<tag>@<digest>" built from both.
			mirrorConditionSourceID := "oci"
			mirrorConditionPrefix := ""
			if f.digest {
				mirrorConditionSourceID = "oci-digest"
				mirrorConditionPrefix = ociVersion + "@"
				if sourceSpec != nil {
					mirrorConditionPrefix = `{{ source "oci" }}@`
				}
			}
			mirrorCredential, _ := dockerimage.RegistryCredentials(ociName, f.spec.Auths)

			var tmpl *template.Template
			if f.digest && sourceSpec != nil {
				tmpl, err = template.New("manifest").Parse(ociRepositoryManifestTemplateDigestAndLatest)
				if err != nil {
					log.Debugf("parsing oci repository file %q: %s", foundFluxFile, err)
					continue
				}
			} else if f.digest && sourceSpec == nil {
				tmpl, err = template.New("manifest").Parse(ociRepositoryManifestTemplateDigest)
				if err != nil {
					log.Debugf("parsing oci repository file %q: %s", foundFluxFile, err)
					continue
				}
			} else if !f.digest && sourceSpec != nil {
				tmpl, err = template.New("manifest").Parse(ociRepositoryManifestTemplateLatest)
				if err != nil {
					log.Debugf("parsing oci repository file %q: %s", foundFluxFile, err)
					continue
				}
			} else {
				log.Infoln("No source spec detected")
				continue
			}

			params := struct {
				ActionID                string
				TargetYAMLDocument      int
				OCIName                 string
				SourceOCIName           string
				OCIVersion              string
				File                    string
				ImageName               string
				VersionFilterKind       string
				VersionFilterPattern    string
				VersionFilterRegex      string
				RegistryUsername        string
				RegistryPassword        string
				RegistryToken           string
				ScmID                   string
				TagFilter               string
				MirrorCondition         bool
				MirrorConditionSourceID string
				MirrorConditionPrefix   string
				MirrorUsername          string
				MirrorPassword          string
				MirrorToken             string
			}{
				ActionID:                f.actionID,
				OCIName:                 ociName,
				SourceOCIName:           sourceOCIName,
				OCIVersion:              ociVersion,
				File:                    relativeFoundFluxFile,
				VersionFilterKind:       versionFilterKind,
				VersionFilterPattern:    versionFilterPattern,
				VersionFilterRegex:      versionFilterRegex,
				ScmID:                   f.scmID,
				RegistryUsername:        registryUsername,
				RegistryPassword:        registryPassword,
				RegistryToken:           registryToken,
				TagFilter:               tagFilter,
				TargetYAMLDocument:      index,
				MirrorCondition:         sourceOCIName != ociName,
				MirrorConditionSourceID: mirrorConditionSourceID,
				MirrorConditionPrefix:   mirrorConditionPrefix,
				MirrorUsername:          mirrorCredential.Username,
				MirrorPassword:          mirrorCredential.Password,
				MirrorToken:             mirrorCredential.Token,
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
