package source

// Changelog is an interface to retrieve changelog description
type Changelog interface {
	Changelog(release string) (string, error)
}
