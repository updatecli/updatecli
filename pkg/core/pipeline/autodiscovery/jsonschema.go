package autodiscovery

import (
	jschema "github.com/invopop/jsonschema"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"
)

// JSONSchema implements the json schema interface to generate the "condition" jsonschema.
func (CrawlersConfig) JSONSchema() *jschema.Schema {
	type CrawlersConfigAlias CrawlersConfig

	return jsonschema.AppendMapToJsonSchema(
		CrawlersConfigAlias{},
		GetAutodiscoverySpecsMapping())
}
