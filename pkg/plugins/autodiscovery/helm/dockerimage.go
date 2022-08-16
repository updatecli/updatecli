package helm

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
	"github.com/updatecli/updatecli/pkg/plugins/resources/helm"
	"github.com/updatecli/updatecli/pkg/plugins/resources/yaml"

	dockerimageutils "github.com/updatecli/updatecli/pkg/plugins/utils/docker/dockerimage"
)

type imageRef struct {
	Repository string
	Tag        string
}

type valuesContent struct {
	Image  imageRef
	Images map[string]imageRef
}

func (h Helm) discoverHelmContainerManifests() ([]config.Spec, error) {

	var manifests []config.Spec

	foundValuesFiles, err := searchChartFiles(
		h.rootDir,
		[]string{"values.yaml", "values.yml"})

	if err != nil {
		return nil, err
	}

	for _, foundValueFile := range foundValuesFiles {

		relativeFoundValueFile, err := filepath.Rel(h.rootDir, foundValueFile)
		if err != nil {
			// Let's try the next chart if one fail
			logrus.Errorln(err)
			continue
		}

		chartRelativeMetadataPath := filepath.Dir(relativeFoundValueFile)
		metadataFilename := filepath.Base(foundValueFile)
		chartName := filepath.Base(chartRelativeMetadataPath)

		// Test if the ignore rule based on path is respected
		if len(h.spec.Ignore) > 0 && h.spec.Ignore.isMatchingIgnoreRule(h.rootDir, relativeFoundValueFile) {
			logrus.Debugf("Ignoring Helm Chart %q from %q, as not matching rule(s)\n",
				chartName,
				chartRelativeMetadataPath)
			continue
		}

		// Test if the only rule based on path is respected
		if len(h.spec.Only) > 0 && !h.spec.Only.isMatchingOnlyRule(h.rootDir, relativeFoundValueFile) {
			logrus.Debugf("Ignoring Helm Chart %q from %q, as not matching rule(s)\n",
				chartName,
				chartRelativeMetadataPath)
			continue
		}

		// Retrieve chart dependencies for each chart

		values, err := getValuesFileContent(foundValueFile)
		if err != nil {
			return nil, err
		}

		if values == nil {
			continue
		}

		type imageData struct {
			repository         string
			tag                string
			yamlRepositoryPath string
			yamlTagPath        string
		}

		var images []imageData

		if values.Image.Repository != "" && values.Image.Tag != "" {
			images = append(images, imageData{
				repository:         values.Image.Repository,
				tag:                values.Image.Tag,
				yamlRepositoryPath: "image.repository",
				yamlTagPath:        "image.tag",
			})

		}

		for id := range values.Images {
			images = append(images, imageData{
				repository:         values.Images[id].Repository,
				tag:                values.Images[id].Tag,
				yamlRepositoryPath: fmt.Sprintf("images.%s.repository", id),
				yamlTagPath:        fmt.Sprintf("images.%s.tag", id),
			})

		}

		for _, image := range images {

			sourceID := image.repository
			conditionID := image.repository
			targetID := image.repository

			yamlRepositoryPath := image.yamlRepositoryPath
			yamlTagPath := image.yamlTagPath

			dockerImageSpec := h.generateSourceDockerImageSpec(image.repository)

			manifest := config.Spec{
				Name: strings.Join([]string{
					chartName,
					image.repository,
				}, "_"),
				Sources: map[string]source.Config{
					sourceID: {
						ResourceConfig: resource.ResourceConfig{
							Name: fmt.Sprintf("Get latest %q Container tag", image.repository),
							Kind: "dockerimage",
							Spec: dockerImageSpec,
						},
					},
				},
				Conditions: map[string]condition.Config{
					conditionID: {
						DisableSourceInput: true,
						ResourceConfig: resource.ResourceConfig{
							Name: fmt.Sprintf("Ensure container repository %q is specified", image.repository),
							Kind: "yaml",
							Spec: yaml.Spec{
								File:  relativeFoundValueFile,
								Key:   yamlRepositoryPath,
								Value: image.repository,
							},
						},
					},
				},
				Targets: map[string]target.Config{
					targetID: {
						SourceID: sourceID,
						ResourceConfig: resource.ResourceConfig{
							Name: fmt.Sprintf("Bump container image tag for image %q in Chart %q", image.repository, chartName),
							Kind: "helmchart",
							Spec: helm.Spec{
								File:             metadataFilename,
								Name:             chartRelativeMetadataPath,
								Key:              yamlTagPath,
								VersionIncrement: "minor",
							},
						},
					},
				},
			}
			manifests = append(manifests, manifest)
		}
	}

	return manifests, nil
}

func sanitizeRegistryEndpoint(repository string) string {
	// amd64 is only there to avoid warning message as architecture doesn't matter anyway.
	image, err := dockerimageutils.New(repository, "amd64")

	if image.Registry == "registry-1.docker.io" {
		image.Registry = "docker.io"
	}

	if err != nil {
		logrus.Errorln(err)
	}

	return image.Registry
}

func (h Helm) generateSourceDockerImageSpec(image string) dockerimage.Spec {
	dockerimagespec := dockerimage.Spec{
		Image: image,
		// Use versionFilter
		// versionFilter:
		// kind: semver
	}

	registry := sanitizeRegistryEndpoint(image)

	credential, found := h.spec.Auths[registry]

	switch found {
	case true:
		if credential.Password != "" {
			dockerimagespec.Password = credential.Password
		}
		if credential.Token != "" {
			dockerimagespec.Token = credential.Token
		}
		if credential.Username != "" {
			dockerimagespec.Username = credential.Username
		}
	default:

		registryAuths := []string{}

		for endpoint := range h.spec.Auths {
			logrus.Printf("Endpoint:\t%q\n", endpoint)
			registryAuths = append(registryAuths, endpoint)
		}

		warningMessage := fmt.Sprintf(
			"no credentials found for docker registry %q hosting image %q, among %q",
			registry,
			image,
			strings.Join(registryAuths, ","))

		if len(registryAuths) == 0 {
			warningMessage = fmt.Sprintf("no credentials found for docker registry %q hosting image %q",
				registry,
				image)
		}

		logrus.Warning(warningMessage)
	}

	return dockerimagespec
}
