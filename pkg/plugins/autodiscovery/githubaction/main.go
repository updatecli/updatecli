package githubaction

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"

	"github.com/updatecli/updatecli/pkg/plugins/scms/github/app"
)

var (
	// defaultWorkflowFiles specifies accepted Action workflow file name
	defaultWorkflowFiles        []string = []string{"*.yaml", "*.yml"}
	defaultCompositeActionNames []string = []string{"*"}
	defaultVersionFilterPattern string   = "*"
	defaultVersionFilterKind    string   = "semver"
	kindGitea                   string   = "gitea"
	kindGitHub                  string   = "github"
	latestVersionIdentifier     string   = "latest"
	kindForgejo                 string   = "forgejo"
	defaultGitProviderURL       string   = "https://github.com"
)

/*
"github/action" defines the specification for the GitHub Action autodiscovery crawler.
It searches workflow files and composite actions, and generates manifests to update the actions and Docker images they use.
The "gitea/action" crawler uses the same specification.
*/
type Spec struct {
	// "files" defines the workflow file name patterns the crawler searches for.
	//
	// default:
	//   ```
	//   files:
	//     - "*.yaml"
	//     - "*.yml"
	//   ```
	//
	// remark:
	//   * the pattern is matched against the file name only, not against its path, so a
	//     pattern such as ".github/workflows/*.yaml" never matches.
	//   * a workflow file must sit directly inside a "workflows" directory whose parent is
	//     ".github", ".gitea", or ".forgejo".
	//
	Files []string `yaml:",omitempty"`
	// "actions" defines the composite action name patterns the crawler searches for.
	//
	// default:
	//   ```
	//   actions:
	//     - "*"
	//   ```
	//
	// remark:
	//   * a composite action is identified by an "action.yaml" or "action.yml" file, and the
	//     pattern is matched against the name of the directory holding it.
	//
	Actions []string `yaml:",omitempty"`

	// "ignore" defines rules to exclude matching actions or Docker images from the autodiscovery.
	//
	// remark:
	//   * an action or Docker image is ignored when it matches at least one rule.
	//
	Ignore MatchingRules `yaml:",omitempty"`
	// "only" defines rules to restrict the autodiscovery to matching actions or Docker images.
	//
	// remark:
	//   * an action or Docker image is kept only when it matches at least one rule.
	//
	Only MatchingRules `yaml:",omitempty"`
	// "rootdir" defines the directory where the crawler starts searching for workflow files and composite actions.
	//
	// default:
	//   the scm directory when "scmid" is set, otherwise the directory relative paths resolve from, by default the working directory.
	//
	// remark:
	//   * a relative path is resolved from the default directory.
	//   * an absolute path is used as is, instead of the scm directory.
	//
	RootDir string `yaml:",omitempty"`
	// "versionfilter" defines the version filter used by the generated manifests.
	//
	// default:
	//   * for an action, kind "semver" with pattern "*", the latest version, when its reference is a semantic version, otherwise kind "latest".
	//   * for a Docker image, kind "semver" with pattern ">=<current tag>", combined with a tag filter derived from the current tag.
	//
	// remark:
	//   * with kind "semver", "pattern" accepts:
	//     * "prerelease": the latest prerelease of the current version.
	//     * "patch": patch updates only.
	//     * "minor": patch and minor updates.
	//     * "minoronly": minor updates only.
	//     * "major": patch, minor and major updates.
	//     * "majoronly": major updates only.
	//     * a version constraint, such as ">= 1.0.0".
	//   * with kind "regex", "pattern" accepts a regular expression.
	//   * more examples at https://www.updatecli.io/docs/core/versionfilter/
	//
	// example:
	//   ```
	//   versionfilter:
	//     kind: semver
	//     pattern: minor
	//   ```
	//
	VersionFilter version.Filter `yaml:",omitempty"`
	// "age" defines the minimum or maximum age of a release, tag, or branch to be considered valid.
	//
	// It is the "dependency cooldown" setting: setting "minimum" keeps Updatecli from
	// suggesting a version that has just been published.
	//
	// default:
	//   empty, no age filtering.
	//
	// remark:
	//   * it accepts a duration string, such as "24h", "7d", "3w" or "1y".
	//   * the age filter is not applied to the Docker images referenced by a workflow.
	//
	// example:
	//   ```
	//   autodiscovery:
	//     crawlers:
	//       github/action:
	//         age:
	//           minimum: '7d'
	//   ```
	//
	Age age.Spec `yaml:",omitempty"`
	// "credentials" defines the credentials used to authenticate with each git provider, keyed by git provider domain.
	//
	// remark:
	//   * without an entry, "gitea.com", "codeberg.org" and "code.forgejo.org" use kind "gitea",
	//     and any other domain uses kind "github" with the "github.com" credentials.
	//
	// example:
	//   ```
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
	//   ```
	//
	Credentials map[string]gitProviderToken `yaml:",omitempty"`
	// "credentialsdocker" defines the registry credentials used for Docker images, keyed by registry host without scheme.
	//
	// remark:
	//   * when empty, Updatecli uses the local OCI credentials, such as the Docker ones.
	//
	// example:
	//   ```
	//   credentialsdocker:
	//     "ghcr.io":
	//       token: "xxx"
	//     "index.docker.io":
	//       username: "admin"
	//       password: "password"
	//   ```
	//
	CredentialsDocker map[string]docker.InlineKeyChain `yaml:",omitempty"`
	// "digest" defines whether the generated manifests pin the digest instead of the branch or tag.
	//
	// default:
	//   true
	//
	// remark:
	//   * digest pinning is supported for GitHub actions and Docker images, not yet for Gitea and Forgejo actions.
	//   * when false, actions referenced by "main", "master" or "latest" are skipped.
	//
	Digest *bool `yaml:",omitempty"`
}

