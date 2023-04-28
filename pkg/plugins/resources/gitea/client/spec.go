package client

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// Spec defines a specification for a "gitea" resource
// parsed from an updatecli manifest file
type Spec struct {
	/*
		"url" defines the Gitea url to interact with
	*/
	URL string `yaml:",omitempty" jsonschema:"required"`
	/*
		"username" defines the username used to authenticate with Gitea API
	*/
	Username string `yaml:",omitempty"`
	/*
		Token specifies the credential used to authenticate with Gitea API
	*/
	Token string `yaml:",omitempty"`
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
