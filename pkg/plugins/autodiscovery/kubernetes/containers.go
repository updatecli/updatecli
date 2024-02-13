package kubernetes

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
)

var (
	// DefaultKubernetesFiles specifies accepted Kubernetes files
	DefaultKubernetesFiles []string = []string{"*.yaml", "*.yml"}
)

func (k Kubernetes) discoverContainerManifests() ([][]byte, error) {

	var manifests [][]byte

	kubernetesFiles, err := searchKubernetesFiles(
		k.rootDir,
		k.files)

	if err != nil {
		return nil, err
	}

	for _, kubernetesFile := range kubernetesFiles {
		logrus.Debugf("parsing file %q", kubernetesFile)

		relativeFoundKubernetesFile, err := filepath.Rel(k.rootDir, kubernetesFile)
		if err != nil {
			// Let's try the next chart if one fail
			logrus.Debugln(err)
			continue
		}

		data, err := getKubernetesManifestData(kubernetesFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		if data == nil {
			continue
		}

		for i, container := range data.Spec.Containers {

			containerName := container.Name
			if containerName == "" {
				containerName = container.Image
			}

			manifest, err := k.generateContainerManifest(
				fmt.Sprintf("$.spec.containers[%d].image", i),
				containerName,
				container.Image,
				relativeFoundKubernetesFile)

			if err != nil {
				logrus.Debugln(err)
				continue
			}

			if manifest == nil {
				continue
			}

			manifests = append(manifests, manifest)
		}

		for i, container := range data.Spec.Template.Spec.Containers {

			containerName := container.Name
			if containerName == "" {
				containerName = container.Image
			}

			manifest, err := k.generateContainerManifest(
				fmt.Sprintf("$.spec.template.spec.containers[%d].image", i),
				containerName,
				container.Image,
				relativeFoundKubernetesFile)

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

func (k Kubernetes) generateContainerManifest(targetKey, containerName, containerImage, relativeFoundKubernetesFile string) ([]byte, error) {
	var err error

	if containerImage == "" {
		return nil, nil
	}

	imageName, imageTag, imageDigest, err := dockerimage.ParseOCIReferenceInfo(containerImage)
	if err != nil {
		return nil, fmt.Errorf("parsing image %q: %s", containerImage, err)
	}

	if imageDigest != "" && imageTag == "" {
		return nil, fmt.Errorf("ignoring image %q because it has a digest but we can't identify the tag", containerImage)
	}

	if len(k.spec.Ignore) > 0 {
		if k.spec.Ignore.isMatchingRules(k.rootDir, relativeFoundKubernetesFile, imageName) {
			logrus.Debugf("Ignoring container %q from %q, as matching ignore rule(s)\n", imageName, relativeFoundKubernetesFile)
			return nil, nil
		}
	}

	if len(k.spec.Only) > 0 {
		if !k.spec.Only.isMatchingRules(k.rootDir, relativeFoundKubernetesFile, imageName) {
			logrus.Debugf("Ignoring container %q from %q, as not matching only rule(s)\n", imageName, relativeFoundKubernetesFile)
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
			versionFilterPattern = "*"
			logrus.Debugf("building version filter pattern: %s", err)
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

	sourceId := containerName

	params := struct {
		ManifestName         string
		ImageName            string
		ImageTag             string
		SourceID             string
		SourceTagFilter      string
		VersionFilterKind    string
		VersionFilterPattern string
		TargetID             string
		TargetKey            string
		TargetPrefix         string
		TargetFile           string
		ScmID                string
	}{
		ManifestName:         fmt.Sprintf("deps: bump container image %q", containerName),
		ImageName:            imageName,
		ImageTag:             imageTag,
		SourceID:             sourceId,
		SourceTagFilter:      tagFilter,
		TargetID:             containerName,
		TargetPrefix:         imageName + ":",
		TargetKey:            targetKey,
		TargetFile:           relativeFoundKubernetesFile,
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
