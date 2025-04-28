package nomad

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

var (
	// DefaultFileMatch specifies the default file shell pattern to identify Nomad files
	DefaultFilePattern []string = []string{"*.nomad", "*.hcl"}
)

// nomadDockerSpec is a struct that contains the information needed to
// generate a manifest for a Nomad job file
type nomadDockerSpec struct {
	File      string
	Value     string
	GroupName string
	TaskName  string
	JobName   string
	Path      string
}

// discoverDockerDriverManifests generates Updatecli manifests for Nomad job files
func (n Nomad) discoverDockerDriverManifests() ([][]byte, error) {
	var manifests [][]byte

	searchFromDir := n.rootDir
	// If the spec.RootDir is an absolute path, then it as already been set
	// correctly in the New function.
	if n.spec.RootDir != "" && !path.IsAbs(n.spec.RootDir) {
		searchFromDir = filepath.Join(n.rootDir, n.spec.RootDir)
	}

	foundNomadFiles, err := searchNomadFiles(searchFromDir, n.filematch)
	if err != nil {
		return nil, err
	}

	for _, foundNomadFile := range foundNomadFiles {
		relativeNomadFile, err := filepath.Rel(n.rootDir, foundNomadFile)
		logrus.Debugf("parsing file %q", foundNomadFile)
		if err != nil {
			// Let's try the next one if it fails
			logrus.Debugln(err)
			continue
		}

		dirname := filepath.Dir(relativeNomadFile)
		basename := filepath.Base(relativeNomadFile)

		nomadDockerSpecs, err := getNomadDockerSpecFromFile(foundNomadFile)
		if err != nil {
			logrus.Debugf("loading potential Nomad job spec from %q: %s", foundNomadFile, err)
			continue
		}

		if nomadDockerSpecs == nil {
			continue
		}

		if len(nomadDockerSpecs) == 0 {
			continue
		}

		for _, nomadDockerSpec := range nomadDockerSpecs {
			if nomadDockerSpec.Value == "" {
				continue
			}

			imageName, imageTag, imageDigest, err := dockerimage.ParseOCIReferenceInfo(nomadDockerSpec.Value)
			if err != nil {
				return nil, fmt.Errorf("parsing image %q: %s", nomadDockerSpec.Value, err)
			}

			/*
				For the time being, it's not possible to retrieve a list of tag for a specific digest
				without a significant amount f api call. More information on following issue
				https://github.com/google/go-containerregistry/issues/1297
				until a better solution, we don't handle docker image digest
			*/
			if imageDigest != "" && imageTag == "" {
				logrus.Debugf("docker digest without specified tag is not supported at the moment for %q", nomadDockerSpec.Value)
				continue
			}

			// Test if the ignore rule based on path is respected
			if len(n.spec.Ignore) > 0 {
				if n.spec.Ignore.isMatchingRule(
					n.rootDir,
					relativeNomadFile,
					nomadDockerSpec.JobName,
					nomadDockerSpec.Value,
				) {

					logrus.Debugf("Ignoring Nomad task %q from file %q from %q, as not matching ignore rule(s)\n",
						nomadDockerSpec.TaskName,
						basename,
						dirname)
					continue
				}
			}

			// Test if the only rule based on path is respected
			if len(n.spec.Only) > 0 {
				if !n.spec.Only.isMatchingRule(
					n.rootDir,
					relativeNomadFile,
					nomadDockerSpec.JobName,
					nomadDockerSpec.Value) {

					logrus.Debugf("Ignoring Nomad task %q from %q from %q, as not matching only rule(s)\n",
						nomadDockerSpec.TaskName,
						basename,
						dirname)
					continue
				}
			}

			sourceSpec := dockerimage.NewDockerImageSpecFromImage(imageName, imageTag, n.spec.Auths)

			versionFilterKind := n.versionFilter.Kind
			versionFilterPattern := n.versionFilter.Pattern
			tagFilter := "*"

			if sourceSpec != nil {
				versionFilterKind = sourceSpec.VersionFilter.Kind
				versionFilterPattern = sourceSpec.VersionFilter.Pattern
				tagFilter = sourceSpec.TagFilter
			}

			// If a versionfilter is specified in the manifest then we want to be sure that it takes precedence
			if !n.spec.VersionFilter.IsZero() {
				versionFilterKind = n.versionFilter.Kind
				versionFilterPattern, err = n.versionFilter.GreaterThanPattern(imageTag)
				tagFilter = ""
				if err != nil {
					logrus.Debugf("building version filter pattern: %s", err)
					sourceSpec.VersionFilter.Pattern = "*"
				}
			}

			var tmpl *template.Template
			if n.digest && sourceSpec != nil {
				tmpl, err = template.New("manifest").Parse(manifestTemplateDigestAndLatest)
				if err != nil {
					return nil, err
				}
			} else if n.digest && sourceSpec == nil {
				tmpl, err = template.New("manifest").Parse(manifestTemplateDigest)
				if err != nil {
					return nil, err
				}
			} else if !n.digest && sourceSpec != nil {
				tmpl, err = template.New("manifest").Parse(manifestTemplateLatest)
				if err != nil {
					return nil, err
				}
			} else {
				logrus.Infof("No source spec detected")
				return nil, nil
			}

			targetPrefix := ""
			if strings.HasPrefix(nomadDockerSpec.Path, "job.") {
				targetPrefix = imageName + ":"
			}

			params := struct {
				ActionID             string
				ImageName            string
				ImageTag             string
				SourceID             string
				TargetID             string
				TargetFile           string
				TargetPath           string
				TargetPrefix         string
				TagFilter            string
				VersionFilterKind    string
				VersionFilterPattern string
				ScmID                string
			}{
				ActionID:             n.actionID,
				ImageName:            imageName,
				ImageTag:             imageTag,
				SourceID:             "default",
				TargetID:             "default",
				TargetFile:           relativeNomadFile,
				TargetPath:           nomadDockerSpec.Path,
				TargetPrefix:         targetPrefix,
				TagFilter:            tagFilter,
				VersionFilterKind:    versionFilterKind,
				VersionFilterPattern: versionFilterPattern,
				ScmID:                n.scmID,
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
