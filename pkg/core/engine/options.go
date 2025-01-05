package engine

import (
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/engine/manifest"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
)

// Options defines application specific behaviors
type Options struct {
	Config        config.Option
	Pipeline      pipeline.Options
	Manifests     []manifest.Manifest
	DisplayFlavor string
	GraphFlavor   string
}
