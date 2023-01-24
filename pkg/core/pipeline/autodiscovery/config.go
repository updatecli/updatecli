package autodiscovery

// Config defines autodiscovery parameters
type Config struct {
	// Crawlers defines a map of crawler configuration where the key represent the crawler type
	Crawlers CrawlersConfig `yaml:",omitempty"`
	// ScmID specifies a scmid configuration to use to generate the manifest
	ScmId string `yaml:",omitempty"`
	// ActionId specifies an action configuration to use to generate the manifest
	ActionId string `yaml:",omitempty"`
	// GroupBy specifies how to group pipeline. The Accepted is one of "all", "individual"
	GroupBy GroupBy
	// !Deprecated in favor of `actionid`
	PullrequestId string `yaml:",omitempty"`
}

// CrawlersConfig is a custom type used to generated the jsonschema.
type CrawlersConfig map[string]interface{}
