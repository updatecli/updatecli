package client

import (
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// Spec defines the settings used to connect to an Azure DevOps repository.
type Spec struct {
	// "organization" defines the Azure DevOps organization name.
	//
	// remark:
	//   * it is required.
	//
	// example:
	//   * organization: updatecli
	//
	Organization string `yaml:",omitempty"`
	// "url" defines the Azure DevOps server URL.
	//
	// default:
	//   https://dev.azure.com
	//
	// remark:
	//   * "https://" is added when the URL has no scheme.
	//
	// example:
	//   * url: https://dev.azure.com
	//
	URL string `yaml:",omitempty"`
	// "project" defines the Azure DevOps project containing the repository.
	//
	// example:
	//   * project: updatecli
	//
	Project string `yaml:",omitempty" jsonschema:"required"`
	// "repository" defines the Azure DevOps repository name.
	//
	// example:
	//   * repository: website
	//
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// "username" defines the username used for git authentication.
	//
	// remark:
	//   * when unset, the environment variable "UPDATECLI_AZURE_DEVOPS_USERNAME" is used.
	//
	Username string `yaml:",omitempty"`
	// "token" defines the personal access token used to authenticate with Azure DevOps.
	//
	// remark:
	//   * when unset, the environment variable "UPDATECLI_AZURE_DEVOPS_TOKEN" is used.
	//
	Token string `yaml:",omitempty"`
}

// Validate validates that a spec contains the required Azure DevOps settings.
func (s Spec) Validate() error {
	missingParameters := []string{}

	if s.Organization == "" {
		missingParameters = append(missingParameters, "organization")
	}

	if len(missingParameters) > 0 {
		logrus.Errorf("missing parameter(s) [%s]", strings.Join(missingParameters, ","))
		return fmt.Errorf("wrong azure devops configuration")
	}

	return nil
}

// Sanitize normalizes a spec content.
func (s *Spec) Sanitize() error {
	if err := s.Validate(); err != nil {
		return err
	}

	s.URL = EnsureValidURL(s.URL)

	return nil
}

// EnsureValidURL normalizes an Azure DevOps organization URL.
func EnsureValidURL(rawURL string) string {
	if rawURL == "" {
		return DefaultAzureDevOpsURL
	}

	if !strings.HasPrefix(rawURL, "https://") && !strings.HasPrefix(rawURL, "http://") {
		rawURL = "https://" + rawURL
	}

	return strings.TrimRight(rawURL, "/")
}

// GitURL returns the repository git URL used by the SCM implementation.
func GitURL(baseURL, organization, project, repository string) string {
	u, err := url.Parse(EnsureValidURL(baseURL))
	if err != nil {
		return strings.TrimRight(EnsureValidURL(baseURL), "/") + "/" + project + "/_git/" + repository
	}

	u.Path = path.Join(
		u.Path,
		url.PathEscape(organization),
		url.PathEscape(project),
		"_git",
		url.PathEscape(repository),
	)

	return u.String()
}

// PullRequestURL returns the Azure DevOps web URL for a pull request.
func PullRequestURL(baseURL, organization, project, repository string, pullRequestID int) string {
	u, err := url.Parse(EnsureValidURL(baseURL))
	if err != nil {
		return strings.TrimRight(EnsureValidURL(baseURL), "/") +
			"/" + organization +
			"/" + project +
			"/_git/" + repository +
			"/pullrequest/" + strconv.Itoa(pullRequestID)
	}

	u.Path = path.Join(
		u.Path,
		url.PathEscape(organization),
		url.PathEscape(project),
		"_git",
		url.PathEscape(repository),
		"pullrequest",
		strconv.Itoa(pullRequestID),
	)

	return u.String()
}
