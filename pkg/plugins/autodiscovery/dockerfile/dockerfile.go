package dockerfile

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/source"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
)

var (
	// DefaultFileMatch specifies accepted Helm chart metadata file name
	DefaultFileMatch []string = []string{
		"Dockerfile",
		"Dockerfile.*",
	}
)

func (h Dockerfile) discoverDockerfileManifests() ([]config.Spec, error) {

	var manifests []config.Spec

	foundDockerfiles, err := searchDockerfiles(
		h.rootDir,
		h.filematch)

	if err != nil {
		return nil, err
	}

	for _, foundDockerfile := range foundDockerfiles {

		relativeFoundDockerfile, err := filepath.Rel(h.rootDir, foundDockerfile)
		if err != nil {
			// Let try the next one if it fails
			logrus.Errorln(err)
			continue
		}

		dirname := filepath.Dir(relativeFoundDockerfile)
		basename := filepath.Base(relativeFoundDockerfile)

		instructions, err := parseDockerfile(foundDockerfile)
		if err != nil {
			return nil, err
		}

		if len(instructions) == 0 {
			continue
		}

		for _, instruction := range instructions {

			// For the time being, it's not possible to retrieve a list of tag for a specific digest
			// without a significant amount f api call. More information on following issue
			// https://github.com/google/go-containerregistry/issues/1297
			// until a better solution, we don't handle docker image digest
			if strings.Contains(instruction.image, "@sha256") {
				logrus.Debugf("Docker Digest is not supported at the moment for %q", instruction.image)
				continue
			}

			imageArray := strings.Split(instruction.image, ":")

			// Get container image name and tag
			imageName := ""
			imageTag := ""
			switch len(imageArray) {
			case 2:
				imageName = imageArray[0]
				imageTag = imageArray[1]
			case 1:
				imageName = imageArray[0]
			}

			manifestName := fmt.Sprintf("Bump Docker Image Tag for %q", imageName)

			// Test if the ignore rule based on path is respected
			if len(h.spec.Ignore) > 0 {
				if h.spec.Ignore.isMatchingRule(
					h.rootDir,
					relativeFoundDockerfile,
					instruction.image,
					instruction.arch,
				) {

					logrus.Debugf("Ignoring Dockerfile %q from %q, as matching ignore rule(s)\n",
						basename,
						dirname)
					continue
				}
			}

			// Test if the only rule based on path is respected
			if len(h.spec.Only) > 0 {
				if !h.spec.Only.isMatchingRule(
					h.rootDir,
					relativeFoundDockerfile,
					instruction.image,
					instruction.arch) {

					logrus.Debugf("Ignoring Dockerfile %q from %q, as not matching only rule(s)\n",
						basename,
						dirname)
					continue
				}
			}

			sourceSpec := dockerimage.NewDockerImageSpecFromImage(imageName, imageTag, h.spec.Auths)

			if sourceSpec == nil {
				logrus.Debugln("no source spec detected")
				continue
			}

			if instruction.arch != "" {
				sourceSpec.Architecture = instruction.arch
			}

			if err != nil {
				logrus.Errorln(err)
				continue

			}

			targetMatcher := ""
			targetKeyword := instruction.name

			// Depending on the instruction the matcher will be different
			switch instruction.name {
			case "FROM":
				targetMatcher = imageName
			case "ARG":
				targetMatcher = instruction.value
				targetMatcher = strings.TrimPrefix(targetMatcher, instruction.trimArgPrefix)
				targetMatcher = strings.TrimSuffix(targetMatcher, instruction.trimArgSuffix)
			}

			manifest := config.Spec{
				Name: manifestName,
				Sources: map[string]source.Config{
					imageName: {
						ResourceConfig: resource.ResourceConfig{
							Name: fmt.Sprintf("[%s] Get latest Docker Image Tag", imageName),
							Kind: "dockerimage",
							Spec: *sourceSpec,
						},
					},
				},
				Targets: map[string]target.Config{
					imageName: {
						SourceID: imageName,
						ResourceConfig: resource.ResourceConfig{
							Name: fmt.Sprintf("[%s] Bump Docker Image tag in %q",
								imageName,
								relativeFoundDockerfile),
							Kind: "dockerfile",
							Spec: dockerfile.Spec{
								File: relativeFoundDockerfile,
								Instruction: map[string]string{
									"keyword": targetKeyword,
									"matcher": targetMatcher,
								},
							},
						},
					},
				},
			}

			// Set scmID if defined
			if h.scmID != "" {
				t := manifest.Targets[imageName]
				t.SCMID = h.scmID
				manifest.Targets[imageName] = t
			}
			manifests = append(manifests, manifest)

			manifests = append(manifests, manifest)
		}
	}

	return manifests, nil
}
