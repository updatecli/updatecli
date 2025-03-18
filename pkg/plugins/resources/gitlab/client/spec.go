package client

// Spec defines a specification for a "GitLab" resource
// parsed from an updatecli manifest file
type Spec struct {
	//  "url" defines the GitLab url to interact with
	//
	//  default:
	//     "gitlab.com"
	URL string `yaml:",omitempty"`
	//  "username" defines the username used to authenticate with GitLab
	Username string `yaml:",omitempty"`
	//  "token" defines the credential used to authenticate with GitLab
	//
	//  remark:
	//    A token is a sensitive information, it's recommended to not set this value directly in the configuration file
	//    but to use an environment variable or a SOPS file.
	//
	//    The value can be set to `{{ requiredEnv "GITLAB_TOKEN"}}` to retrieve the token from the environment variable `GITLAB_TOKEN`
	//	  or `{{ .gitlab.token }}` to retrieve the token from a SOPS file.
	//
	//	  For more information, about a SOPS file, please refer to the following documentation:
	//    https://github.com/getsops/sops
	Token string `yaml:",omitempty"`

	// "tokentype" defines type of provided token. Valid values are "private" and "bearer"
	//
	//  default:
	// 		"private"
	TokenType string `yaml:",omitempty"`
}
