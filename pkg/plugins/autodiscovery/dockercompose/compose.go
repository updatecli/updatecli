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

func (d DockerCompose) discoverDockerComposeImageManifests() ([][]byte, error) {
	var manifests [][]byte

	foundDockerComposeFiles, err := searchDockerComposeFiles(d.rootDir, d.filematch)
	if err != nil {
		return nil, err
	}

	for _, foundDockerComposefile := range foundDockerComposeFiles {
		relativeFoundDockerComposeFile, err := filepath.Rel(d.rootDir, foundDockerComposefile)
		logrus.Debugf("parsing file %q", foundDockerComposefile)
		if err != nil {
			// Let's try the next one if it fails
			logrus.Debugln(err)
			continue
		}

		dirname := filepath.Dir(relativeFoundDockerComposeFile)
		basename := filepath.Base(relativeFoundDockerComposeFile)

		// Retrieve chart dependencies for each chart
		svcList, err := getDockerComposeSpecFromFile(foundDockerComposefile)
		if err != nil {
			logrus.Debugln(err)
			continue
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
			if len(d.spec.Ignore) > 0 {
				if d.spec.Ignore.isMatchingRule(
					d.rootDir,
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
			if len(d.spec.Only) > 0 {
				if !d.spec.Only.isMatchingRule(
					d.rootDir,
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

			sourceSpec := dockerimage.NewDockerImageSpecFromImage(serviceImageName, serviceImageTag, d.spec.Auths)
			if sourceSpec == nil {
				logrus.Infoln("No source spec detected")
				continue
			}

			// If a versionfilter is specified in the manifest then we want to be sure that it takes precedence
			if !d.spec.VersionFilter.IsZero() {
				sourceSpec.VersionFilter.Kind = d.versionFilter.Kind
				sourceSpec.VersionFilter.Pattern, err = d.versionFilter.GreaterThanPattern(serviceImageTag)
				sourceSpec.TagFilter = ""
				if err != nil {
					logrus.Debugf("building version filter pattern: %s", err)
					sourceSpec.VersionFilter.Pattern = "*"
				}
			}

			if arch != "" {
				sourceSpec.Architecture = arch
			}

			tmpl, err := template.New("manifest").Parse(manifestTemplate)
			if err != nil {
				logrus.Debugln(err)
				continue
			}

			params := struct {
				ManifestName         string
				ImageName            string
				ImageArchitecture    string
				SourceID             string
				SourceName           string
				TargetID             string
				TargetFile           string
				TargetName           string
				TargetKey            string
				TargetPrefix         string
				TagFilter            string
				VersionFilterKind    string
				VersionFilterPattern string
				ScmID                string
			}{
				ManifestName:         fmt.Sprintf("Bump Docker image tag for %q", serviceImageName),
				ImageName:            serviceImageName,
				ImageArchitecture:    sourceSpec.Architecture,
				SourceID:             svc.Name,
				SourceName:           fmt.Sprintf("[%s] Get latest Docker image tag", serviceImageName),
				TargetID:             svc.Name,
				TargetName:           fmt.Sprintf("[%s] Bump Docker image tag in %q", serviceImageName, relativeFoundDockerComposeFile),
				TargetFile:           relativeFoundDockerComposeFile,
				TargetKey:            fmt.Sprintf("$.services.%s.image", svc.Name),
				TargetPrefix:         serviceImageName + ":",
				TagFilter:            sourceSpec.TagFilter,
				VersionFilterKind:    sourceSpec.VersionFilter.Kind,
				VersionFilterPattern: sourceSpec.VersionFilter.Pattern,
				ScmID:                d.scmID,
			}

			manifest := bytes.Buffer{}
			if err := tmpl.Execute(&manifest, params); err != nil {
				logrus.Debugln(err)
				continue
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
