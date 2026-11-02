package pipeline

import (
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
)

// buildScm builds the scm handler of the pipeline with newScm, then points it at the local
// checkout when the pipeline runs against one and the scm describes the same repository.
// Any other scm is cloned as usual, and its targets run in dry run, see targetOptions.
func (p *Pipeline) buildScm(scmConfig scm.Config, baseDir string) (scm.Scm, error) {
	s, err := newScm(scmConfig, baseDir, p.Config.Spec.PipelineID)
	if err != nil {
		return scm.Scm{}, err
	}

	// Disabled and search scms have no handler. The engine expands search scms into one scm
	// per repository before any pipeline is built, and each is then matched like any other.
	if p.Options.IsLocalScm() && s.Handler != nil && scm.UsesLocalCheckout(s.Handler, p.Options.LocalScmDirectory) {
		s.Handler = scm.NewLocal(s.Handler, p.Options.LocalScmDirectory)
	}

	return s, nil
}

// localTargetKinds lists the target kinds only changing files: in dev mode, they are the only
// ones applied, and only against the local checkout or without any scm. Any other kind, such
// as gitbranch, gittag or a release, may change a remote, so it runs in dry run. A new kind
// runs in dry run until it is added here.
var localTargetKinds = map[string]bool{
	"bazelmod":           true,
	"cargopackage":       true,
	"csv":                true,
	"dockerfile":         true,
	"file":               true,
	"golang":             true,
	"golang/gomod":       true,
	"hcl":                true,
	"helmchart":          true,
	"json":               true,
	"shell":              true,
	"terraform/file":     true,
	"terraform/lock":     true,
	"terraform/provider": true,
	"toml":               true,
	"toolversions":       true,
	"xml":                true,
	"yaml":               true,
}

// targetOptions returns the options the target runs with.
//
// In dev mode, commit and push are already off, but some targets, such as gitbranch, gittag
// or releases, change a remote by themselves. Only a target changing files, against the local
// checkout or without any scm, is therefore applied: any other runs in dry run, so no remote
// and no other repository is ever changed.
func (p *Pipeline) targetOptions(id string, t target.Target) target.Options {
	options := p.Options.Target

	if !p.Options.IsLocalScm() || options.DryRun {
		return options
	}

	switch {
	case t.Scm != nil && !scm.IsLocal(*t.Scm):
		log.Infof("Target %q runs in dry run: its scm is not the local checkout %q", id, p.Options.LocalScmDirectory)
	case !localTargetKinds[t.Config.Kind]:
		log.Infof("Target %q runs in dry run: in dev mode, a target of kind %q may change a remote", id, t.Config.Kind)
	default:
		return options
	}

	options.DryRun = true

	return options
}
