package dockerfile

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"path"

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

func (d Dockerfile) discoverDockerfileManifests() ([][]byte, error) {

	var manifests [][]byte

	searchFromDir := d.rootDir
	// If the spec.RootDir is an absolute path, then it as already been set
	// correctly in the New function.
	if d.spec.RootDir != "" && !path.IsAbs(d.spec.RootDir) {
		searchFromDir = filepath.Join(d.rootDir, d.spec.RootDir)
	}

	foundDockerfiles, err := searchDockerfiles(
		searchFromDir,
		d.filematch)

	if err != nil {
		return nil, err
	}

	for _, foundDockerfile := range foundDockerfiles {

		logrus.Debugf("parsing file %q", foundDockerfile)
		relativeFoundDockerfile, err := filepath.Rel(d.rootDir, foundDockerfile)
		if err != nil {
			// Let try the next one if it fails
			logrus.Debugln(err)
			continue
		}

		dirname := filepath.Dir(relativeFoundDockerfile)
		basename := filepath.Base(relativeFoundDockerfile)

		instructions, err := parseDockerfile(foundDockerfile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		if len(instructions) == 0 {
			continue
		}

		for _, instruction := range instructions {

			imageName, imageTag, imageDigest, err := dockerimage.ParseOCIReferenceInfo(instruction.image)
			if err != nil {
				return nil, fmt.Errorf("parsing image %q: %s", instruction.image, err)
			}

			/*
				// For the time being, it's not possible to retrieve a list of tag for a specific digest
				// without a significant amount f api call. More information on following issue
				// https://github.com/google/go-containerregistry/issues/1297
				// until a better solution, we don't handle docker image digest
			*/
			if imageDigest != "" && imageTag == "" {
				logrus.Debugf("docker digest without specified tag is not supported at the moment for %q", instruction.image)
				continue
			}

			// Test if the ignore rule based on path is respected
			if len(d.spec.Ignore) > 0 {
				if d.spec.Ignore.isMatchingRule(
					d.rootDir,
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
			if len(d.spec.Only) > 0 {
				if !d.spec.Only.isMatchingRule(
					d.rootDir,
					relativeFoundDockerfile,
					instruction.image,
					instruction.arch) {

					logrus.Debugf("Ignoring Dockerfile %q from %q, as not matching only rule(s)\n",
						basename,
						dirname)
					continue
				}
			}

			sourceSpec := dockerimage.NewDockerImageSpecFromImage(imageName, imageTag, d.spec.Auths)

			if sourceSpec == nil && !d.digest {
				logrus.Debugln("no source spec detected")
				continue
			}

			versionFilterKind := d.versionFilter.Kind
			versionFilterPattern := d.versionFilter.Pattern
			tagFilter := "*"

			if sourceSpec != nil {
				versionFilterKind = sourceSpec.VersionFilter.Kind
				versionFilterPattern = sourceSpec.VersionFilter.Pattern
				tagFilter = sourceSpec.TagFilter
			}

			// If a versionfilter is specified in the manifest then we want to be sure that it takes precedence
			if !d.spec.VersionFilter.IsZero() {
				versionFilterKind = d.versionFilter.Kind
				versionFilterPattern, err = d.versionFilter.GreaterThanPattern(imageTag)
				tagFilter = ""
				if err != nil {
					logrus.Debugf("building version filter pattern: %s", err)
					sourceSpec.VersionFilter.Pattern = "*"
				}
			}

			if err != nil {
				logrus.Debugln(err)
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
				ImageName            string
				ImageTag             string
				ScmID                string
				SourceID             string
				TargetID             string
				TargetFile           string
				TargetKeyword        string
				TargetMatcher        string
				TagFilter            string
				VersionFilterKind    string
				VersionFilterPattern string
			}{
				ImageName:            imageName,
				ImageTag:             imageTag,
				ScmID:                d.scmID,
				SourceID:             imageName,
				TargetID:             imageName,
				TargetFile:           relativeFoundDockerfile,
				TargetKeyword:        instruction.name,
				TargetMatcher:        targetMatcher,
				TagFilter:            tagFilter,
				VersionFilterKind:    versionFilterKind,
				VersionFilterPattern: versionFilterPattern,
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
