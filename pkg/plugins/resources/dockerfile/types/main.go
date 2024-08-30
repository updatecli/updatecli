package types

// DockerfileParser is an interface that any updatecli's Dockerfile parser must verifies to be used
type DockerfileParser interface {
	FindInstruction(dockerfileContent []byte) bool
	GetInstruction(dockerfileContent []byte) []StageInstructionValue
	ReplaceInstructions(dockerfileContent []byte, sourceValue string) ([]byte, ChangedLines, error)
}

// ChangedLine is struct to store a single (Dockerfile-)line change's information
type ChangedLines map[int]LineDiff

type LineDiff struct {
	Original string
	New      string
}

type Instruction interface{}

// StagesInstruction is struct to store instruction value for each Dockerfile stages
type StageInstructionValue struct {
	StageName string
	Value     string
}
