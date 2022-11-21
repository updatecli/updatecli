package config

import (
	"github.com/updatecli/updatecli/pkg/core/pipeline/action"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Config defines autodiscovery parameters
type Config struct {
	// Crawlers specifies crawler configuration
	Crawlers map[string]interface{} `yaml:",omitempty"`
	// ScmID specifies a scmid configuration to use to generate the manifest
	ScmId string `yaml:",omitempty"`
	// ActionId specifies an action configuration to use to generate the manifest
	ActionId string `yaml:",omitempty"`
	// GroupBy specifies how to group pipeline. The Accepted is one of "all", "individual"
	GroupBy GroupBy
	// !Deprecated in favor of `actionid`
	PullrequestId string `yaml:",omitempty"`
}

type Input struct {
	// ScmSpec defines the scm specification
	ScmSpec *scm.Config
	// ScmID defines the scmid associated to the scm specification
	ScmID string
	// ActionConfig defines the action specification
	ActionConfig *action.Config
	// ActionID defines the scmid associated to the scm specification
	ActionID string
}
