package client

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/driver/gitea"
	"github.com/drone/go-scm/scm/transport/oauth2"
	"github.com/sirupsen/logrus"
)

// Spec defines a specification for a "gitea" resource
// parsed from an updatecli manifest file
type Spec struct {
	// [S][C][T] URL specifies the default github url in case of Gitea enterprise
	URL string `yaml:",omitempty" jsonschema:"required"`
	// Username specifies the username used to authenticate with Gitea API
	Username string `yaml:",omitempty"`
	// [S][C][T] Token specifies the credential used to authenticate with
	Token string `yaml:",omitempty"`
}

type Client *scm.Client

func New(s Spec) (Client, error) {

	client, err := gitea.New(s.URL)

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

// Validate validates that a spec contains good content
func (s Spec) Validate() (err error) {

	if len(s.URL) == 0 {
		logrus.Errorf("missing %q parameter", "url")
		return fmt.Errorf("wrong configuration")
	}

	return nil
}

// Sanitize parse and update if needed a spec content
func (s *Spec) Sanitize() (err error) {

	if !strings.HasPrefix(s.URL, "https://") && !strings.HasPrefix(s.URL, "http://") {
		s.URL = "https://" + s.URL
	}

	return nil
}
