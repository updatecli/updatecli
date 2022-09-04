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
