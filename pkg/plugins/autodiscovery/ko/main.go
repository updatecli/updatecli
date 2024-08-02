package ko

import (
	"path"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines the parameters which can be provided to the Kubernetes builder.
type Spec struct {
	// Auths provides a map of registry credentials where the key is the registry URL without scheme
	Auths map[string]docker.InlineKeyChain `yaml:",omitempty"`
	/*
		digest provides parameters to specify if the generated manifest should use a digest on top of the tag.
	*/
	Digest *bool `yaml:",omitempty"`
	/* Files allows to specify a list of Files to analyze.

	    The pattern syntax is:
	       pattern:
	         { term }
	       term:
	         '*'         matches any sequence of non-Separator characters
	         '?'         matches any single non-Separator character
	         '[' [ '^' ] { character-range } ']' character class (must be non-empty)
	         c           matches character c (c != '*', '?', '\\', '[')
	         '\\' c      matches character c

		    character-range:
		    	c           matches character c (c != '\\', '-', ']')
	         '\\' c      matches character c
	         lo '-' hi   matches character c for lo <= c <= hi

	        Match requires pattern to match all of name, not just a substring.
	        The only possible returned error is ErrBadPattern, when pattern
	        is malformed.

	        On Windows, escaping is disabled. Instead, '\\' is treated as
	        path separator.
	*/
	Files []string `yaml:",omitempty"`
	// RootDir defines the root directory used to recursively search for Kubernetes files
	RootDir string `yaml:",omitempty"`
	// Ignore allows to specify rule to ignore autodiscovery a specific Kubernetes manifest based on a rule
	Ignore MatchingRules `yaml:",omitempty"`
	// Only allows to specify rule to only autodiscover manifest for a specific Kubernetes manifest based on a rule
	Only MatchingRules `yaml:",omitempty"`
	/*
		versionfilter provides parameters to specify the version pattern used when generating manifest.

		kind - semver
			versionfilter of kind `semver` uses semantic versioning as version filtering
			pattern accepts one of:
				`patch` - patch only update patch version
				`minor` - minor only update minor version
				`major` - major only update major versions
				`a version constraint` such as `>= 1.0.0`

		kind - regex
			versionfilter of kind `regex` uses regular expression as version filtering
			pattern accepts a valid regular expression

		example:
		```
			versionfilter:
				kind: semver
				pattern: minor
		```

		and its type like regex, semver, or just latest.
	*/
	VersionFilter version.Filter `yaml:",omitempty"`
}

// Ko holds all information needed to generate Ko manifests.
type Ko struct {
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for Kubernetes manifests
	rootDir string
	// scmID hold the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	// files holds the list of files to analyze
	files []string
	// digest holds the value of the digest parameter
	digest bool
}

// New return a new valid Ko object.
func New(spec interface{}, rootDir, scmID string) (Ko, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Ko{}, err
	}

	dir := rootDir
	if path.IsAbs(s.RootDir) {
		if scmID != "" {
			logrus.Warningf("rootdir %q is an absolute path, scmID %q will be ignored", s.RootDir, scmID)
		}
		dir = s.RootDir
	}

	// If no RootDir have been provided via settings,
	// then fallback to the current process path.
	if len(dir) == 0 {
		logrus.Errorln("no working directory defined")
		return Ko{}, err
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		// By default, helm versioning uses semantic versioning. Containers is not but...
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	files := DefaultKoFiles
	if len(s.Files) > 0 {
		files = s.Files
	}

	digest := true
	if s.Digest != nil {
		digest = *s.Digest
	}

	return Ko{
		digest:        digest,
		spec:          s,
		rootDir:       dir,
		files:         files,
		scmID:         scmID,
		versionFilter: newFilter,
	}, nil

}

func (f Ko) DiscoverManifests() ([][]byte, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Ko"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Ko")+1))

	return f.discoverContainerManifests()
}
