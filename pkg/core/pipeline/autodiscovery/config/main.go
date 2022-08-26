package config

import (
	"github.com/updatecli/updatecli/pkg/core/pipeline/pullrequest"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Config defines autodiscover parameters
type Config struct {
	// Crawlers specifies crawler configuration
	Crawlers map[string]interface{} `yaml:",omitempty"`
	// ScmID specifies a scmid configuration to use to generate the manifest
	ScmId string `yaml:",omitempty"`
	// PullrequestID specifies a pullrequest configuration to use to generate the manifest
	PullrequestId string `yaml:",omitempty"`
	// GroupBy specifies how to group pipeline. The Accepted is one of "all", "individual"
	GroupBy GroupBy
}

type Input struct {
	// ScmSpec defines the scm specification
	ScmSpec *scm.Config
	// ScmID defines the scmid associated to the scm specification
	ScmID string
	// PullRequestSpecSpec defines the pullrequest specificiation
	PullRequestSpec *pullrequest.Config
	// ScmID defines the scmid associated to the scm specification
	PullrequestID string
}
