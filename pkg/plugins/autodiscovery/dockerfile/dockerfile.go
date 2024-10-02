package dockerfile

import (
	"bytes"
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
	// GlobalIgnore specifies a list of globally ignored docker image, like the `scratch` image which
	// never needs to be updated as it's not a "real" image
	ScratchIgnore MatchingRule = MatchingRule{
		Images: []string{"scratch"},
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

		instructions, args, err := parseDockerfile(foundDockerfile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		if len(instructions) == 0 {
			continue
		}

		// Let's build a list of stage name to ignore
		ignoreStage := []string{}
		for _, instruction := range instructions {
			if instruction.Alias != "" {
				ignoreStage = append(ignoreStage, instruction.Alias)
			}
		}
		globalIgnore := MatchingRules{ScratchIgnore}
		if len(ignoreStage) > 0 {
			globalIgnore = append(globalIgnore, MatchingRule{
				Images: ignoreStage,
			})
		}

		for _, instruction := range instructions {
			// Replace args when needed

			targetMatcher := instruction.Image
			targetInstruction := instruction.Keyword
			image := instruction.Image
			tag := instruction.Tag
			digest := instruction.Digest
			platform := instruction.Platform
			for arg_type, fromArg := range instruction.Args {
				arg, ok := args[fromArg.Name]
				if !ok {
					continue
				}
				switch arg_type {
				case "image":
					image = strings.Replace(image, "${"+fromArg.Name+"}", arg.Value, -1)
				case "tag":
					value := arg.Value
					if value == "" {
						// Consider latest
						value = "latest"
					}
					tag = strings.Replace(tag, "${"+fromArg.Name+"}", value, -1)
				case "digest":
					digest = strings.Replace(digest, "${"+fromArg.Name+"}", arg.Value, -1)
				case "platform":
					platform = strings.Replace(platform, "${"+fromArg.Name+"}", arg.Value, -1)
				}
				targetMatcher = fromArg.Name
				targetInstruction = arg.Keyword
			}

			/*
				// For the time being, it's not possible to retrieve a list of tag for a specific digest
				// without a significant amount f api call. More information on following issue
				// https://github.com/google/go-containerregistry/issues/1297
				// until a better solution, we don't handle docker image digest
			*/
			if digest != "" && tag == "" {
				logrus.Debugf("docker digest without specified tag is not supported at the moment for %q", image)
				continue
			}

			// Remove globally ignore images
			if len(globalIgnore) > 0 {
				if globalIgnore.isMatchingRule(d.rootDir, relativeFoundDockerfile, image, platform) {
					logrus.Debugf("Ignoring Dockerfile %q from %q, as global matching ignore rule(s)\n",
						basename,
						dirname)
					continue
				}
			}

			// Test if the ignore rule based on path is respected
			if len(d.spec.Ignore) > 0 {
				if d.spec.Ignore.isMatchingRule(
					d.rootDir,
					relativeFoundDockerfile,
					image,
					platform,
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
					image,
					platform) {

					logrus.Debugf("Ignoring Dockerfile %q from %q, as not matching only rule(s)\n",
						basename,
						dirname)
					continue
				}
			}

			sourceSpec := dockerimage.NewDockerImageSpecFromImage(image, tag, d.spec.Auths)

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
				versionFilterPattern, err = d.versionFilter.GreaterThanPattern(tag)
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
				ImageName:            image,
				ImageTag:             tag,
				ScmID:                d.scmID,
				SourceID:             image,
				TargetID:             image,
				TargetFile:           relativeFoundDockerfile,
				TargetKeyword:        targetInstruction,
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
