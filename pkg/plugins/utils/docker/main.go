package docker

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
)

// InlineKeyChain defines the credentials used to authenticate with an OCI registry.
type InlineKeyChain struct {
	// "username" defines the container registry username used for authentication.
	//
	// default:
	//   credentials are retrieved from the local environment, such as `~/.docker/config.json`.
	//
	// remark:
	//   * "username" requires "password".
	//   * "token" cannot be combined with both "username" and "password".
	//
	Username string `yaml:",omitempty"`
	// "password" defines the container registry password used for authentication.
	//
	// default:
	//   credentials are retrieved from the local environment, such as `~/.docker/config.json`.
	//
	// remark:
	//   * "password" requires "username".
	//   * "token" cannot be combined with both "username" and "password".
	//
	Password string `yaml:",omitempty"`
	// "token" defines the container registry bearer token used for authentication.
	//
	// default:
	//   credentials are retrieved from the local environment, such as `~/.docker/config.json`.
	//
	// remark:
	//   * "token" cannot be combined with both "username" and "password".
	//
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

// Empty returns true if the keychain is empty
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
