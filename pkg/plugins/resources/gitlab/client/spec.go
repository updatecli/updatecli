package client

// Spec defines a specification for a "GitLab" resource
// parsed from an updatecli manifest file
type Spec struct {
	/*
		"url" defines the GitLab url to interact with

		default:
			url defaults to "gitlab.com"
	*/
	URL string `yaml:",omitempty" jsonschema:"required"`
	/*
		"username" defines the username used to authenticate with GitLab
	*/
	Username string `yaml:",omitempty"`
	/*
		"token" defines the credential used to authenticate with GitLab
	*/
	Token string `yaml:",omitempty"`
}
