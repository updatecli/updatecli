package engine

import "github.com/olblak/updateCli/pkg/core/engine/target"

// Options defines application specific behaviors
type Options struct {
	File         string
	ValuesFiles  []string
	SecretsFiles []string
	Target       target.Options
}
