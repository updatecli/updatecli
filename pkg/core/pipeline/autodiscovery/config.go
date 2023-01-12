package autodiscovery

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
