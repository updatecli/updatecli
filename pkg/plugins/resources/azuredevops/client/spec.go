package client

import (
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// Spec defines a specification for an Azure DevOps resource
// parsed from an updatecli manifest file.
type Spec struct {
	// Organization defines the Azure DevOps organization URL to interact with.
	Organization string `yaml:",omitempty"`
	// "url" defines the Azure DevOps organization URL to interact with.
	URL string `yaml:",omitempty"`
	// "project" defines the Azure DevOps project containing the repository.
	Project string `yaml:",omitempty" jsonschema:"required"`
	// "repository" defines the Azure DevOps repository name.
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// "username" defines the username used for git authentication.
	Username string `yaml:",omitempty"`
	// "token" specifies the personal access token used to authenticate with Azure DevOps.
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
