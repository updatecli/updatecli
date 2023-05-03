package client

// Spec defines a specification for a "Gitlab" resource
// parsed from an updatecli manifest file
type Spec struct {
	/*
		"url" defines the Gitlab url to interact with

		default:
			url defaults to "gitlab.com"
	*/
	URL string `yaml:",omitempty" jsonschema:"required"`
	/*
		"username" defines the username used to authenticate with Gitlab
	*/
	Username string `yaml:",omitempty"`
	/*
		"token" defines the credential used to authenticate with Gitlab
	*/
	Token string `yaml:",omitempty"`
}
