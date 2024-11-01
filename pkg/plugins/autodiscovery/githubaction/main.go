package githubaction

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

var (
	// defaultWorkflowFiles specifies accepted GitHub Action workflow file name
	defaultWorkflowFiles        []string = []string{"*.yaml", "*.yml"}
	defaultVersionFilterPattern string   = "*"
	defaultVersionFilterKind    string   = "semver"
	kindGitea                   string   = "gitea"
	kindGitHub                  string   = "github"
	kindForgejo                 string   = "forgejo"
	defaultGitProviderURL       string   = "https://github.com"
)

// Spec defines the parameters which can be provided to the Github Action crawler.
type Spec struct {
	Files []string `yaml:",omitempty"`
	// ignore allows to specify rule to ignore autodiscovery a specific Flux helmrelease based on a rule
	//
	// default: empty
	//
	Ignore MatchingRules `yaml:",omitempty"`
	// only allows to specify rule to only autodiscover manifest for a specific Flux helm release based on a rule
	//
	// default: empty
	//
	Only MatchingRules `yaml:",omitempty"`
	// OCIRepository allows to specify if OCI repository files should be updated
	//
	// default: true
	RootDir string `yaml:",omitempty"`
	// versionfilter provides parameters to specify the version pattern used when generating manifest.
	//
	// kind - semver
	//		versionfilter of kind `semver` uses semantic versioning as version filtering
	//		pattern accepts one of:
	//			`patch` - patch only update patch version
	//			`minor` - minor only update minor version
	//			`major` - major only update major versions
	//			`a version constraint` such as `>= 1.0.0`
	//
	//	kind - regex
	//		versionfilter of kind `regex` uses regular expression as version filtering
	//		pattern accepts a valid regular expression
	//
	//	example:
	//	```
	//		versionfilter:
	//			kind: semver
	//			pattern: minor
	//	```
	//
	//	and its type like regex, semver, or just latest.
	//
	VersionFilter version.Filter `yaml:",omitempty"`
	// Credentials allows to specify the credentials to use to authenticate to the git provider
	Credentials map[string]gitProviderToken `yaml:",omitempty"`
}

// GitHubAction holds all information needed to generate GitHubAction manifest.
type GitHubAction struct {
	// credentials defines the credentials to use to authenticate to the git provider
	credentials map[string]gitProviderToken
	// files defines the accepted GitHub Action workflow file name
	files []string
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the  oot directory from where looking for Flux
	rootDir string
	// scmID hold the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	// workflowFiles is a list of HelmRelease files found
	workflowFiles []string
	// token allows to specify a GitHub token to use for GitHub API requests
	token string
}

type gitProviderToken struct {
	// Kind defines the Kind of git provider to use
	//
	// accepted values: ['github','gitea','forgejo']
	Kind string `yaml:",omitempty"`
	// Token defines the Token to use to authenticate to the git provider
	//
	// The default value depeneds on the action domain
	// For 'github.com', the default value is set to first environment detected
	//  1. "UPDATECLI_GITHUB_TOKEN"
	//  2. "GITHUB_TOKEN"
	//
	// For 'gitea.com' and 'codeberg.org', the default value is set to first environment detected
	//  1. "GITEA_TOKEN"
	Token string `yaml:",omitempty"`
}

// New return a new valid Flux object.
func New(spec interface{}, rootDir, scmID string) (GitHubAction, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return GitHubAction{}, err
	}

	dir := rootDir
	if len(s.RootDir) > 0 {
		dir = s.RootDir
	}

	// If no RootDir have been provided via settings,
	// then fallback to the current process path.
	if len(dir) == 0 {
		logrus.Errorln("no working directory defined")
		return GitHubAction{}, err
	}

	files := defaultWorkflowFiles
	if len(s.Files) > 0 {
		files = s.Files
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		// By default, helm versioning uses semantic versioning.
		newFilter.Kind = defaultVersionFilterKind
		newFilter.Pattern = defaultVersionFilterPattern
	}

	return GitHubAction{
		credentials:   s.Credentials,
		spec:          s,
		files:         files,
		rootDir:       dir,
		scmID:         scmID,
		versionFilter: newFilter,
	}, nil

}

func (g GitHubAction) DiscoverManifests() ([][]byte, error) {
	logrus.Infof("\n\n%s\n", strings.ToTitle("GitHub Action"))
	logrus.Infof("%s\n", strings.Repeat("=", len("GitHub Action")+1))

	err := g.searchWorkflowFiles(g.rootDir, g.files)
	if err != nil {
		return nil, err
	}

	manifests := g.discoverWorkflowManifests()

	return manifests, err
}
