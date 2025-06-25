package autodiscovery

// Config defines autodiscovery parameters
type Config struct {
	/*
		Crawlers defines a map of crawler configuration where the key represent the crawler type
	*/
	Crawlers CrawlersConfig `yaml:",omitempty"`
	/*
		scmid is a unique identifier used to retrieve the scm configuration from the configuration file.
	*/
	ScmId string `yaml:",omitempty"`
	/*
		actionid is a unique identifier used to retrieve the action configuration from the configuration file.
	*/
	ActionId string `yaml:",omitempty"`
	/*
		groupby specifies how to group pipeline. The Accepted is one of "all", "individual". Default is "all"

		default:
			all
	*/
	GroupBy GroupBy
	// !Deprecated in favor of `actionid`
	PullrequestId string `yaml:",omitempty"`
}

// CrawlersConfig is a custom type used to generated the jsonschema.
type CrawlersConfig map[string]any
