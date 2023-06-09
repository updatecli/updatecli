package result

import "fmt"

// Target holds target execution result
type Target struct {
	// Name holds the target name
	Name string
	/*
		Result holds the target result, accepted values must be one:
			* "SUCCESS"
			* "FAILURE"
			* "ATTENTION"
			* "SKIPPED"
	*/
	Result string
	// OldInformation stores the old information detected by the target execution
	OldInformation string
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
}

func (t *Target) String() string {
	str := fmt.Sprintf("%q => %q", t.OldInformation, t.NewInformation)
	str = str + fmt.Sprintf("\n%s - %s", t.Result, t.Description)
	return str
}
