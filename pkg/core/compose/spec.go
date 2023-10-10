package compose

type Spec struct {
	// Policies contains a list of policies
	Policies []Policy
	// Environment contains a list of environment variables
	Environments Environments `yaml:",omitempty"`
	// Env_files contains a list of environment files
	Env_files EnvFiles `yaml:"env_files,omitempty"`
}

type Policy struct {
	// Config contains a list of config ids to be used for the policy
	Configs []string `yaml:",omitempty"`
	// Policy contains the policy OCI name
	Policy string `yaml:",omitempty"`
	// Config contains a list of Updatecli config file path
	Config []string `yaml:",omitempty"`
	// Values contains a list of Updatecli config file path
	Values []string `yaml:",omitempty"`
	// Secrets contains a list of Updatecli secret file path
	Secrets []string `yaml:",omitempty"`
}
