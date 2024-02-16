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

//nolint:funlen
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

		appendImages := func(registry, repository, tag, yamlRegistryPath, yamlRepositoryPath, yamlTagPath string) {
			// In case a digest is specified in the repository, we want to remove it
			if strings.Contains(repository, "@sha256") {
				repository = strings.Split(repository, "@")[0]
			}

			// In case a digest is specified in the tag, we want to remove it
			if strings.Contains(tag, "@sha256") {
				tagArray := strings.Split(tag, "@")
				if len(tagArray) > 1 {
					tag = tagArray[0]
				}
			}

			if repository == "" {
				return
			}

			images = append(images, imageData{
				registry:           registry,
				repository:         repository,
				tag:                tag,
				yamlRegistryPath:   yamlRegistryPath,
				yamlRepositoryPath: yamlRepositoryPath,
				yamlTagPath:        yamlTagPath,
			})
		}

		appendImages(
			values.Image.Registry,
			values.Image.Repository,
			values.Image.Tag,
			"$.image.registry",
			"$.image.repository",
			"$.image.tag")

		for id := range values.Images {
			appendImages(
				values.Images[id].Registry,
				values.Images[id].Repository,
				values.Images[id].Tag,
				fmt.Sprintf("$.images.%s.registry", id),
				fmt.Sprintf("$.images.%s.repository", id),
				fmt.Sprintf("$.images.%s.tag", id),
			)
		}

		for _, image := range images {

			// Compose the container source considering the registry and repository
			var imageName string
			if image.registry == "" {
				imageName = image.repository
			} else {
				imageName = strings.Join([]string{
					strings.Trim(image.registry, "/"),
					strings.Trim(image.repository, "/"),
				}, "/")
			}

			imageSourceSlug := strings.ReplaceAll(imageName, "/", "_")

			// Try to be smart by detecting the best versionfilter
			sourceSpec := dockerimage.NewDockerImageSpecFromImage(imageName, image.tag, h.spec.Auths)

			versionFilterKind := h.versionFilter.Kind
			versionFilterPattern := h.versionFilter.Pattern
			tagFilter := "*"

			if sourceSpec != nil {
				versionFilterKind = sourceSpec.VersionFilter.Kind
				versionFilterPattern = sourceSpec.VersionFilter.Pattern
				tagFilter = sourceSpec.TagFilter
			}

			// If a versionfilter is specified in the manifest then we want to be sure that it takes precedence
			if !h.spec.VersionFilter.IsZero() {
				versionFilterKind = h.versionFilter.Kind
				versionFilterPattern, err = h.versionFilter.GreaterThanPattern(image.tag)
				tagFilter = ""
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

			var tmpl *template.Template
			if h.digest && sourceSpec != nil {
				tmpl, err = template.New("manifest").Parse(manifestTemplateDigestAndLatest)
				if err != nil {
					return nil, err
				}
			} else if h.digest && sourceSpec == nil {
				tmpl, err = template.New("manifest").Parse(manifestTemplateDigest)
				if err != nil {
					return nil, err
				}
			} else if !h.digest && sourceSpec != nil {
				tmpl, err = template.New("manifest").Parse(manifestTemplateLatest)
				if err != nil {
					return nil, err
				}
			} else {
				logrus.Infoln("No source spec detected")
				return nil, nil
			}

			params := struct {
				ImageName                   string
				ImageTag                    string
				ChartName                   string
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
				SourceVersionFilterKind     string
				SourceVersionFilterPattern  string
				SourceImageName             string
				SourceTagFilter             string
				TargetID                    string
				TargetKey                   string
				TargetFile                  string
				TargetChartName             string
				TargetChartVersionIncrement string
				File                        string
				ScmID                       string
			}{
				ImageName:                   imageName,
				ImageTag:                    image.tag,
				ChartName:                   chartName,
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
				SourceVersionFilterKind:     versionFilterKind,
				SourceVersionFilterPattern:  versionFilterPattern,
				SourceImageName:             imageName,
				SourceTagFilter:             tagFilter,
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
				fmt.Println(err)
				return nil, err
			}

			manifests = append(manifests, manifest.Bytes())
		}
	}

	return manifests, nil
}
