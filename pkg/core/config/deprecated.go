package config

import (
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

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
