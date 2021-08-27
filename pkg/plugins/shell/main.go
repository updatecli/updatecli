package shell

type ShellSpec struct {
	Command string
}

type Shell struct {
	executor commandExecutor
	spec     ShellSpec
}

func New(spec ShellSpec) (*Shell, error) {
	if spec.Command == "" {
		return nil, &ErrEmptyCommand{}
	}
	return &Shell{
		executor: &nativeCommandExecutor{},
		spec:     spec,
	}, nil
}

// Append the source as last argument if not empty. 
func (s *Shell) appendSource(source string) string {
	// Append the source as last argument if not empty
	if source != "" {
		return s.spec.Command + " " + source
	}

	return s.spec.Command
}
