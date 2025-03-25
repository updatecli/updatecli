package githubrelease

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

const (
	KeyName  = "name"
	KeyHash  = "hash"
	KeyTitle = "title"
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
	// [s][c] Username specifies the username used to authenticate with GitHub API
	Username string `yaml:",omitempty"`
	// [s] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
	// [s][c] TypeFilter specifies the GitHub Release type to retrieve before applying the versionfilter rule
	TypeFilter github.ReleaseType `yaml:",omitempty"`
	// [c] Tag allows to check for a specific release tag, default to source output
	Tag string `yaml:",omitempty"`
	//  "key" of the tag object to retrieve.
	//
	//  Accepted values: ['name','hash','title'].
	//
	//  Default: 'name'
	//  Compatible:
	//    * source
	//    * condition
	Key string `yaml:",omitempty"`
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
	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return &GitHubRelease{}, err
	}

	validationErrors := []string{}
	if newSpec.Key != "" && newSpec.Key != KeyHash && newSpec.Key != KeyName && newSpec.Key != KeyTitle {
		validationErrors = append(validationErrors, "The only valid values for Key are 'name', 'hash', 'title', or empty.")
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
		versionFilter: newFilter,
		typeFilter:    newReleaseType,
		spec:          newSpec,
	}, nil
}
