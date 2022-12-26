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

		// Test if the ignore rule based on path doesn't match
		if len(h.spec.Ignore) > 0 && h.spec.Ignore.isMatchingIgnoreRule(h.rootDir, relativeFoundValueFile) {
			logrus.Debugf("Ignoring Helm Chart %q from %q, as not matching rule(s)\n",
				chartName,
				chartRelativeMetadataPath)
			continue
		}

		// Test if the only rule based on path match
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
			// Docker Image Digest isn't supported at this time.
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
			// Docker Image Digest isn't supported at this time.
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
			sourceSpec := dockerimage.NewDockerImageSpecFromImage(image.repository, image.tag, h.spec.Auths)

			if sourceSpec == nil {
				continue
			}

			tmpl, err := template.New("manifest").Parse(containerManifest)
			if err != nil {
				logrus.Errorln(err)
				continue
			}

			params := struct {
				ManifestName               string
				ConditionID                string
				ConditionKey               string
				ConditionValue             string
				ConditionName              string
				SourceID                   string
				SourceName                 string
				SourceVersionFilterKind    string
				SourceVersionFilterPattern string
				SourceImageName            string
				SourceTagFilter            string
				TargetName                 string
				TargetID                   string
				TargetKey                  string
				TargetChartName            string
				File                       string
				ScmID                      string
			}{
				ManifestName:               fmt.Sprintf("Bump Docker image %q for Helm chart %q", image.repository, chartName),
				ConditionID:                image.repository,
				ConditionKey:               image.yamlRepositoryPath,
				ConditionName:              fmt.Sprintf("Ensure container repository %q is specified", image.repository),
				ConditionValue:             image.repository,
				SourceID:                   image.repository,
				SourceName:                 fmt.Sprintf("Get latest %q container tag", image.repository),
				SourceVersionFilterKind:    "semver",
				SourceVersionFilterPattern: "'*'",
				SourceImageName:            sourceSpec.Image,
				SourceTagFilter:            fmt.Sprintf("'%s'", sourceSpec.TagFilter),
				TargetName:                 fmt.Sprintf("Bump container image tag for image %q in chart %q", image.repository, chartName),
				TargetID:                   image.repository,
				TargetKey:                  image.yamlTagPath,
				TargetChartName:            chartRelativeMetadataPath,
				File:                       relativeFoundValueFile,
				ScmID:                      h.scmID,
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
