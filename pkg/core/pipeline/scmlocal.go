package pipeline

import (
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// buildScm builds the scm handler of the pipeline with newScm, then points it at the local
// checkout when the pipeline runs against one.
func (p *Pipeline) buildScm(scmConfig scm.Config, baseDir string) (scm.Scm, error) {
	s, err := newScm(scmConfig, baseDir, p.Config.Spec.PipelineID)
	if err != nil {
		return scm.Scm{}, err
	}

	// Disabled and search scms have no handler to point at the local checkout. The engine
	// expands search scms into one scm per repository before any pipeline is built.
	if p.Options.IsLocalScm() && s.Handler != nil {
		s.Handler = scm.NewLocal(s.Handler, p.Options.LocalScmDirectory)
	}

	return s, nil
}
