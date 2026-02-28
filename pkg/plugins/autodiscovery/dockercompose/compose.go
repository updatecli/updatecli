package dockercompose

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
)

// DefaultFileMatch specifies the default file shell pattern to identify Docker Compose files
// Ref. https://pkg.go.dev/path/filepath#Match and https://go.dev/play/p/y2b7tt03r8Q to test
var DefaultFilePattern = []string{
	"docker-compose*.y*ml",
	"compose*.y*ml",
}

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

	searchFromDir := d.rootDir
	// If the spec.RootDir is an absolute path, then it as already been set
	// correctly in the New function.
	if d.spec.RootDir != "" && !path.IsAbs(d.spec.RootDir) {
		searchFromDir = filepath.Join(d.rootDir, d.spec.RootDir)
	}

	foundDockerComposeFiles, err := searchDockerComposeFiles(searchFromDir, d.filematch)
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
			} else if (strings.Contains(svc.Spec.Image, "${")) && (strings.Contains(svc.Spec.Image, "}")) {
				logrus.Debugf("Skipping image %q as it contains environment variable, which is not supported at the moment", svc.Spec.Image)
				continue
			}

			imageName, imageTag, imageDigest, err := dockerimage.ParseOCIReferenceInfo(svc.Spec.Image)
			if err != nil {
				return nil, fmt.Errorf("parsing image %q: %s", svc.Spec.Image, err)
			}

			/*
				For the time being, it's not possible to retrieve a list of tag for a specific digest
				without a significant amount f api call. More information on following issue
				https://github.com/google/go-containerregistry/issues/1297
				until a better solution, we don't handle docker image digest
			*/
			if imageDigest != "" && imageTag == "" {
				logrus.Debugf("docker digest without specified tag is not supported at the moment for %q", svc.Spec.Image)
				continue
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

			sourceSpec := dockerimage.NewDockerImageSpecFromImage(imageName, imageTag, d.spec.Auths)

			versionFilterKind := d.versionFilter.Kind
			versionFilterPattern := d.versionFilter.Pattern
			versionFilterRegex := d.versionFilter.Regex
			tagFilter := "*"
			architecture := ""

			registryUsername := ""
			registryPassword := ""
			registryToken := ""

			if sourceSpec != nil {
				versionFilterKind = sourceSpec.VersionFilter.Kind
				versionFilterPattern = sourceSpec.VersionFilter.Pattern
				versionFilterRegex = sourceSpec.VersionFilter.Regex
				tagFilter = sourceSpec.TagFilter
				architecture = sourceSpec.Architecture

				registryUsername = sourceSpec.Username
				registryPassword = sourceSpec.Password
				registryToken = sourceSpec.Token
			}

			// If a versionfilter is specified in the manifest then we want to be sure that it takes precedence
			if !d.spec.VersionFilter.IsZero() {
				versionFilterKind = d.versionFilter.Kind
				versionFilterPattern, err = d.versionFilter.GreaterThanPattern(imageTag)
				versionFilterRegex = d.versionFilter.Regex
				tagFilter = ""
				if err != nil {
					logrus.Debugf("building version filter pattern: %s", err)
					sourceSpec.VersionFilter.Pattern = "*"
				}
			}

			if arch != "" {
				architecture = arch
			}

			var tmpl *template.Template
			if d.digest && sourceSpec != nil {
				tmpl, err = template.New("manifest").Parse(manifestTemplateDigestAndLatest)
				if err != nil {
					return nil, err
				}
			} else if d.digest && sourceSpec == nil {
				tmpl, err = template.New("manifest").Parse(manifestTemplateDigest)
				if err != nil {
					return nil, err
				}
			} else if !d.digest && sourceSpec != nil {
				tmpl, err = template.New("manifest").Parse(manifestTemplateLatest)
				if err != nil {
					return nil, err
				}
			} else {
				logrus.Infoln("No source spec detected")
				return nil, nil
			}

			params := struct {
				ActionID             string
				ImageName            string
				ImageTag             string
				ImageArchitecture    string
				SourceID             string
				RegistryUsername     string
				RegistryPassword     string
				RegistryToken        string
				TargetID             string
				TargetFile           string
				TargetKey            string
				TargetPrefix         string
				TagFilter            string
				VersionFilterKind    string
				VersionFilterPattern string
				VersionFilterRegex   string
				ScmID                string
			}{
				ActionID:             d.actionID,
				ImageName:            imageName,
				ImageTag:             imageTag,
				ImageArchitecture:    architecture,
				RegistryUsername:     registryUsername,
				RegistryPassword:     registryPassword,
				RegistryToken:        registryToken,
				SourceID:             svc.Name,
				TargetID:             svc.Name,
				TargetFile:           relativeFoundDockerComposeFile,
				TargetKey:            fmt.Sprintf("$.services.%s.image", svc.Name),
				TargetPrefix:         imageName + ":",
				TagFilter:            tagFilter,
				VersionFilterKind:    versionFilterKind,
				VersionFilterPattern: versionFilterPattern,
				VersionFilterRegex:   versionFilterRegex,
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
