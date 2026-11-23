package autodiscovery

import (
	"strings"

	jschema "github.com/invopop/jsonschema"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"
)

// JSONSchema implements the json schema interface to generate the "condition" jsonschema.
func (CrawlersConfig) JSONSchema() *jschema.Schema {
	type CrawlersConfigAlias CrawlersConfig

	schema := jsonschema.AppendMapToJsonSchema(
		CrawlersConfigAlias{},
		GetAutodiscoverySpecsMapping())

	appendVariantSpecToCrawlerSchemas(schema)

	return schema
}

// appendVariantSpecToCrawlerSchemas adds the settings shared by every crawler, such as "versionfilters",
// to each crawler schema, except the WASM plugin one.
func appendVariantSpecToCrawlerSchemas(schema *jschema.Schema) {
	if schema == nil || schema.Properties == nil {
		return
	}

	variantSchema := jsonschema.AppendMapToJsonSchema(VariantSpec{}, nil)
	if variantSchema == nil || variantSchema.Properties == nil {
		return
	}

	for pair := schema.Properties.Oldest(); pair != nil; pair = pair.Next() {
		if pair.Key == pluginCrawlerKind || strings.HasSuffix(pair.Key, ".wasm") {
			continue
		}

		crawlerSchema := pair.Value
		if crawlerSchema == nil || crawlerSchema.Properties == nil {
			continue
		}

		for variantPair := variantSchema.Properties.Oldest(); variantPair != nil; variantPair = variantPair.Next() {
			crawlerSchema.Properties.Set(variantPair.Key, variantPair.Value)
		}
	}
}
