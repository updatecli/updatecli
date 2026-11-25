package pipeline

// Gate orders the pipelines that autodiscovery generates from "tracks".
//
// Pipelines sharing a Key update the same dependency, or the same group of dependencies.
// A pipeline waits while one with the same Key and a lower Rank still has a pending update.
type Gate struct {
	// Key identifies the dependency, or group of dependencies, the pipeline updates.
	Key string
	// Rank is the position of the pipeline track in "tracks".
	Rank int
	// Track is the name of the pipeline track.
	Track string
	// NoWait is set when the pipeline never waits for earlier tracks.
	// Later tracks still wait for it.
	NoWait bool
	// Waiting is set when the pipeline was skipped during this run, waiting for an earlier track.
	// A waiting pipeline never ran, so its actions are neither run nor cleaned up.
	Waiting bool
}

// IsWaiting reports whether the pipeline was skipped during this run, waiting for an earlier track.
func (p *Pipeline) IsWaiting() bool {
	return p.Gate != nil && p.Gate.Waiting
}
