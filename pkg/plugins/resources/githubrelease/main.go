package githubrelease

import (
	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for a "gittag" resource
// parsed from an updatecli manifest file
type Spec struct {
	// [s][c] Owner specifies repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// [s][c] Repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// [s][c] Token specifies the credential used to authenticate with
	Token string `yaml:",omitempty" jsonschema:"required"`
	// [s][c] URL specifies the default github url in case of GitHub enterprise
	URL string `yaml:",omitempty"`
	// [s][c] Username specifies the username used to authenticate with Github API
	Username string `yaml:",omitempty" jsonschema:"required"`
	// [s][c] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// [s][c] Type specifies the Github Release type to interact with
	Type github.ReleaseType `yaml:",omitempty"`
	// [c] Tag allows to check for a specific release tag
	Tag string `yaml:",omitempty"`
}

// GitHubRelease defines a resource of kind "githubrelease"
type GitHubRelease struct {
	ghHandler     github.GithubHandler
	versionFilter version.Filter // Holds the "valid" version.filter, that might be different than the user-specified filter (Spec.VersionFilter)
	foundVersion  version.Version
	spec          Spec
	releaseType   github.ReleaseType
}

var (
	// deprecationTagSearchMessage is display if ReleaseType is specified
	deprecationTagSearchMessage string = `Deprecation announcement, githubrelease will soon stop fallback to git tags if no release could be found.
This behavior is not compatible with release type filtering.
If you need to manipulate Git tags, please use the gittag resource.
You can dismiss this warning by adding a release type filter rule such as

>  spec:
>    owner: updatecli
>    repository: updatecli
>    token: '{{ requiredEnv "UPDATECLI_GITHUB_TOKEN" }}'
>    username: '{{ requiredEnv "UPDATECLI_GITHUB_ACTOR" }}'
>    type:
>      draft: true

`
)

// New returns a new valid GitHubRelease object.
func New(spec interface{}) (*GitHubRelease, error) {
	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return &GitHubRelease{}, err
	}

	newHandler, err := github.New(github.Spec{
		Owner:      newSpec.Owner,
		Repository: newSpec.Repository,
		Token:      newSpec.Token,
		URL:        newSpec.URL,
		Username:   newSpec.Username,
	}, "")
	if err != nil {
		return &GitHubRelease{}, err
	}

	newFilter, err := newSpec.VersionFilter.Init()
	if err != nil {
		return &GitHubRelease{}, err
	}

	newReleaseType := newSpec.Type
	newReleaseType.Init()

	return &GitHubRelease{
		ghHandler:     newHandler,
		versionFilter: newFilter,
		releaseType:   newReleaseType,
		spec:          newSpec,
	}, nil
}
