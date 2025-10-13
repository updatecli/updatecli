package githubaction

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"

	"github.com/updatecli/updatecli/pkg/plugins/scms/github/app"
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
	// files allows to specify the accepted GitHub Action workflow file name
	//
	// default:
	//   - ".github/workflows/*.yaml",
	//   - ".github/workflows/*.yml",
	//   - ".gitea/workflows/*.yaml",
	//   - ".gitea/workflows/*.yml",
	//   - ".forgejo/workflows/*.yaml",
	//   - ".forgejo/workflows/*.yml",
	Files []string `yaml:",omitempty"`
	// ignore allows to specify rule to ignore autodiscovery a specific GitHub action based on a rule
	//
	// default: empty
	//
	Ignore MatchingRules `yaml:",omitempty"`
	// only allows to specify rule to only autodiscover manifest for a specific GitHub action based on a rule
	//
	// default: empty
	//
	Only MatchingRules `yaml:",omitempty"`
	// rootDir allows to specify the root directory from where looking for GitHub Action
	//
	// default: empty
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
	// The ID of the credential must be the domain of the git provider to configure
	//
	// default: empty
	//
	// examples:
	// ```
	//   autodiscovery:
	//     crawlers:
	//       github/action:
	//         credentials:
	//           "code.forgejo.com":
	//             kind: gitea
	//             token: xxx
	//           "github.com":
	//             kind: github
	//             token: '{{ requiredEnv "GITHUB_TOKEN" }}'
	// ```
	Credentials map[string]gitProviderToken `yaml:",omitempty"`
	// CredentialsDocker provides a map of registry credentials where the key is the registry URL without scheme
	CredentialsDocker map[string]docker.InlineKeyChain `yaml:",omitempty"`
	// Digest provides parameters to specify if the generated manifest should use a digest instead of the branch or tag.
	//
	// Remark:
	// 	- The digest is only supported for GitHub Action and docker image tag update.
	//    Feel free to open an issue for the Gitea and Forgejo integration.
	Digest *bool `yaml:",omitempty"`
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
	// actionID hold the actionID used by the newly generated manifest
	actionID string
	// scmID hold the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	// workflowFiles is a list of HelmRelease files found
	workflowFiles []string
	// digest holds the value of the digest parameter
	digest bool
}

type gitProviderToken struct {
	// Kind defines the Kind of git provider to use
	//
	// accepted values: ['github','gitea','forgejo']
	Kind string `yaml:",omitempty"`
	// Token defines the Token to use to authenticate to the git provider
	//
	// The default value depends on the action domain
	// For 'github.com', the default value is set to first environment detected
	//  1. "UPDATECLI_GITHUB_TOKEN"
	//  2. "GITHUB_TOKEN"
	//
	// For 'gitea.com' and 'codeberg.org', the default value is set to first environment detected
	//  1. "UPDATECLI_GITHUB_TOKEN"
	//  1. "GITEA_TOKEN"
	Token string `yaml:",omitempty"`
	// App defines the GitHub App credentials used to authenticate with GitHub API.
	// It is not compatible with the "token" field.
	// It is recommended to use the GitHub App authentication method for better security and granular permissions.
	// For more information, please refer to the following documentation:
	// https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/authenticating-as-a-github-app-installation
	App *app.Spec `yaml:",omitempty"`
}

// New return a new valid Flux object.
func New(spec interface{}, rootDir, scmID, actionID string) (GitHubAction, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return GitHubAction{}, err
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
	digest := false
	if s.Digest != nil {
		digest = *s.Digest
	}

	return GitHubAction{
		actionID:      actionID,
		credentials:   s.Credentials,
		spec:          s,
		files:         files,
		rootDir:       dir,
		scmID:         scmID,
		versionFilter: newFilter,
		digest:        digest,
	}, nil

}

func (g GitHubAction) DiscoverManifests() ([][]byte, error) {
	logrus.Infof("\n\n%s\n", strings.ToTitle("GitHub Action"))
	logrus.Infof("%s\n", strings.Repeat("=", len("GitHub Action")+1))

	searchFromDir := g.rootDir
	// If the spec.RootDir is an absolute path, then it as already been set
	// correctly in the New function.
	if g.spec.RootDir != "" && !path.IsAbs(g.spec.RootDir) {
		searchFromDir = filepath.Join(g.rootDir, g.spec.RootDir)
	}

	err := g.searchWorkflowFiles(searchFromDir, g.files)
	if err != nil {
		return nil, err
	}

	manifests := g.discoverWorkflowManifests()

	return manifests, err
}
