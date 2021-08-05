package engine

import "github.com/updatecli/updatecli/pkg/core/engine/target"

// Options defines application specific behaviors
type Options struct {
	File         string
	ValuesFiles  []string
	SecretsFiles []string
	Target       target.Options
}
