package engine

import (
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
)

// Options defines application specific behaviors
type Options struct {
	Config    config.Option
	Pipeline  pipeline.Options
	Manifests []Manifest
}

type Manifest struct {
	// Manifests is a list of Updatecli manifest file
	Manifests []string
	// Values is a list of Updatecli value file
	Values []string
	// Secrets is a list of Updatecli secret file
	Secrets []string
}

func (m Manifest) IsZero() bool {
	if len(m.Manifests) == 0 &&
		len(m.Values) == 0 &&
		len(m.Secrets) == 0 {
		return true
	}
	return false
}
