package engine

import "github.com/olblak/updateCli/pkg/core/engine/target"

// Options defines application specific behaviors
type Options struct {
	File       string
	ValuesFile string
	Target     target.Options
}
