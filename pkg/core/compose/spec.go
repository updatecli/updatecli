package compose

type Spec struct {
	// Name contains the compose name
	Name string `yaml:",omitempty" jsonschema:"required"`
	// Policies contains a list of policies
	Policies []Policy
	// Environment contains a list of environment variables
	//
	// Example:
	//   ENV_VAR1: value1
	//   ENV_VAR2: value2
	Environments Environments `yaml:",omitempty"`
	// Env_files contains a list of environment files
	//
	// Example:
	//   - env_file1.env
	//   - env_file2.env
	Env_files EnvFiles `yaml:"env_files,omitempty"`
	// ValuesInline contains a list of inline values for Updatecli config
	// This is the default inline values that will be applied to all policies if not overridden by the policy inline values.
	// A deep merge will be performed between global inline values and policy inline values if both are defined.
	//
	// Example:
	//   key1: value1
	//   key2: value2
	ValuesInline *map[string]any `yaml:",omitempty"`
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
	// ValuesInline contains a list of inline values for Updatecli config
	// A deep merge will be performed between global inline values and policy inline values if both are defined.
	//
	// Example:
	//   key1: value1
	//   key2: value2
	ValuesInline *map[string]any `yaml:",omitempty"`
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

	if p.ValuesInline != nil {
		return false
	}

	if p.Policy != "" {
		return false
	}

	return true
}
