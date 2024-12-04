package result

import (
	"bytes"
)

// SourceInformation defines the information detected by the source execution such as a version
type SourceInformation struct {
	// Key is an identifier for the elem
	Key string
	// Value contains the value retrieved from a source
	Value string
}

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
	Information []SourceInformation
	// Description stores the source execution description
	Description string
	// Scm stores scm information
	Scm SCM
	// ID contains a uniq identifier for the source
	ID string
	//Changelog holds the changelog description
	Changelog string
	// ConsoleOutput stores the console output of the source execution
	ConsoleOutput string
}

// SetConsoleOutput sets the console output of the source execution
func (s *Source) SetConsoleOutput(out *bytes.Buffer) {
	s.ConsoleOutput = out.String()
}
