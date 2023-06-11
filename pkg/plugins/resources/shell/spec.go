package shell

// Spec defines a specification for a "shell" resource
// parsed from an updatecli manifest file
type Spec struct {
	// command specifies the shell command to execute by Updatecli
	Command string `yaml:",omitempty" jsonschema:"required"`
	// environments allows to pass environment variable(s) to the shell script. By default no environment variable are shared.
	Environments Environments `yaml:",omitempty"`
	// ChangedIf defines how to interpreted shell command success criteria. What a success means, what an error means, and what a warning would mean
	ChangedIf SpecChangedIf `yaml:",omitempty" json:",omitempty"`
	// Shell specifies which shell interpreter to use. Default to powershell(Windows) and "/bin/sh" (Darwin/Linux)
	Shell string `yaml:",omitempty"`
	// workdir specifies the working directory path from where to execute the command. It defaults to the current context path (scm or current shell). Updatecli join the current path and the one specified in parameter if the parameter one contains a relative path.
	WorkDir string `yaml:",omitempty"`
}

func (s Spec) Atomic() Spec {
	return Spec{
		Command: s.Command,
	}
}
