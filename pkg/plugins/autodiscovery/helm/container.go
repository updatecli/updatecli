package helm

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
)

type imageRef struct {
	Registry   string
	Repository string
	Tag        string
}

type valuesContent struct {
	Image  imageRef
	Images map[string]imageRef
}

func (h Helm) discoverHelmContainerManifests() ([][]byte, error) {

	var manifests [][]byte

	foundValuesFiles, err := searchChartFiles(
		h.rootDir,
		[]string{"values.yaml", "values.yml"})

	if err != nil {
		return nil, err
	}

	for _, foundValueFile := range foundValuesFiles {

		logrus.Debugf("parsing file %q", foundValueFile)

		relativeFoundValueFile, err := filepath.Rel(h.rootDir, foundValueFile)
		if err != nil {
			// Jump to the next Helm chart if current failed
			logrus.Errorln(err)
			continue
		}

		chartRelativeMetadataPath := filepath.Dir(relativeFoundValueFile)
		chartName := filepath.Base(chartRelativeMetadataPath)

		// Retrieve chart dependencies for each chart
		values, err := getValuesFileContent(foundValueFile)
		if err != nil {
			return nil, err
		}

		if values == nil {
			continue
		}

		type imageData struct {
			registry           string
			repository         string
			tag                string
			yamlRegistryPath   string
			yamlRepositoryPath string
			yamlTagPath        string
		}

		var images []imageData

		if values.Image.Repository != "" && values.Image.Tag != "" {
			// Docker Image Digest isn't supported at this time.
			if strings.HasSuffix(values.Image.Repository, "sha256") {
				logrus.Debugf("Docker image digest detected, skipping as not supported yet")
				continue
			}
			images = append(images, imageData{
				registry:           values.Image.Registry,
				repository:         values.Image.Repository,
				tag:                values.Image.Tag,
				yamlRegistryPath:   "$.image.registry",
				yamlRepositoryPath: "$.image.repository",
				yamlTagPath:        "$.image.tag",
			})

		}

		for id := range values.Images {
			// Docker Image Digest isn't supported at this time.
			if strings.HasSuffix(values.Images[id].Repository, "sha256") {
				logrus.Debugf("Docker image digest detected, skipping as not supported yet")
				continue
			}

			images = append(images, imageData{
				registry:           values.Images[id].Registry,
				repository:         values.Images[id].Repository,
				tag:                values.Images[id].Tag,
				yamlRegistryPath:   fmt.Sprintf("$.images.%s.registry", id),
				yamlRepositoryPath: fmt.Sprintf("$.images.%s.repository", id),
				yamlTagPath:        fmt.Sprintf("$.images.%s.tag", id),
			})

		}

		for _, image := range images {

			// Compose the container source considering the registry and repository
			var imageSource string
			if image.registry == "" {
				imageSource = image.repository
			} else {
				imageSource = strings.Join([]string{
					strings.Trim(image.registry, "/"),
					strings.Trim(image.repository, "/"),
				}, "/")
			}

			imageSourceSlug := strings.ReplaceAll(imageSource, "/", "_")

			// Try to be smart by detecting the best versionfilter
			sourceSpec := dockerimage.NewDockerImageSpecFromImage(imageSource, image.tag, h.spec.Auths)
			if sourceSpec == nil {
				continue
			}

			// If a versionfilter is specified in the manifest then we want to be sure that it takes precedence
			if !h.spec.VersionFilter.IsZero() {
				sourceSpec.VersionFilter.Kind = h.versionFilter.Kind
				sourceSpec.VersionFilter.Pattern, err = h.versionFilter.GreaterThanPattern(image.tag)
				sourceSpec.TagFilter = ""

				if err != nil {
					logrus.Debugf("building version filter pattern: %s", err)
					sourceSpec.VersionFilter.Pattern = "*"
				}
			}

			// Test if the ignore rule based on path is respected
			if len(h.spec.Ignore) > 0 {
				if h.spec.Ignore.isMatchingRules(h.rootDir, chartRelativeMetadataPath, "", "", sourceSpec.Image, sourceSpec.Tag) {
					logrus.Debugf("Ignoring container version update from file %q, as matching ignore rule(s)\n", relativeFoundValueFile)
					continue
				}
			}

			// Test if the only rule based on path is respected
			if len(h.spec.Only) > 0 {
				if !h.spec.Only.isMatchingRules(h.rootDir, chartRelativeMetadataPath, "", "", sourceSpec.Image, sourceSpec.Tag) {
					logrus.Debugf("Ignoring container version update from %q, as not matching only rule(s)\n", relativeFoundValueFile)
					continue
				}
			}

			tmpl, err := template.New("manifest").Parse(containerManifest)
			if err != nil {
				logrus.Errorln(err)
				continue
			}

			params := struct {
				ManifestName                string
				HasRegistry                 bool
				ConditionRegistryID         string
				ConditionRegistryKey        string
				ConditionRegistryName       string
				ConditionRegistryValue      string
				ConditionRepositoryID       string
				ConditionRepositoryKey      string
				ConditionRepositoryName     string
				ConditionRepositoryValue    string
				SourceID                    string
				SourceName                  string
				SourceVersionFilterKind     string
				SourceVersionFilterPattern  string
				SourceImageName             string
				SourceTagFilter             string
				TargetName                  string
				TargetID                    string
				TargetKey                   string
				TargetFile                  string
				TargetChartName             string
				TargetChartVersionIncrement string
				File                        string
				ScmID                       string
			}{
				ManifestName:                fmt.Sprintf("Bump Docker image %q for Helm chart %q", imageSource, chartName),
				HasRegistry:                 image.registry != "",
				ConditionRegistryID:         imageSourceSlug + "-registry",
				ConditionRegistryKey:        image.yamlRegistryPath,
				ConditionRegistryName:       fmt.Sprintf("Ensure container registry %q is specified", image.registry),
				ConditionRegistryValue:      image.registry,
				ConditionRepositoryID:       imageSourceSlug + "-repository",
				ConditionRepositoryKey:      image.yamlRepositoryPath,
				ConditionRepositoryName:     fmt.Sprintf("Ensure container repository %q is specified", image.repository),
				ConditionRepositoryValue:    image.repository,
				SourceID:                    imageSourceSlug,
				SourceName:                  fmt.Sprintf("Get latest %q container tag", imageSource),
				SourceVersionFilterKind:     sourceSpec.VersionFilter.Kind,
				SourceVersionFilterPattern:  sourceSpec.VersionFilter.Pattern,
				SourceImageName:             sourceSpec.Image,
				SourceTagFilter:             sourceSpec.TagFilter,
				TargetName:                  fmt.Sprintf("Bump container image tag for image %q in chart %q", imageSource, chartName),
				TargetID:                    imageSourceSlug,
				TargetKey:                   image.yamlTagPath,
				TargetChartName:             chartRelativeMetadataPath,
				TargetChartVersionIncrement: h.spec.VersionIncrement,
				TargetFile:                  filepath.Base(foundValueFile),
				File:                        relativeFoundValueFile,
				ScmID:                       h.scmID,
			}

			manifest := bytes.Buffer{}
			if err := tmpl.Execute(&manifest, params); err != nil {
				return nil, err
			}

			manifests = append(manifests, manifest.Bytes())
		}
	}

	return manifests, nil
}
