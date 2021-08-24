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
