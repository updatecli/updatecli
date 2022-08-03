package config

// Config defines autodiscover parameters
type Config struct {
	// Crawlers specifies crawler configuration
	Crawlers map[string]interface{} `yaml:",omitempty"`
	// ScmID specifies a scmid configuration to use to generate the manifest
	ScmId string `yaml:",omitempty"`
	// PullrequestID specifies a pullrequest configuration to use to generate the manifest
	PullrequestId string `yaml:",omitempty"`
}
