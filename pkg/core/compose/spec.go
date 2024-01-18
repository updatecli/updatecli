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
	// Name contains the policy name
	Name string `yaml:",omitempty"`
	// Policy contains the policy OCI name
	Policy string `yaml:",omitempty"`
	// Config contains a list of Updatecli config file path
	Config []string `yaml:",omitempty"`
	// Values contains a list of Updatecli config file path
	Values []string `yaml:",omitempty"`
	// Secrets contains a list of Updatecli secret file path
	Secrets []string `yaml:",omitempty"`
}

func (p Policy) IsZero() bool {

	if len(p.Config) > 0 {
		return false
	}

	if len(p.Values) > 0 {
		return false
	}

	if len(p.Secrets) > 0 {
		return false
	}

	if p.Policy != "" {
		return false
	}

	return true
}
