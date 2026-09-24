package client

// Spec defines the settings used to connect to a GitLab instance.
type Spec struct {
	// "url" defines the GitLab url to interact with.
	//
	// default:
	//   gitlab.com
	//
	// remark:
	//   * "https://" is added when the url has no "http://" or "https://" scheme.
	//
	// example:
	//   * url: gitlab.com
	//   * url: https://gitlab.example.com
	//
	URL string `yaml:",omitempty"`
	// "username" defines the username used to authenticate with GitLab.
	//
	Username string `yaml:",omitempty"`
	// "token" defines the credential used to authenticate with GitLab.
	//
	// remark:
	//   * a token is sensitive information. Do not set it directly in the manifest,
	//     use an environment variable or a SOPS file instead.
	//   * `{{ requiredEnv "GITLAB_TOKEN" }}` retrieves the token from the environment variable `GITLAB_TOKEN`.
	//   * `{{ .gitlab.token }}` retrieves the token from a SOPS file.
	//   * for more information about SOPS files, see https://github.com/getsops/sops
	//
	// example:
	//   * token: '{{ requiredEnv "GITLAB_TOKEN" }}'
	//
	Token string `yaml:",omitempty"`
}
