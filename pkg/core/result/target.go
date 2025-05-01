package result

import (
	"bytes"
	"fmt"
)

// Target holds target execution result
type Target struct {
	// Name holds the target name
	Name string
	// DryRun defines if a target was executed in DryRun mode
	DryRun bool
	/*
		Result holds the target result, accepted values must be one:
			* "SUCCESS"
			* "FAILURE"
			* "ATTENTION"
			* "SKIPPED"
	*/
	Result string
	// Information stores the old information detected by the target execution
	Information string
	// NewInformation stores the new information updated by during the target execution
	NewInformation string
	// Description stores the target execution description
	Description string
	// Files holds the list of files modified by a target execution
	Files []string
	// Changed specifies if the target was modify during the pipeline execution
	Changed bool
	// Scm stores scm information
	Scm SCM
	// ID contains a uniq identifier for the target
	ID string
	// ConsoleOutput stores the console output of the target execution
	ConsoleOutput string
	//Changelogs holds the changelog description
	Changelogs []Changelog
	// Config stores the source configuration
	Config any
	// SourceID stores the source ID used by the target
	SourceID string
}

func (t *Target) String() string {
	str := fmt.Sprintf("%q => %q", t.Information, t.NewInformation)
	str = str + fmt.Sprintf("\n%s - %s", t.Result, t.Description)
	return str
}

// SetConsoleOutput sets the console output of the target execution
func (t *Target) SetConsoleOutput(out *bytes.Buffer) {
	t.ConsoleOutput = out.String()
}
