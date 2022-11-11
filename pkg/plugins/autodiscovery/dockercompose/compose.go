package dockercompose

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/core/transformer"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
	"github.com/updatecli/updatecli/pkg/plugins/resources/yaml"
)

var (
	// DefaultFilePattern specifies accepted Helm chart metadata file name
	DefaultFilePattern [4]string = [4]string{
		"docker-compose.yaml",
		"docker-compose.*.yaml",
		"docker-compose.yml",
		"docker-compose.*.yml"}
)

type Service struct {
	Image string
}

type DockerComposeSpec struct {
	Services map[string]Service
}

func (h DockerCompose) discoverDockerComposeImageManifests() ([]config.Spec, error) {

	var manifests []config.Spec

	foundDockerComposeFiles, err := searchDockerComposeFiles(
		h.rootDir,
		DefaultFilePattern[:])

	if err != nil {
		return nil, err
	}

	for _, foundDockerComposefile := range foundDockerComposeFiles {

		relativeFoundDockerComposeFile, err := filepath.Rel(h.rootDir, foundDockerComposefile)
		if err != nil {
			// Let's try the next one if it fails
			logrus.Errorln(err)
			continue
		}

		dirname := filepath.Dir(relativeFoundDockerComposeFile)
		basename := filepath.Base(dirname)

		// Test if the ignore rule based on path is respected
		if len(h.spec.Ignore) > 0 && h.spec.Ignore.isMatchingIgnoreRule(h.rootDir, relativeFoundDockerComposeFile) {
			logrus.Debugf("Ignoring Docker Compose file %q from %q, as not matching rule(s)\n",
				basename,
				dirname)
			continue
		}

		// Test if the only rule based on path is respected
		if len(h.spec.Only) > 0 && !h.spec.Only.isMatchingOnlyRule(h.rootDir, relativeFoundDockerComposeFile) {
			logrus.Debugf("Ignoring Docker Compose file %q from %q, as not matching rule(s)\n",
				basename,
				dirname)
			continue
		}

		// Retrieve chart dependencies for each chart

		spec, err := getDockerComposeData(foundDockerComposefile)
		if err != nil {
			return nil, err
		}

		if spec == nil {
			continue
		}

		if len(spec.Services) == 0 {
			continue
		}

		for id, service := range spec.Services {

			if strings.Contains(service.Image, "@sha256") {
				logrus.Debugf("Docker Digest is not supported at the moment for %q", service.Image)
				continue
			}

			serviceImageArray := strings.Split(service.Image, ":")
			serviceImageName := serviceImageArray[0]

			manifestName := fmt.Sprintf("Bump %q Docker compose service image version for %q",
				serviceImageName,
				relativeFoundDockerComposeFile)

			sourceSpec := dockerimage.NewDockerImageSpecFromImage(serviceImageName, h.spec.Auths)

			manifest := config.Spec{
				Name: manifestName,
				Sources: map[string]source.Config{
					id: {
						ResourceConfig: resource.ResourceConfig{
							Name: fmt.Sprintf("Get latest %q Docker Image Tag", serviceImageName),
							Kind: "dockerimage",
							Spec: sourceSpec,
						},
					},
				},
				Targets: map[string]target.Config{
					id: {
						SourceID: id,
						ResourceConfig: resource.ResourceConfig{
							Name: fmt.Sprintf("Bump %q Docker Image tag for docker compose file %q",
								serviceImageName,
								relativeFoundDockerComposeFile),
							Kind: "yaml",
							Spec: yaml.Spec{
								File: foundDockerComposefile,
								Key:  fmt.Sprintf("services.%s.image", id),
							},
							Transformers: transformer.Transformers{
								transformer.Transformer{
									AddPrefix: serviceImageName + ":",
								},
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
