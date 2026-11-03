package maven

// repository contains repository information retrieved from a pom.xml
type repository struct {
	URL string
	ID  string
}

// dependency contains dependency information retrieved from a pom.xml
type dependency struct {
	GroupID    string
	ArtifactID string
	Version    string
	// Type and Scope are only read for managed dependencies, to tell BOM imports apart
	Type  string
	Scope string
}

// isBOM returns true for a managed dependency importing a bill of materials.
func (d dependency) isBOM() bool {
	return d.Type == "pom" && d.Scope == "import"
}

// parentPom contains parent yinformation retrieved from a pom.xml
type parentPom struct {
	GroupID      string
	ArtifactID   string
	Version      string
	RelativePath string
	Packaging    string
}

var (
	// pomFileName contains the list of accepted pom filename
	pomFileName []string = []string{"pom.xml"}
)
