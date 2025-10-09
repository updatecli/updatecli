package github

// ReleaseType specifies accepted GitHub Release type
type ReleaseType struct {
	// "Draft" enable/disable GitHub draft release
	Draft bool
	// "PreRelease" enable/disable GitHub PreRelease
	PreRelease bool
	// "Release" enable/disable GitHub release
	Release bool
	// "Latest" if set to true will only filter the release flag as latest.
	Latest bool
}

// Init initializes the ReleaseType struct
func (r *ReleaseType) Init() {
	// If all release type are disable then fallback to stable one only
	if !r.Draft && !r.PreRelease && !r.Release {
		r.Release = true
	}
}

// IsZero checks if all release type are set to disable
func (r ReleaseType) IsZero() bool {
	var empty ReleaseType
	return empty == r
}
