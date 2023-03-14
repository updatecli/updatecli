package client

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/driver/azure"
	"github.com/drone/go-scm/scm/transport/oauth2"
	"github.com/sirupsen/logrus"
)

const (
	AZUREDOMAIN string = "dev.azure.com"
)

// Spec defines a specification for a "azure" resource
// parsed from an updatecli manifest file
type Spec struct {
	// [S][C][T] Owner defines the Azure DevOps owner]
	Owner string `yaml:",omitempty"`
	// [S][C][T] Project defines the Azure DevOps project]
	Project string `yaml:",omitempty"`
	// [S][C][T] URL specifies the default github url
	URL string `yaml:",omitempty"`
	// [S][C][T] Username specifies the username used to authenticate with Azure DevOps
	Username string `yaml:",omitempty"`
	// [S][C][T] Token specifies the credential used to authenticate with
	Token string `yaml:",omitempty"`
	// [S][C][T] RepoID specifies the azure devops repository ID
	RepoID string `yaml:",omitempty"`
}

type Client *scm.Client

func New(s Spec) (Client, error) {

	client, err := azure.New(s.URL, s.Owner, s.Project)

	if err != nil {
		return nil, err
	}

	client.Client = &http.Client{}

	if len(s.Token) >= 0 {
		client.Client = &http.Client{
			Transport: &oauth2.Transport{
				Source: oauth2.StaticTokenSource(
					&scm.Token{
						Token: s.Token,
					},
				),
			},
		}
	}

	return client, nil

}

func (s Spec) Validate() error {
	gotError := false
	missingParameters := []string{}

	if len(s.Owner) == 0 {
		gotError = true
		missingParameters = append(missingParameters, "owner")
	}

	if len(s.Project) == 0 {
		gotError = true
		missingParameters = append(missingParameters, "project")
	}

	if len(missingParameters) > 0 {
		logrus.Errorf("missing parameter(s) [%s]", strings.Join(missingParameters, ","))
	}

	if gotError {
		return fmt.Errorf("wrong azure configuration")
	}

	return nil
}

// Sanitize parse and update if needed a spec content
func (s *Spec) Sanitize() error {

	err := s.Validate()
	if err != nil {
		return err
	}

	if !strings.HasPrefix(s.URL, "https://") && !strings.HasPrefix(s.URL, "http://") {
		s.URL = "https://" + s.URL
	}

	return nil
}
