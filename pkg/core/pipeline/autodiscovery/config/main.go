package config

// Config defines autodiscover parameters
type Config struct {
	// Crawlers specifies crawler configuration
	Crawlers map[string]interface{} `yaml:",omitempty"`
	// ScmID specifies a scmid configuration to use to generate the manfiest
	ScmId string `yaml:",omitempty"`
}
