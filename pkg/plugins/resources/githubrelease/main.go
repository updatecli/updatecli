package githubrelease

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

const (
	DeprecatedKeyTagHash = "hash"
	DeprecatedKeyTagName = "name"
	KeyTagName           = "tagname"
	KeyTagHash           = "taghash"
	KeyTitle             = "title"
)

// Spec defines a specification for a "gittag" resource
// parsed from an updatecli manifest file
type Spec struct {
	// owner defines repository owner to interact with.
	//
	// required: true
	//
	// compatible:
	//  * source
	//  * condition
	//
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// repository defines the repository name to interact with.
	//
	// required: true
	//
	// compatible:
	//  * source
	//  * condition
	//
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// token defines the GitHub personnal access token used to authenticate with.
	//
	// more information on https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens
	//
	// required: true
	//
	// compatible:
	//  * source
	//  * condition
	//
	Token string `yaml:",omitempty" jsonschema:"required"`
	// URL defines the default github url in case of GitHub enterprise.
	//
	// default: https://github.com
	//
	// compatible:
	//  * source
	//  * condition
	URL string `yaml:",omitempty"`
	// username defines the username used to authenticate with GitHub API.
	//
	// compatible:
	//  * source
	//  * condition
	Username string `yaml:",omitempty"`
	// versionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	//
	// default: latest
	//
	// compatible:
	//  * source
	//
	VersionFilter version.Filter `yaml:",omitempty"`
	// typeFilter specifies the GitHub Release type to retrieve before applying the versionfilter rule
	//
	// default:
	//  * draft: false
	//  * prerelease: false
	//  * release: true
	//  * latest: false
	//
	// compatible:
	//  * source
	// 	* condition
	//
	TypeFilter github.ReleaseType `yaml:",omitempty"`
	// tag allows to check for a specific release tag, release tag hash, or release title depending on a the parameter key.
	//
	// compatible:
	//   * condition
	//
	// default: source input
	//
	Tag string `yaml:",omitempty"`
	// "key" defines the GitHub release information we are looking for.
	// It accepts one of the following inputs:
	//    * "name": returns the "latest" tag name
	//    * "hash": returns the commit associated with the latest tag name
	//    * "title": returns the latest release title
	//
	// accepted values:
	//  * taghash
	//  * tagname
	//  * title
	//  * hash (deprecated)
	//  * name (deprecated)
	//
	// default: 'tagname'
	//
	// compatible:
	//   * source
	//   * condition
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
		spec:          newSpec,
		typeFilter:    newReleaseType,
		versionFilter: newFilter,
	}, nil
}
