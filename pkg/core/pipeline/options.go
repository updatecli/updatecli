package pipeline

import (
	"github.com/updatecli/updatecli/pkg/core/pipeline/autodiscovery"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
)

// Options hold target parameters

type Options struct {
	Target        target.Options
	AutoDiscovery autodiscovery.Options
}
