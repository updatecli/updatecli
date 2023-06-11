package file

// Spec defines a specification for a "file" resource
// parsed from an updatecli manifest file
type Spec struct {
	// File contains the file path(s) to take in account and is incompatible with Files
	File string `yaml:",omitempty"`
	// Files contains the file path(s) to take in account and is incompatible with File
	Files []string `yaml:",omitempty"`
	// Line contains the line of the file(s) to take in account
	Line int `yaml:",omitempty"`
	// Content specifies the content to take in account instead of the file content
	Content string `yaml:",omitempty"`
	// ForceCreate specifies if nonexistent file(s) should be created if they are targets
	ForceCreate bool `yaml:",omitempty"`
	// MatchPattern specifies the regexp pattern to match on the file(s)
	MatchPattern string `yaml:",omitempty"`
	// ReplacePattern specifies the regexp replace pattern to apply on the file(s) content
	ReplacePattern string `yaml:",omitempty"`
}

func (s Spec) Atomic() Spec {
	return Spec{
		Files:   s.Files,
		Line:    s.Line,
		Content: s.Content,
		File:    s.File,
	}
}
