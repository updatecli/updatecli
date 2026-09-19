package result

import "bytes"

// Conditions holds condition execution result
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
	// IsRun reports whether the pipeline has finalized this condition, meaning it
	// either executed or was explicitly skipped through a dependsOn condition.
	// It distinguishes a condition still pending execution (Result defaults to
	// SKIPPED while IsRun is false) from one that actually ran or was
	// deliberately skipped (IsRun is true).
	IsRun bool
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
	// Config stores the source configuration
	Config any
	// SourceID stores the source ID used by the condition
	SourceID string
}

// SetConsoleOutput sets the console output of the condition execution
func (c *Condition) SetConsoleOutput(out *bytes.Buffer) {
	c.ConsoleOutput = out.String()
}
