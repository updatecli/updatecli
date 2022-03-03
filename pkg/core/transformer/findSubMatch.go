package transformer

// FindSubMatch is struct used to feed regexp.findSubMatch
type FindSubMatch struct {
	Pattern      string `yaml:"pattern"`
	CaptureIndex int    `yaml:"captureIndex"`
}
