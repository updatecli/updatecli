package pipeline

import (
	"maps"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"go.yaml.in/yaml/v3"
)

// scmDirectoryAttribute is the attribute every scm kind uses to name the local checkout
// directory: git, github, gitlab, gitea, stash, bitbucket and azuredevops all share it.
const scmDirectoryAttribute = "directory"

// resolveScmDirectory makes a relative scm "directory" resolve against baseDir instead of
// the process working directory.
//
// The spec is still the raw manifest content at this point, so working on the map covers
// every scm kind at once. The manifest keeps its raw values, so anything reading them later,
// such as autodiscovery, resolves them once.
func resolveScmDirectory(scmConfig *scm.Config, baseDir string) {
	if baseDir == "" {
		return
	}

	updateScmDirectory(scmConfig, func(directory string) string {
		return filepath.Join(baseDir, directory)
	})
}

// newScm builds the scm handler of a pipeline from its manifest configuration, with a
// relative "directory" resolved against baseDir. Init and Update both rebuild the scms, so
// they share it to always resolve the directory the same way.
//
// scmConfig is received by value: the handler keeps a pointer to its own copy.
func newScm(scmConfig scm.Config, baseDir, pipelineID string) (scm.Scm, error) {
	resolveScmDirectory(&scmConfig, baseDir)

	return scm.New(&scmConfig, pipelineID)
}

// PinScmDirectory returns a copy of scmConfig whose relative "directory" is made absolute
// against baseDir, or against the process working directory when baseDir is empty.
//
// Autodiscovery uses it for the scms it copies into the manifests it generates. Those
// manifests have their own base directory, and an absolute directory is never resolved
// again.
func PinScmDirectory(scmConfig scm.Config, baseDir string) scm.Config {
	updateScmDirectory(&scmConfig, func(directory string) string {
		absDirectory, err := filepath.Abs(filepath.Join(baseDir, directory))
		if err != nil {
			logrus.Debugf("unable to make scm directory %q absolute: %s", directory, err)
			return filepath.Join(baseDir, directory)
		}
		return absDirectory
	})

	return scmConfig
}

// updateScmDirectory rewrites a relative, non empty scm "directory" with resolve.
// It works on a copy of the spec map, which it then assigns to scmConfig.
//
// The scms discovered by githubsearch, gitlabsearch or azuredevopssearch hold a typed spec.
// Config.Update turns it into a map through a YAML round trip before every resource runs,
// so it is converted the same way here to resolve its directory like any other scm. A
// typed spec is only replaced when its directory is rewritten.
func updateScmDirectory(scmConfig *scm.Config, resolve func(directory string) string) {
	spec, isMap := scmConfig.Spec.(map[string]interface{})
	if isMap {
		spec = maps.Clone(spec)
	} else {
		var ok bool
		if spec, ok = specToMap(scmConfig.Spec); !ok {
			return
		}
	}

	updated := false
	for key, value := range spec {
		if !strings.EqualFold(key, scmDirectoryAttribute) {
			continue
		}

		directory, ok := value.(string)
		if !ok || directory == "" || filepath.IsAbs(directory) {
			continue
		}

		spec[key] = resolve(directory)
		updated = true
	}

	if isMap || updated {
		scmConfig.Spec = spec
	}
}

// specToMap converts a typed scm spec into the map Config.Update would turn it into.
func specToMap(spec interface{}) (map[string]interface{}, bool) {
	if spec == nil {
		return nil, false
	}

	content, err := yaml.Marshal(spec)
	if err != nil {
		logrus.Debugf("unable to read the scm spec directory: %s", err)
		return nil, false
	}

	specMap := map[string]interface{}{}
	if err := yaml.Unmarshal(content, &specMap); err != nil {
		logrus.Debugf("unable to read the scm spec directory: %s", err)
		return nil, false
	}

	return specMap, true
}
