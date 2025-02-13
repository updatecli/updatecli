package ko

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
)

var (
	// DefaultKoFiles specifies accepted Kubernetes files
	DefaultKoFiles []string = []string{".ko.yaml"}
)

func (k Ko) discoverContainerManifests() ([][]byte, error) {

	var manifests [][]byte

	searchFromDir := k.rootDir
	// If the spec.RootDir is an absolute path, then it as already been set
	// correctly in the New function.
	if k.spec.RootDir != "" && !path.IsAbs(k.spec.RootDir) {
		searchFromDir = filepath.Join(k.rootDir, k.spec.RootDir)
	}

	koFiles, err := searchKosFiles(
		searchFromDir,
		DefaultKoFiles)
	if err != nil {
		return nil, err
	}

	for _, koFile := range koFiles {
		logrus.Debugf("parsing file %q", koFile)

		relativeFoundKoFile, err := filepath.Rel(k.rootDir, koFile)
		if err != nil {
			// Let's try the next chart if one fail
			logrus.Debugln(err)
			continue
		}

		data, err := getKoManifestData(koFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		if data == nil {
			continue
		}

		for i, image := range data.BaseImageOverrides {

			manifest, err := k.generateContainerManifest(
				fmt.Sprintf(`$.baseImageOverrides.'%s'`, i),
				image,
				relativeFoundKoFile)
			if err != nil {
				logrus.Debugln(err)
				continue
			}

			if manifest == nil {
				continue
			}

			manifests = append(manifests, manifest)
		}

		if data.DefaultBaseImage != "" {

			manifest, err := k.generateContainerManifest("$.defaultBaseImage", data.DefaultBaseImage, relativeFoundKoFile)
			if err != nil {
				logrus.Debugln(err)
				continue
			}

			if manifest == nil {
				continue
			}

			manifests = append(manifests, manifest)
		}

	}

	return manifests, nil
}

func (k Ko) generateContainerManifest(targetKey, image, relativeFoundKoFile string) ([]byte, error) {
	var err error

	if image == "" {
		return nil, nil
	}

	imageName, imageTag, imageDigest, err := dockerimage.ParseOCIReferenceInfo(image)
	if err != nil {
		return nil, fmt.Errorf("parsing image %q: %s", image, err)
	}

	if imageDigest != "" && imageTag == "" {
		return nil, fmt.Errorf("ignoring image %q because it has a digest but we can't identify the tag", image)
	}

	if len(k.spec.Ignore) > 0 {
		if k.spec.Ignore.isMatchingRules(k.rootDir, relativeFoundKoFile, imageName) {
			logrus.Debugf("Ignoring container %q from %q, as matching ignore rule(s)\n", imageName, relativeFoundKoFile)
			return nil, nil
		}
	}

	if len(k.spec.Only) > 0 {
		if !k.spec.Only.isMatchingRules(k.rootDir, relativeFoundKoFile, imageName) {
			logrus.Debugf("Ignoring container %q from %q, as not matching only rule(s)\n", imageName, relativeFoundKoFile)
			return nil, nil
		}
	}

	sourceSpec := dockerimage.NewDockerImageSpecFromImage(imageName, imageTag, k.spec.Auths)
	if sourceSpec == nil && !k.digest {
		logrus.Infoln("No source spec detected")
		return nil, nil
	}

	versionFilterKind := k.versionFilter.Kind
	versionFilterPattern := k.versionFilter.Pattern
	tagFilter := "*"

	if sourceSpec != nil {
		versionFilterKind = sourceSpec.VersionFilter.Kind
		versionFilterPattern = sourceSpec.VersionFilter.Pattern
		tagFilter = sourceSpec.TagFilter
	}

	// If a versionfilter is specified in the manifest then we want to be sure that it takes precedence
	if !k.spec.VersionFilter.IsZero() {
		versionFilterKind = k.versionFilter.Kind
		versionFilterPattern, err = k.versionFilter.GreaterThanPattern(imageTag)
		if err != nil {
			logrus.Debugf("building version filter pattern: %s", err)
			versionFilterPattern = "*"
		}
	}

	var tmpl *template.Template
	if k.digest && sourceSpec != nil {
		tmpl, err = template.New("manifest").Parse(manifestTemplateDigestAndLatest)
		if err != nil {
			return nil, err
		}
	} else if k.digest && sourceSpec == nil {
		tmpl, err = template.New("manifest").Parse(manifestTemplateDigest)
		if err != nil {
			return nil, err
		}
	} else if !k.digest && sourceSpec != nil {
		tmpl, err = template.New("manifest").Parse(manifestTemplateLatest)
		if err != nil {
			return nil, err
		}
	} else {
		logrus.Infoln("No source spec detected")
		return nil, nil
	}

	sourceId := imageName

	params := struct {
		ActionID             string
		ManifestName         string
		ImageName            string
		ImageTag             string
		SourceID             string
		SourceName           string
		SourceKind           string
		SourceTagFilter      string
		VersionFilterKind    string
		VersionFilterPattern string
		TargetID             string
		TargetKey            string
		TargetPrefix         string
		TargetFile           string
		TargetName           string
		ScmID                string
	}{
		ActionID:             k.actionID,
		ManifestName:         fmt.Sprintf("deps: bump container image %q", imageName),
		ImageName:            imageName,
		ImageTag:             imageTag,
		SourceID:             sourceId,
		SourceName:           fmt.Sprintf("get latest container image tag for %q", imageName),
		SourceKind:           "dockerimage",
		SourceTagFilter:      tagFilter,
		TargetID:             imageName,
		TargetPrefix:         imageName + ":",
		TargetKey:            targetKey,
		TargetFile:           relativeFoundKoFile,
		TargetName:           fmt.Sprintf("deps: bump container image %q to {{ source %q }}", imageName, sourceId),
		VersionFilterKind:    versionFilterKind,
		VersionFilterPattern: versionFilterPattern,
		ScmID:                k.scmID,
	}

	manifest := bytes.Buffer{}
	if err := tmpl.Execute(&manifest, params); err != nil {
		return nil, err
	}

	return manifest.Bytes(), nil
}
