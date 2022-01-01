package types

// DockerfileParser is an interface that any updatecli's Dockerfile parser must verifies to be used
type DockerfileParser interface {
	FindInstruction(dockerfileContent []byte) bool
	ReplaceInstructions(dockerfileContent []byte, sourceValue string) ([]byte, ChangedLines, error)
}

// ChangedLine is struct to store a single (Dockerfile-)line change's informations
type ChangedLines map[int]LineDiff

type LineDiff struct {
	Original string
	New      string
}

type Instruction interface{}
