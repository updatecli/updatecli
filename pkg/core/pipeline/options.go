package pipeline

import (
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
)

// Options hold target parameters

type Options struct {
	Target target.Options
	// DisableChangelog disables changelog retrieval for targets.
	DisableChangelog bool
	// LocalScmDirectory, when set, makes every scm of the pipeline use this existing
	// checkout instead of cloning. See EnforceLocalScm for what that turns off.
	LocalScmDirectory string
}

// IsLocalScm reports whether the scms run against an existing local checkout.
func (o Options) IsLocalScm() bool {
	return o.LocalScmDirectory != ""
}

// EnforceLocalScm turns off commit and push when the scms run against a local checkout,
// which is never committed to nor pushed. Actions are left out too, by initActions and
// updateActions. It is safe to call more than once.
func (o *Options) EnforceLocalScm() {
	if !o.IsLocalScm() {
		return
	}

	o.Target.Commit = false
	o.Target.Push = false
}
