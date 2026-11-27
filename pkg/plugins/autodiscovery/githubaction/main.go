package githubaction

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/core/log"
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

When the GitHub Actions lockfile ".github/workflows/actions.lock" lists a workflow, the manifests
updating the actions of that workflow also run "gh actions-lock" to update the lockfile. The command
uses the GitHub token Updatecli resolves, so "UPDATECLI_GITHUB_TOKEN" and the GitHub App settings take
precedence over the gh login. When the "gh actions-lock" extension isn't installed, the actions of
locked workflows are skipped, since bumping them without updating the lockfile would stop the workflow
from running.
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
	//     ".github". For Gitea and Forgejo, the crawler also searches workflow subdirectories
	//     when "rootdir" is unset or relative to the checkout.
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
	//   * a semver reference such as "v4.2.1" is never updated to a lower version, unless the
	//     "versionfilter" kind isn't semver, such as "latest". While every newer release is
	//     too recent, the update is skipped. Set "allowdowngrade" to true to fall back to the
	//     newest release old enough instead.
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
	Age age.AutodiscoverySpec `yaml:",omitempty"`
	// "credentials" defines the credentials used to authenticate with each git provider, keyed by git provider domain.
	//
	// remark:
	//   * without an entry, "gitea.com", "codeberg.org", "code.forgejo.org" and "data.forgejo.org" use kind "gitea",
	//     and any other domain uses kind "github" with the "github.com" credentials.
	//
	// example:
	//   ```
	//   autodiscovery:
	//     crawlers:
	//       github/action:
	//         credentials:
	//           "code.forgejo.org":
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
	// "registrymappings" maps an image name prefix, as written in the workflow files,
	// to the image name prefix where Updatecli looks up new tags and digests.
	//
	// default:
	//   empty, new versions are looked up under the image name written in the files.
	//
	// remark:
	//   * it applies to Docker images, such as "docker://" step images and job container images,
	//     not to actions.
	//   * use it for images pulled through a mirror, so new versions are looked up upstream
	//     while the files keep the mirror name.
	//   * a prefix matches whole path segments only: "my-registry.com/docker-io" matches
	//     "my-registry.com/docker-io/nginx" but not "my-registry.com/docker-iox/nginx".
	//   * when several prefixes match, the longest one wins.
	//   * a prefix can be a full image name, to map a single image, for instance
	//     when one mirror holds images from several upstream registries.
	//   * "credentialsdocker" credentials for the mirror registry are set on the condition.
	//   * "ignore" and "only" rules match the image name written in the files.
	//   * a condition checks the mirror already serves the new tag, or the new tag
	//     and digest with "digest" enabled, so the files are only updated once the
	//     mirror has caught up. With "digest" enabled, the mirror must serve the
	//     image unchanged for its digest to match the upstream one.
	//
	// example:
	//   ```
	//   registrymappings:
	//     "my-registry.com/docker-io/": "docker.io/"
	//     "my-registry.com/ghcr-io/": "ghcr.io/"
	//     # a single image mirrored under another name
	//     "my-registry.com/tools/updatecli": "ghcr.io/updatecli/updatecli"
	//   ```
	//
	RegistryMappings docker.RegistryMappings `yaml:",omitempty"`
	// "digest" defines whether the generated manifests pin the digest instead of the branch or tag.
	//
	// default:
	//   true
	//
	// remark:
	//   * digest pinning is supported for GitHub, Gitea and Forgejo actions, and Docker images.
	//   * when false, actions referenced by "main", "master" or "latest" are skipped.
	//   * when unset, the actions of a workflow listed in ".github/workflows/actions.lock" are not
	//     pinned, since the lockfile records their commit. An action already pinned stays pinned.
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
	// lockedWorkflows holds the workflows onboarded in the GitHub Actions lockfile, nil without lockfile
	lockedWorkflows map[string]bool
	// lockfileInvalid reports whether the GitHub Actions lockfile exists but can't be read
	lockfileInvalid bool
	// lockAvailable reports whether "gh actions-lock" can update the lockfile
	lockAvailable bool
	// lockEnvironments holds the environment variables of the "gh actions-lock" command
	lockEnvironments []lockEnvironment
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

	if err := s.RegistryMappings.Validate(); err != nil {
		return GitHubAction{}, err
	}

	dir := rootDir
	if path.IsAbs(s.RootDir) {
		if scmID != "" {
			log.Warningf("rootdir %q is an absolute path, scmID %q will be ignored", s.RootDir, scmID)
		}
		dir = s.RootDir
	}

	// If no RootDir have been provided via settings,
	// then fallback to the current process path.
	if len(dir) == 0 {
		log.Warningln("no working directory defined")
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
	log.Heading("GitHub Action")

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

	g.lockedWorkflows, err = lockedWorkflows(g.rootDir)
	if err != nil {
		// Without a readable lockfile, any workflow next to it may be locked, so none of their actions is bumped
		log.Warningf("skipping the actions of the workflows in %q: %s", path.Dir(actionsLockFile), err)
		g.lockfileInvalid = true
		err = nil
	}

	if len(g.lockedWorkflows) > 0 {
		g.lockAvailable, g.lockEnvironments = g.lockSetup()
	}

	manifests := g.discoverManifests()

	return manifests, err
}
