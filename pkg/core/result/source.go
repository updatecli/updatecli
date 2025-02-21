package result

import (
	"bytes"
)

// Source holds source execution result
type Source struct {
	// Name holds the source name
	Name string
	/*
		Result holds the source result, accepted values must be one:
			* "SUCCESS"
			* "FAILURE"
			* "ATTENTION"
			* "SKIPPED"
	*/
	Result string
	// Information stores the information detected by the source execution such as a version
	Information string
	// Description stores the source execution description
	Description string
	// Scm stores scm information
	Scm SCM
	// ID contains a uniq identifier for the source
	ID string
	// ConsoleOutput stores the console output of the source execution
	ConsoleOutput string
}

// SetConsoleOutput sets the console output of the source execution
func (s *Source) SetConsoleOutput(out *bytes.Buffer) {
	s.ConsoleOutput = out.String()
}
