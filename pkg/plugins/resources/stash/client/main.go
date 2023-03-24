package client

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/driver/stash"
	"github.com/drone/go-scm/scm/transport/oauth2"
	"github.com/sirupsen/logrus"
)

// Spec defines a specification for a "bitbucket" resource
// parsed from an updatecli manifest file
type Spec struct {
	// [S][C][T] URL specifies the default github url in case of Bitbucket enterprise
	URL string `yaml:",omitempty" jsonschema:"required"`
	// [S][C][T] Username specifies the username used to authenticate with Bitbucket API
	Username string `yaml:",omitempty"`
	// [S][C][T] Token specifies the credential used to authenticate with
	Token string `yaml:",omitempty"`
	// [S][C][T] Token specifies the credential used to authenticate with
	Password string `yaml:",omitempty"`
	// Owner specifies repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// Repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
}

type Client *scm.Client

func New(s Spec) (Client, error) {
	client, err := stash.New(s.URL)

	if err != nil {
		return nil, err
	}

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
func (s Spec) Validate() error {

	if len(s.URL) == 0 {
		logrus.Errorf("missing %q parameter", "url")
		return fmt.Errorf("wrong configuration")
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
