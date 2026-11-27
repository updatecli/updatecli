package engine

import (
	"github.com/updatecli/updatecli/pkg/core/log"
)

// enforceLocalScm turns off what a run against a local checkout must not do, before any
// pipeline is built: committing, pushing and publishing reports to Udash. A report would
// describe the changes of the local checkout as changes of the remote repository, and
// carry local paths.
func (e *Engine) enforceLocalScm() {
	if !e.Options.Pipeline.IsLocalScm() {
		return
	}

	e.Options.Pipeline.EnforceLocalScm()

	if !e.Options.DisableUdashReport {
		log.Debugf("not publishing reports to Udash: the pipelines run against the local checkout %q",
			e.Options.Pipeline.LocalScmDirectory)
		e.Options.DisableUdashReport = true
	}
}
