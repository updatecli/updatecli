package docker

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
)

type InlineKeyChain struct {
	// [S][C][T] Username specifies the container registry username to use for authentication. Not compatible with token
	Username string `yaml:",omitempty"`
	// [S][C][T] Password specifies the container registry password to use for authentication. Not compatible with token
	Password string `yaml:",omitempty"`
	// [S][C][T] Token specifies the container registry token to use for authentication. Not compatible with username/password
	Token string `yaml:",omitempty"`
}

// Resolve the inline keychain and return an authenticator
func (kc InlineKeyChain) Resolve(authn.Resource) (authn.Authenticator, error) {
	return authn.FromConfig(authn.AuthConfig{
		Username:      kc.Username,
		Password:      kc.Password,
		RegistryToken: kc.Token,
	}), nil
}

func (kc InlineKeyChain) Empty() bool {
	return kc.Username == "" && kc.Password == "" && kc.Token == ""
}

// Validate validates the object and returns an error (with all the failed validation messages) if it is not valid
func (kc InlineKeyChain) Validate() error {
	var validationErrors []string

	if len(kc.Token) > 0 {
		if len(kc.Username) > 0 && len(kc.Password) > 0 {
			validationErrors = append(validationErrors, "Specifying a (bearer) token is invalid when a username and a password are provided.")
		}
	}

	if len(kc.Username) > 0 && len(kc.Password) == 0 {
		validationErrors = append(validationErrors, "Docker registry username provided but not the password")
	} else if len(kc.Username) == 0 && len(kc.Password) > 0 {
		validationErrors = append(validationErrors, "Docker registry password provided but not the username")
	}

	// Return all the validation errors if found any
	if len(validationErrors) > 0 {
		return fmt.Errorf("validation error: the provided manifest configuration had the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	return nil
}
