package dockercompose

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
)

const (
	// DefaultFileMatch specifies the default file shell pattern to identify Docker Compose files
	// Ref. https://pkg.go.dev/path/filepath#Match and https://go.dev/play/p/y2b7tt03r8Q to test
	DefaultFilePattern string = "*docker-compose*.y*ml"
)

type dockerComposeServiceSpec struct {
	Image string
	// platform defines the target platform containers for this service will run on
	Platform string
}

type dockerComposeService struct {
	Name string
	Spec dockerComposeServiceSpec
}

type dockercomposeServicesList []dockerComposeService

func (h DockerCompose) discoverDockerComposeImageManifests() ([][]byte, error) {
	var manifests [][]byte

	foundDockerComposeFiles, err := searchDockerComposeFiles(h.rootDir, h.filematch)
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
		svcList, err := getDockerComposeSpecFromFile(foundDockerComposefile)
		if err != nil {
			return nil, err
		}

		if len(svcList) == 0 {
			continue
		}

		for _, svc := range svcList {
			if svc.Spec.Image == "" {
				continue
			}

			// For the time being, it's not possible to retrieve a list of tag for a specific digest
			// without a significant amount f api call. More information on following issue
			// https://github.com/google/go-containerregistry/issues/1297
			// until a better solution, we don't handle docker image digest
			if strings.Contains(svc.Spec.Image, "@sha256") {
				logrus.Debugf("Docker Digest is not supported at the moment for %q", svc.Spec.Image)
				continue
			}

			serviceImageArray := strings.Split(svc.Spec.Image, ":")

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

			_, arch, _ := parsePlatform(svc.Spec.Platform)

			// Test if the ignore rule based on path is respected
			if len(h.spec.Ignore) > 0 {
				if h.spec.Ignore.isMatchingRule(
					h.rootDir,
					relativeFoundDockerComposeFile,
					svc.Name,
					svc.Spec.Image,
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
					svc.Name,
					svc.Spec.Image,
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

			tmpl, err := template.New("manifest").Parse(manifestTemplate)
			if err != nil {
				logrus.Errorln(err)
				continue
			}

			params := struct {
				ImageName            string
				SourceID             string
				TargetID             string
				TargetFile           string
				TargetName           string
				TargetKey            string
				TargetPrefix         string
				TagFilter            string
				VersionFilterKind    string
				VersionFilterPattern string
				ScmID                string
				ActionID             string
			}{
				ImageName:            serviceImageName,
				SourceID:             serviceImageName,
				TargetID:             serviceImageName,
				TargetName:           fmt.Sprintf("[%s] Bump Docker Image tag in %q", serviceImageName, relativeFoundDockerComposeFile),
				TargetFile:           relativeFoundDockerComposeFile,
				TargetKey:            fmt.Sprintf("services.%s.image", svc.Name),
				TargetPrefix:         serviceImageName + ":",
				TagFilter:            sourceSpec.TagFilter,
				VersionFilterKind:    sourceSpec.VersionFilter.Kind,
				VersionFilterPattern: sourceSpec.VersionFilter.Pattern,
				ScmID:                h.scmID,
				ActionID:             h.actionID,
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
