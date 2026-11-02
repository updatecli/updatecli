package pipeline

import (
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
)

// Options hold target parameters

type Options struct {
	Target target.Options
	// DisableChangelog disables changelog retrieval for targets.
	DisableChangelog bool
	// LocalScmDirectory, when set, makes every scm describing the repository of this existing
	// checkout use it instead of cloning. Other scms are cloned, and their targets run in dry
	// run. See EnforceLocalScm for what else that turns off.
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
