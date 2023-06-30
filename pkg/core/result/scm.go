package result

// SCM defines scm information
type SCM struct {
	// URL defines the git URL
	URL string
	// Branch defines the different branches used by Updatecli
	Branch GitBranch
	// ID contains a uniq identifier for the condition
	ID string
}

type GitBranch struct {
	// Source defines the branch used as a source
	Source string
	// Working defines the working branch used by Updatecli
	Working string
	// Target defines the branch used by Updatecli to apply the changes done from the working branch
	Target string
}
