package config

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// GetChangelogTitle try to guess a specific target based on various information available for
// a specific job
func (config *Config) GetChangelogTitle(ID string, fallback string) (title string) {
	if len(config.Spec.Title) > 0 {
		// If a pipeline title has been defined, then use it for pull request title
		title = fmt.Sprintf("[updatecli] %s",
			config.Spec.Title)

	} else if len(config.Spec.Targets) == 1 && len(config.Spec.Targets[ID].Name) > 0 {
		// If we only have one target then we can use it as fallback.
		// Reminder, map in golang are not sorted so the order can't be kept between updatecli run
		title = fmt.Sprintf("[updatecli] %s", config.Spec.Targets[ID].Name)
	} else {
		// At the moment, we don't have an easy way to describe what changed
		// I am still thinking to a better solution.
		logrus.Warning("**Fallback** Please add a title to you configuration using the field 'title: <your pipeline>'")
		title = fmt.Sprintf("[updatecli][%s] Bump version to %s",
			config.Spec.Sources[config.Spec.Targets[ID].SourceID].Kind,
			fallback)
	}
	return title
}

// EnsureLocalScm ensures the config receiver has a
// "local" SCM configuration if needed
func (config *Config) EnsureLocalScm() error {
	if !config.needScm(LOCALSCMIDENTIFIER) {
		return nil
	}
	// Is there already a "local" SCM specified by the user
	localScmConfig, ok := config.Spec.SCMs[LOCALSCMIDENTIFIER]
	if ok && localScmConfig.Disabled {
		// No need to "auto-guess" if user disabled the local SCM explicitly
		return nil
	}

	// Autoguess (localScmConfig is a copy of the config tree)
	if err := localScmConfig.AutoGuess(LOCALSCMIDENTIFIER, "", config.gitHandler); err != nil {
		return err
	}

	if config.Spec.SCMs == nil {
		logrus.Infof("\nINFO: A configuration for the 'local' SCM is automatically guessed. To opt-out, please set `scms.local.disabled` to the value `true`\n")
		config.Spec.SCMs = make(map[string]scm.Config)
	}
	config.Spec.SCMs[LOCALSCMIDENTIFIER] = localScmConfig
	return nil
}

// NormalizeTargets ensures that any target specification of the provided Config object is set up with any default if required.
func (config *Config) NormalizeTargets() error {
	for id, targetSpec := range config.Spec.Targets {
		// Automatically set sourceid if there is only 1 source (and source input is enabled)
		if !targetSpec.DisableSourceInput && len(targetSpec.SourceID) == 0 && len(config.Spec.Sources) == 1 {
			// Nifty trick to retrieve the key of the only element of the map
			for id := range config.Spec.Sources {
				targetSpec.SourceID = id
			}
			logrus.Debugf("setting sourceid to %q for target %q", targetSpec.SourceID, id)
		}

		// Assign the updated targetSpec in the config object
		config.Spec.Targets[id] = targetSpec
	}
	return nil
}

// NormalizeConditions ensures that any condition specification of the provided Config object is set up with any default if required.
func (config *Config) NormalizeConditions() error {
	for id, conditionSpec := range config.Spec.Conditions {
		// Automatically set sourceid if there is only 1 source (and source input is enabled)
		if !conditionSpec.DisableSourceInput && len(conditionSpec.SourceID) == 0 && len(config.Spec.Sources) == 1 {
			// Nifty trick to retrieve the key of the only element of the map
			for id := range config.Spec.Sources {
				conditionSpec.SourceID = id
			}
			logrus.Debugf("setting sourceid to %q for condition %q", conditionSpec.SourceID, id)
		}

		// Assign the updated conditionSpec in the config object
		config.Spec.Conditions[id] = conditionSpec
	}
	return nil
}

func (config Config) needScm(requiredScmId string) bool {
	for _, sourceSpec := range config.Spec.Sources {
		if sourceSpec.SCMID == requiredScmId {
			return true
		}
	}
	for _, conditionSpec := range config.Spec.Conditions {
		if conditionSpec.SCMID == requiredScmId {
			return true
		}
	}
	for _, targetSpec := range config.Spec.Targets {
		if targetSpec.SCMID == requiredScmId {
			return true
		}
	}
	for _, actionSpec := range config.Spec.Actions {
		if actionSpec.ScmID == requiredScmId {
			return true
		}
	}
	return false
}
