package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/condition"
	"github.com/updatecli/updatecli/pkg/core/pipeline/pullrequest"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
)

func generateScmFromLegacyCondition(id string, c *condition.Config, config *Config) {
	logrus.Warningf("The directive 'scm' for the condition[%q] is now deprecated. Please use the new top level scms syntax", id)
	if len(c.SCMID) == 0 {
		if _, ok := config.Spec.SCMs["condition_"+id]; !ok {
			for kind, spec := range c.Scm {

				if config.Spec.SCMs == nil {
					config.Spec.SCMs = make(map[string]scm.Config, 1)
				}

				config.Spec.SCMs["condition_"+id] = scm.Config{
					Kind: kind,
					Spec: spec}

			}
		}
		c.SCMID = "condition_" + id
		// Ensure we clean this deprecated variable
	} else {
		logrus.Warning("condition.SCMID is also defined, ignoring condition.Scm")
	}
	c.Scm = map[string]interface{}{}
}

func generateScmFromLegacyTarget(id string, t *target.Config, config *Config) error {
	logrus.Warningf("The directive 'scm' for the target[%q] is now deprecated. Please use the new top level scms syntax", id)
	if len(t.SCMID) == 0 {
		if _, ok := config.Spec.SCMs["target_"+id]; !ok {
			for kind, spec := range t.Scm {
				if config.Spec.SCMs == nil {
					config.Spec.SCMs = make(map[string]scm.Config, 1)
				}

				config.Spec.SCMs["target_"+id] = scm.Config{
					Kind: kind,
					Spec: spec,
				}

				// Temporary code until we fully deprecated the old pullRequest syntax.
				// At this stage we can only have github pullrequest so to not break backward
				// compatibility, we automatically add a pullRequest configuration in case of github scm
				// https://github.com/updatecli/updatecli/pull/388
				if kind == "github" {
					if config.Spec.PullRequests == nil {
						config.Spec.PullRequests = make(map[string]pullrequest.Config, 1)
					}

					githubSpec := github.Spec{}

					err := mapstructure.Decode(spec, &githubSpec)
					if err != nil {
						return err
					}

					config.Spec.PullRequests["target_"+id] = pullrequest.Config{
						Kind:    kind,
						Spec:    githubSpec.PullRequest,
						ScmID:   "target_" + id,
						Targets: []string{id},
					}
				}

			}
		}
		t.SCMID = "target_" + id
	} else {
		logrus.Warning("target.SCMID is also defined, ignoring target.Scm")
	}
	t.Scm = map[string]interface{}{}

	return nil
}
