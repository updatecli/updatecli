package manifest

type Manifest struct {
	// Manifests is a list of Updatecli manifest file
	Manifests []string
	// Values is a list of Updatecli value file
	Values []string
	// Secrets is a list of Updatecli secret file
	Secrets []string
	// GraphOutput is a path to output the manifest graph to
	GraphOutput string
}

func (m Manifest) IsZero() bool {
	if len(m.Manifests) == 0 &&
		len(m.Values) == 0 &&
		len(m.Secrets) == 0 {
		return true
	}
	return false
}
