package github

// ReleaseType specifies accepted GitHub Release type
type ReleaseType struct {
	// Draft enable/disable GitHub draft release
	Draft bool
	// PreRelease enable/disable GitHub PreRelease
	PreRelease bool
	// Release enable/disable GitHub release
	Release bool
	// Latest enable/disable the latest Github Release
	Latest bool
}

func (r *ReleaseType) Init() {
	// If all release type are disable then fallback to stable one only
	if !r.Draft && !r.PreRelease && !r.Release {
		r.Release = true
	}
}
