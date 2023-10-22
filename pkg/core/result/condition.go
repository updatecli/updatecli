package result

import "bytes"

// Condition holds condition execution result
type Condition struct {
	//Name holds the condition name
	Name string
	/*
		Result holds the condition result, accepted values must be one:
			* "SUCCESS"
			* "FAILURE"
			* "ATTENTION"
			* "SKIPPED"
	*/
	Result string
	// Pass stores the information detected by the condition execution.
	Pass bool
	// Description stores the condition execution description.
	Description string
	// Scm stores scm information
	Scm SCM
	// ID contains a uniq identifier for the condition
	ID string
	// ConsoleOutput stores the console output of the condition execution
	ConsoleOutput string
}

// SetConsoleOutput sets the console output of the condition execution
func (c *Condition) SetConsoleOutput(out *bytes.Buffer) {
	c.ConsoleOutput = out.String()
}
