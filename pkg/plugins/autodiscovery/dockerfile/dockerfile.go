package dockerfile

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
)

var (
	// DefaultFileMatch specifies accepted Helm chart metadata file name
	DefaultFileMatch []string = []string{
		"Dockerfile",
		"Dockerfile.*",
	}
)

func (h Dockerfile) discoverDockerfileManifests() ([][]byte, error) {

	var manifests [][]byte

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
			// Depending on the instruction the matcher will be different
			switch instruction.name {
			case "FROM":
				targetMatcher = imageName
			case "ARG":
				targetMatcher = instruction.value
				targetMatcher = strings.TrimPrefix(targetMatcher, instruction.trimArgPrefix)
				targetMatcher = strings.TrimSuffix(targetMatcher, instruction.trimArgSuffix)
			}

			tmpl, err := template.New("manifest").Parse(manifestTemplate)
			if err != nil {
				logrus.Errorln(err)
				continue
			}

			params := struct {
				ManifestName string
				ImageName            string
				SourceID             string
				TargetID             string
				TargetFile           string
				TargetName           string
				TargetKeyword        string
				TargetMatcher 		 string
				TagFilter            string
				VersionFilterKind    string
				VersionFilterPattern string
				ScmID                string
			}{
				ManifestName:         fmt.Sprintf("Bump Docker image tag for %q", imageName),
				ImageName:            imageName,
				SourceID:             imageName,
				TargetID:             imageName,
				TargetName:           fmt.Sprintf("[%s] Bump Docker image tag in %q",imageName,relativeFoundDockerfile),
				TargetFile:           relativeFoundDockerfile,
				TargetKeyword:        instruction.name,
				TargetMatcher: 		  targetMatcher,
				TagFilter:            sourceSpec.TagFilter,
				VersionFilterKind:    sourceSpec.VersionFilter.Kind,
				VersionFilterPattern: sourceSpec.VersionFilter.Pattern,
				ScmID:                h.scmID,
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
