package target

// Options hold target parameters
type Options struct {
	// Commit indicates whether to git commit changes
	Commit bool
	// Push indicates whether to git push changes
	Push bool
	// Clean indicates whether to perform cleaning operations. For example removing temporary files.
	Clean bool
	// DryRun indicates whether to perform a dry-run (no changes applied).
	DryRun bool
	// CleanGitBranches indicates whether to delete git branches if no changes are detected.
	CleanGitBranches bool
	// ExistingOnly indicates whether to skip publishing when a new pipeline is detected
	ExistingOnly bool
}
