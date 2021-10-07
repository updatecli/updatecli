package shell

type ShellSpec struct {
	Command        string
	SuppressSource bool
}

type Shell struct {
	executor commandExecutor
	spec     ShellSpec
}

// New returns a reference to a newly initialized Shell object from a shellspec
// or an error if the provided shellspec triggers a validation error.
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
	if source != "" && !s.spec.SuppressSource {
		return s.spec.Command + " " + source
	}

	return s.spec.Command
}