// GitHubAction holds all information needed to generate GitHubAction manifest.
type GitHubAction struct {
	// credentials defines the credentials to use to authenticate to the git provider
	credentials map[string]gitProviderToken
	// files defines the accepted Action workflow file name
	files []string
	// actions defines the accepted Composite Action name
	actions []string
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
	// workflowFiles is a list of workflow files found
	workflowFiles []string
	// compositeActionFiles is a list of Composite action files found
	compositeActionFiles []string
	// digest holds the value of the digest parameter
	digest bool
}

// gitProviderToken defines the credentials used to authenticate with a git provider.
type gitProviderToken struct {
	// "kind" defines the kind of git provider.
	//
	// remark:
	//   * accepted values are "github", "gitea" and "forgejo".
	//
	Kind string `yaml:",omitempty"`
	// "token" defines the token used to authenticate with the git provider.
	//
	// default:
	//   * for kind "github", the "GITHUB_TOKEN" environment variable.
	//   * for kind "gitea" or "forgejo", the first environment variable set among
	//     "UPDATECLI_GITEA_TOKEN" and "GITEA_TOKEN".
	//
	// remark:
	//   * for kind "github", the "UPDATECLI_GITHUB_TOKEN" environment variable and the GitHub App
	//     environment variables take precedence over this setting.
	//
	Token string `yaml:",omitempty"`
	// "app" defines the GitHub App credentials used to authenticate with the GitHub API.
	//
	// remark:
	//   * "token" takes precedence over "app" when both are set.
	//   * a GitHub App is recommended for better security and more granular permissions.
	//   * see https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/authenticating-as-a-github-app-installation
	//
	App *app.Spec `yaml:",omitempty"`
}

// New return a new valid Flux object.
func New(spec interface{}, rootDir, scmID, actionID string) (GitHubAction, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return GitHubAction{}, err
	}

	// Validate ignore rules
	if err := s.Ignore.Validate(); err != nil {
		return GitHubAction{}, fmt.Errorf("invalid ignore spec: %w", err)
	}

	// Validate only rules
	if err := s.Only.Validate(); err != nil {
		return GitHubAction{}, fmt.Errorf("invalid only spec: %w", err)
	}

	if err := s.Age.Validate(); err != nil {
		return GitHubAction{}, fmt.Errorf("invalid age spec: %w", err)
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
		logrus.Warningln("no working directory defined")
		dir, err = filepath.Abs(".")
		if err != nil {
			return GitHubAction{}, err
		}
	}

	files := defaultWorkflowFiles
	if len(s.Files) > 0 {
		files = s.Files
	}

	actions := defaultCompositeActionNames
	if len(s.Actions) > 0 {
		actions = s.Actions
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		// By default, helm versioning uses semantic versioning.
		newFilter.Kind = defaultVersionFilterKind
		newFilter.Pattern = defaultVersionFilterPattern
	}
	digest := true
	if s.Digest != nil {
		digest = *s.Digest
	}

	return GitHubAction{
		actionID:      actionID,
		credentials:   s.Credentials,
		spec:          s,
		files:         files,
		actions:       actions,
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

	err := g.searchWorkflowFiles(searchFromDir)
	if err != nil {
		return nil, err
	}

	err = g.searchCompositeActionFiles(searchFromDir)
	if err != nil {
		return nil, err
	}

	manifests := g.discoverManifests()

	return manifests, err
}
