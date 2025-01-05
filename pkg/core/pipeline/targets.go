package pipeline

import (
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

var (
	// ErrRunTargets is return when at least one error happened during targets execution
	ErrRunTargets error = errors.New("something went wrong during target execution")
)

func (p *Pipeline) updateTarget(id, result string) {

	target := p.Targets[id]
	target.Result.Result = result
	p.Targets[id] = target
	p.Report.Targets[id] = &target.Result
}

// RunTarget run a target by id
func (p *Pipeline) RunTarget(id string) (r string, changed bool, err error) {
	target := p.Targets[id]
	target.Config = p.Config.Spec.Targets[id]
	// Ensure the result named contains the up to date target name after templating
	target.Result.Name = target.Config.ResourceConfig.Name
	target.Result.DryRun = target.DryRun
	err = target.Run(p.Sources[target.Config.SourceID].Output, &p.Options.Target)

	if err != nil {
		p.Report.Result = result.FAILURE
		target.Result.Result = result.FAILURE
		err = fmt.Errorf("something went wrong in target %q : %q", id, err)
	}

	p.Targets[id] = target
	p.Report.Targets[id] = &target.Result

	return target.Result.Result, target.Result.Changed, err
}
