package shell

// ShellSpec defines a specification for a "shell" resource
// parsed from an updatecli manifest file
type ShellSpec struct {
	Command string
}

// Shell defines a resource of type "shell"
type Shell struct {
	executor commandExecutor
	spec     ShellSpec
}

// New returns a reference to a newly initialized Shell object from a ShellSpec
// or an error if the provided ShellSpec triggers a validation error.
func New(spec ShellSpec) (*Shell, error) {
	if spec.Command == "" {
		return nil, &ErrEmptyCommand{}
	}
	return &Shell{
		executor: &nativeCommandExecutor{},
		spec:     spec,
	}, nil
}

// appendSource appends the source as last argument if not empty.
func (s *Shell) appendSource(source string) string {
	// Append the source as last argument if not empty
	if source != "" {
		return s.spec.Command + " " + source
	}

	return s.spec.Command
}
