package result

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
	// Pass stores the information detected by the condition execution.
	Pass bool
	// Description stores the condition execution description.
	Description string
}
