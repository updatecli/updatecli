package pipeline

import (
	"path/filepath"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// scmDirectoryAttribute is the attribute every scm kind uses to name the local checkout
// directory: git, github, gitlab, gitea, stash, bitbucket and azuredevops all share it.
const scmDirectoryAttribute = "directory"

// resolveScmDirectory makes a relative scm "directory" resolve against baseDir instead of
// the process working directory.
//
// The spec is still the raw manifest content at this point, so working on the map covers
// every scm kind at once rather than repeating the same rewrite seven times.
func resolveScmDirectory(scmConfig *scm.Config, baseDir string) {
	if baseDir == "" {
		return
	}

	spec, ok := scmConfig.Spec.(map[string]interface{})
	if !ok {
		return
	}

	for key, value := range spec {
		if !strings.EqualFold(key, scmDirectoryAttribute) {
			continue
		}

		directory, ok := value.(string)
		if !ok || directory == "" || filepath.IsAbs(directory) {
			continue
		}

		spec[key] = filepath.Join(baseDir, directory)
	}
}
