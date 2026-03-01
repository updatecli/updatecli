package kubernetes

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
	// DefaultKubernetesFiles specifies accepted Kubernetes files
	DefaultKubernetesFiles []string = []string{"*.yaml", "*.yml"}
)

func (k Kubernetes) discoverContainerManifests() ([][]byte, error) {
	var manifests [][]byte
	searchFromDir := k.rootDir
	// If the spec.RootDir is an absolute path, then it as already been set
	// correctly in the New function.
	if k.spec.RootDir != "" && !path.IsAbs(k.spec.RootDir) {
		searchFromDir = filepath.Join(k.rootDir, k.spec.RootDir)
	}

	kubernetesFiles, err := searchKubernetesFiles(
		searchFromDir,
		k.files)

	if err != nil {
		return nil, err
	}

	for _, kubernetesFile := range kubernetesFiles {
		logrus.Debugf("parsing file %q", kubernetesFile)

		relativeFoundKubernetesFile, err := filepath.Rel(k.rootDir, kubernetesFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		switch k.flavor {
		case FlavorKubernetes:
			manifests = append(manifests, k.discoverKubernetesManifest(kubernetesFile, relativeFoundKubernetesFile)...)
		case FlavorProw:
			manifests = append(manifests, k.discoverProwManifest(kubernetesFile, relativeFoundKubernetesFile)...)
		default:
			return nil, fmt.Errorf("Kubernetes manifest %q not supported", k.flavor)
		}
	}

	return manifests, nil
}

func (k Kubernetes) generateContainerManifest(targetKey, containerName, containerImage, relativeFoundKubernetesFile, manifestNameSuffix string, yamlDocument int) ([]byte, error) {
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
	versionFilterRegex := k.versionFilter.Regex
	tagFilter := "*"

	registryUsername := ""
	registryPassword := ""
	registryToken := ""

	if sourceSpec != nil {
		versionFilterKind = sourceSpec.VersionFilter.Kind
		versionFilterPattern = sourceSpec.VersionFilter.Pattern
		versionFilterRegex = sourceSpec.VersionFilter.Regex
		tagFilter = sourceSpec.TagFilter

		registryUsername = sourceSpec.Username
		registryPassword = sourceSpec.Password
		registryToken = sourceSpec.Token
	}

	// If a versionfilter is specified in the manifest then we want to be sure that it takes precedence
	if !k.spec.VersionFilter.IsZero() {
		versionFilterKind = k.versionFilter.Kind
		versionFilterPattern, err = k.versionFilter.GreaterThanPattern(imageTag)
		versionFilterRegex = k.versionFilter.Regex
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
		ActionID             string
		ManifestName         string
		ImageName            string
		ImageTag             string
		SourceID             string
		SourceTagFilter      string
		VersionFilterKind    string
		VersionFilterPattern string
		VersionFilterRegex   string
		RegistryUsername     string
		RegistryPassword     string
		RegistryToken        string
		TargetID             string
		TargetKey            string
		TargetPrefix         string
		TargetFile           string
		TargetYamlDocument   int
		ScmID                string
	}{
		ActionID:             k.actionID,
		ManifestName:         fmt.Sprintf("deps: bump container image %q%s", containerName, manifestNameSuffix),
		ImageName:            imageName,
		ImageTag:             imageTag,
		SourceID:             sourceId,
		SourceTagFilter:      tagFilter,
		RegistryUsername:     registryUsername,
		RegistryPassword:     registryPassword,
		RegistryToken:        registryToken,
		TargetID:             containerName,
		TargetPrefix:         imageName + ":",
		TargetKey:            targetKey,
		TargetFile:           relativeFoundKubernetesFile,
		TargetYamlDocument:   yamlDocument,
		VersionFilterKind:    versionFilterKind,
		VersionFilterPattern: versionFilterPattern,
		VersionFilterRegex:   versionFilterRegex,
		ScmID:                k.scmID,
	}

	manifest := bytes.Buffer{}
	if err := tmpl.Execute(&manifest, params); err != nil {
		return nil, err
	}

	return manifest.Bytes(), nil

}
