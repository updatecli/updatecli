package githubrelease

import (
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/app"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

const (
	DeprecatedKeyTagHash = "hash"
	DeprecatedKeyTagName = "name"
	KeyTagName           = "tagname"
	KeyTagHash           = "taghash"
	KeyTitle             = "title"
)

/*
"githubrelease" defines the specification for retrieving and checking GitHub releases.
It can be used as a "source" or a "condition".
*/
type Spec struct {
	// "owner" defines the owner of the GitHub repository.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * "owner" is required.
	//
	// example:
	//   * owner: updatecli
	//
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// "repository" defines the name of the GitHub repository.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * "repository" is required.
	//
	// example:
	//   * repository: updatecli
	//
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// "token" defines the GitHub personal access token used to authenticate with the GitHub API.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * the environment variable "UPDATECLI_GITHUB_TOKEN" takes precedence over "token".
	//   * without "token" or "app", the environment variable "GITHUB_TOKEN" is used.
	//   * without any credential, Updatecli sends unauthenticated requests, so operations
	//     requiring authentication fail.
	//   * more information on https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens
	//
	Token string `yaml:",omitempty"`
	// "url" defines the GitHub URL, for GitHub Enterprise.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   https://github.com
	//
	// remark:
	//   * "https://" is added when the URL has no scheme.
	//
	URL string `yaml:",omitempty"`
	// "username" defines the username used to authenticate with the GitHub API.
	//
	// compatible:
	//   * source
	//   * condition
	//
	Username string `yaml:",omitempty"`
	// "versionfilter" defines the version pattern and its type, such as regex, semver, or latest.
	//
	// compatible:
	//   * source
	//
	// default:
	//   latest
	//
	VersionFilter version.Filter `yaml:",omitempty"`
	// "age" defines the minimum or maximum age of a release to be considered valid.
	//
	// compatible:
	//   * source
	//
	// remark:
	//   * "minimum" and "maximum" accept a duration string such as "24h", "7d", "3w" or "1y".
	//   * when the age filter discards every release, the source is skipped.
	//   * the age filter cannot be applied to the git tag fallback used when a repository
	//     does not publish any GitHub release.
	//
	// example:
	// ```
	//   age:
	//     minimum: 7d
	// ```
	//
	Age age.Spec `yaml:",omitempty"`
	// "typefilter" defines the GitHub release types to retrieve before applying "versionfilter".
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   * draft: false
	//   * prerelease: false
	//   * release: true
	//   * latest: false
	//
	// remark:
	//   * when "draft", "prerelease" and "release" are all false, "release" is set to true.
	//   * when "typefilter" is unset and the repository has no release, Updatecli falls back to git tags.
	//
	TypeFilter github.ReleaseType `yaml:",omitempty"`
	// "tag" defines the release tag name, tag hash, or release title to check, depending on "key".
	//
	// compatible:
	//   * condition
	//
	// default:
	//   the output of the associated source.
	//
	Tag string `yaml:",omitempty"`
	// "key" defines which release information Updatecli looks for.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   tagname
	//
	// remark:
	//   * accepted values are:
	//     * "tagname": the release tag name
	//     * "taghash": the commit hash of the release tag
	//     * "title": the release title
	//   * "name" is a deprecated alias of "tagname".
	//   * "hash" is a deprecated alias of "taghash".
	//
	// example:
	//   * key: taghash
	//
	Key string `yaml:",omitempty"`
	// "app" defines the GitHub App credentials used to authenticate with the GitHub API.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * "app" is not compatible with "token" and "username".
	//   * a GitHub App is the recommended authentication method, for better security and granular permissions.
	//   * more information on https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/authenticating-as-a-github-app-installation
	//
	App *app.Spec `yaml:",omitempty"`
}

// GitHubRelease defines a resource of kind "githubrelease"
type GitHubRelease struct {
	ghHandler     github.GithubHandler
	versionFilter version.Filter // Holds the "valid" version.filter, that might be different than the user-specified filter (Spec.VersionFilter)
	foundVersion  version.Version
	spec          Spec
	typeFilter    github.ReleaseType
}

// New returns a new valid GitHubRelease object.
func New(spec interface{}) (*GitHubRelease, error) {
	validationErrors := []string{}

	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return &GitHubRelease{}, err
	}

	switch newSpec.Key {
	case "":
		newSpec.Key = KeyTagName
		logrus.Debugf("configuration \"key\" not set, defaulting to %q", KeyTagName)
	case KeyTagHash, KeyTagName, KeyTitle:
		// Nothing to do
	case DeprecatedKeyTagName:
		logrus.Warningf("configuration \"key\" set to %q is deprecated and should be replaced by %q", DeprecatedKeyTagName, KeyTagName)
		newSpec.Key = KeyTagName
	case DeprecatedKeyTagHash:
		logrus.Warningf("configuration \"key\" set to %q is deprecated and should be replaced by %q", DeprecatedKeyTagHash, KeyTagHash)
		newSpec.Key = KeyTagHash
	default:
		validationErrors = append(
			validationErrors,
			fmt.Sprintf(
				"Value %q detected for key \"key\", accepted values for Key are 'name', %q, %q, %q, or empty.",
				newSpec.Key, KeyTagName, KeyTagHash, KeyTitle,
			),
		)

	}

	if err := newSpec.Age.Validate(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Return all the validation errors if found any
	if len(validationErrors) > 0 {
		return &GitHubRelease{}, fmt.Errorf("validation error: the provided manifest configuration has the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	newHandler, err := github.New(github.Spec{
		Owner:      newSpec.Owner,
		Repository: newSpec.Repository,
		Token:      newSpec.Token,
		URL:        newSpec.URL,
		Username:   newSpec.Username,
		App:        newSpec.App,
	}, "")
	if err != nil {
		return &GitHubRelease{}, err
	}

	newFilter, err := newSpec.VersionFilter.Init()
	if err != nil {
		return &GitHubRelease{}, err
	}

	newReleaseType := newSpec.TypeFilter
	newReleaseType.Init()

	return &GitHubRelease{
		ghHandler:     newHandler,
		spec:          newSpec,
		typeFilter:    newReleaseType,
		versionFilter: newFilter,
	}, nil
}

// ReportConfig returns a new configuration object with only the necessary fields
// to identify the resource without any sensitive information
// and context specific data.
func (g *GitHubRelease) ReportConfig() interface{} {
	return Spec{
		Owner:         g.spec.Owner,
		Repository:    g.spec.Repository,
		VersionFilter: g.spec.VersionFilter,
		Age:           g.spec.Age,
		TypeFilter:    g.spec.TypeFilter,
		URL:           redact.URL(g.spec.URL),
		Tag:           g.spec.Tag,
		Key:           g.spec.Key,
	}
}
