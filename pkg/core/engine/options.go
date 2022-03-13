package engine

import (
	"github.com/updatecli/updatecli/pkg/core/options"
)

// Options defines application specific behaviors
type Options struct {
	Config   options.Config
	Pipeline options.Pipeline
}
