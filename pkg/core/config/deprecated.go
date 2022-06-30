package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/pullrequest"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
)

// If an scm of kind "git" is used alone, so without pullrequest referencing it
// Then we disable the ability to use temporary working branch such as updatecli_xxx
// The motivation is to give enough time to manifest to migrate to the new default behavior
// more context is available on https://github.com/updatecli/updatecli/pull/745
func (c *Config) migrateToGitTmpWorkingBranch() error {
	gitScmIDs := []string{}

	for id, s := range c.Spec.SCMs {
		if s.Kind == "git" {
			gitScmIDs = append(gitScmIDs, id)
		}
	}

	if len(gitScmIDs) == 0 {
		return nil
	}

	for _, gitScmID := range gitScmIDs {
		shouldGitTmpWorkingBranchBeDisabled := true

		for _, s := range c.Spec.PullRequests {
			if s.ScmID == gitScmID {
				shouldGitTmpWorkingBranchBeDisabled = false
			}
		}
		if shouldGitTmpWorkingBranchBeDisabled {
			gitSpec := git.Spec{}

			err := mapstructure.Decode(c.Spec.SCMs[gitScmID], &gitSpec)
			if err != nil {
				return err
			}
			gitSpec.DisableWorkingBranch = true
			s := c.Spec.SCMs[gitScmID]
			s.Spec = gitSpec
			c.Spec.SCMs[gitScmID] = s
			break
		}

	}

	return nil

}

func generateScmFromLegacyCondition(id string, config *Config) {

	c := config.Spec.Conditions[id]

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
	config.Spec.Conditions[id] = c
}

func generateScmFromLegacyTarget(id string, config *Config) error {
	t := config.Spec.Targets[id]

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
	config.Spec.Targets[id] = t

	return nil
}
