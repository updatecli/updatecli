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
			// Docker Image Digest is not supported at this time.
			if strings.HasSuffix(values.Image.Repository, "sha256") {
				logrus.Debugf("Docker image digest detected, skipping as not supported yet")
				continue
			}
			images = append(images, imageData{
				repository:         values.Image.Repository,
				tag:                values.Image.Tag,
				yamlRepositoryPath: "image.repository",
				yamlTagPath:        "image.tag",
			})

		}

		for id := range values.Images {
			// Docker Image Digest is not supported at this time.
			if strings.HasSuffix(values.Images[id].Repository, "sha256") {
				logrus.Debugf("Docker image digest detected, skipping as not supported yet")
				continue
			}

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

			sourceSpec := dockerimage.NewDockerImageSpecFromImage(image.repository, image.tag, h.spec.Auths)

			if sourceSpec == nil {
				continue
			}

			manifestName := fmt.Sprintf("Bump Docker Image %q for Helm Chart %q", image.repository, chartName)

			manifest := config.Spec{
				Name: manifestName,
				Sources: map[string]source.Config{
					sourceID: {
						ResourceConfig: resource.ResourceConfig{
							Name: fmt.Sprintf("Get latest %q Container tag", image.repository),
							Kind: "dockerimage",
							Spec: sourceSpec,
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
