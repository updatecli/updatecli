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

type service struct {
	Image string

	// platform defines the target platform containers for this service will run on
	Platform string
}

type dockerComposeSpec struct {
	Services map[string]service
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
		basename := filepath.Base(relativeFoundDockerComposeFile)

		// Retrieve chart dependencies for each chart

		spec, err := getDockerComposeSpecFromFile(foundDockerComposefile)
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
			if service.Image == "" {
				continue
			}

			// For the time being, it's not possible to retrieve a list of tag for a specific digest
			// without a significant amount f api call. More information on following issue
			// https://github.com/google/go-containerregistry/issues/1297
			// until a better solution, we don't handle docker image digest
			if strings.Contains(service.Image, "@sha256") {
				logrus.Debugf("Docker Digest is not supported at the moment for %q", service.Image)
				continue
			}

			serviceImageArray := strings.Split(service.Image, ":")

			// Get container image name and tag
			serviceImageName := ""
			serviceImageTag := ""
			switch len(serviceImageArray) {
			case 2:
				serviceImageName = serviceImageArray[0]
				serviceImageTag = serviceImageArray[1]
			case 1:
				serviceImageName = serviceImageArray[0]
			}

			manifestName := fmt.Sprintf("Bump %q Docker compose service image version for %q",
				serviceImageName,
				relativeFoundDockerComposeFile)

			_, arch, _ := parsePlatform(service.Platform)

			// Test if the ignore rule based on path is respected
			if len(h.spec.Ignore) > 0 {
				if h.spec.Ignore.isMatchingRule(
					h.rootDir,
					relativeFoundDockerComposeFile,
					id,
					service.Image,
					arch) {

					logrus.Debugf("Ignoring Docker Compose file %q from %q, as not matching ignore rule(s)\n",
						basename,
						dirname)
					continue
				}
			}

			// Test if the only rule based on path is respected
			if len(h.spec.Only) > 0 {
				if !h.spec.Only.isMatchingRule(
					h.rootDir,
					relativeFoundDockerComposeFile,
					id,
					service.Image,
					arch) {

					logrus.Debugf("Ignoring Docker Compose file %q from %q, as not matching only rule(s)\n",
						basename,
						dirname)
					continue
				}
			}

			sourceSpec := dockerimage.NewDockerImageSpecFromImage(serviceImageName, serviceImageTag, h.spec.Auths)

			if sourceSpec == nil {
				logrus.Infoln("No source spec detected")
				continue
			}

			if arch != "" {
				sourceSpec.Architecture = arch
			}

			manifest := config.Spec{
				Name: manifestName,
				Sources: map[string]source.Config{
					id: {
						ResourceConfig: resource.ResourceConfig{
							Name: fmt.Sprintf("Get latest %q Docker Image Tag", serviceImageName),
							Kind: "dockerimage",
							Spec: *sourceSpec,
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

func parsePlatform(platform string) (os, arch, variant string) {

	p := strings.Split(platform, "/")

	switch len(p) {
	case 3:
		os = p[0]
		arch = p[1]
		variant = p[2]

	case 2:
		os = p[0]
		arch = p[1]

	case 1:
		os = p[0]
	}

	return os, arch, variant
}
