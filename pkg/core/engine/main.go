package engine

import (
	"errors"

	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/tmp"
)

var (
	// ErrNoManifestDetected is the error message returned by Updatecli if it can't find manifest
	ErrNoManifestDetected error = errors.New("no Updatecli manifest detected")
)

// Engine defined parameters for a specific engine run.
type Engine struct {
	configurations []*config.Config
	Pipelines      []*pipeline.Pipeline
	Options        Options
	Reports        reports.Reports
}

// Clean remove every traces from an updatecli run.
func (e *Engine) Clean() (err error) {
	err = tmp.Clean()
	return
}
