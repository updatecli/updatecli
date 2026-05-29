package pipeline

import (
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
)

// Options hold target parameters

type Options struct {
	Target target.Options
	// DisableChangelog disables changelog retrieval for targets.
	DisableChangelog bool
}
