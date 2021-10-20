package engine

import "github.com/updatecli/updatecli/pkg/core/pipeline"

// Options defines application specific behaviors
type Options struct {
	File         string
	ValuesFiles  []string
	SecretsFiles []string
	Pipeline     pipeline.Options
}
